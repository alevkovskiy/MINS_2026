package client

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "warehouse_microservices/proto/compatibility"
)

type CompatibilityClient struct {
	conn         *grpc.ClientConn
	client       pb.CompatibilityServiceClient
	addr         string
	healthy      atomic.Bool
	lastFailTime time.Time
	failCount    atomic.Int32
}

func NewCompatibilityClient(addr string) (*CompatibilityClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Service B: %w", err)
	}

	c := &CompatibilityClient{
		conn:   conn,
		client: pb.NewCompatibilityServiceClient(conn),
		addr:   addr,
	}
	c.healthy.Store(true)

	go c.healthChecker()

	return c, nil
}

func (c *CompatibilityClient) Close() error {
	return c.conn.Close()
}

func (c *CompatibilityClient) isAvailable() bool {
	state := c.conn.GetState()
	return c.healthy.Load() && (state == connectivity.Ready || state == connectivity.Idle)
}

func (c *CompatibilityClient) healthChecker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := c.client.HealthCheck(ctx, &emptypb.Empty{})
		cancel()

		if err != nil {
			failCount := c.failCount.Add(1)
			if c.healthy.Load() {
				log.Printf("[HealthCheck] Service B стал НЕДОСТУПЕН: %v (ошибка #%d)", err, failCount)
			}
			c.healthy.Store(false)
			c.lastFailTime = time.Now()
		} else {
			if !c.healthy.Load() {
				log.Printf("[HealthCheck] Service B снова ДОСТУПЕН после %d ошибок", c.failCount.Load())
				c.failCount.Store(0)
			}
			c.healthy.Store(true)
		}
	}
}

func (c *CompatibilityClient) CheckCompatibilityWithFallback(ctx context.Context, traceID, category string, existingCategories []string) (compatible bool, message string, err error) {
	if c == nil {
		log.Printf("[TraceID: %s] Service B недоступен (клиент не инициализирован), используем fallback", traceID)
		return true, "Сервис проверки совместимости не запущен, товар добавлен без проверки", nil
	}

	if !c.isAvailable() {
		log.Printf("[TraceID: %s] Service B недоступен (health check провален), используем fallback", traceID)
		return true, "Сервис проверки совместимости временно недоступен, товар добавлен без проверки совместимости", nil
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.CompatibilityRequest{
		TraceId:            traceID,
		Category:           category,
		ExistingCategories: existingCategories,
	}

	log.Printf("[TraceID: %s] Отправка запроса в Service B: категория='%s', existing=%v",
		traceID, category, existingCategories)

	resp, err := c.client.CheckCompatibility(ctx, req)
	if err != nil {
		return c.handleGRPCError(traceID, err)
	}

	log.Printf("[TraceID: %s] Получен ответ от Service B: compatible=%v, message='%s'",
		traceID, resp.Compatible, resp.Message)

	return resp.Compatible, resp.Message, nil
}

func (c *CompatibilityClient) handleGRPCError(traceID string, err error) (bool, string, error) {
	st, ok := status.FromError(err)

	if !ok {
		log.Printf("[TraceID: %s] ❌ Не gRPC ошибка: %v", traceID, err)
		c.healthy.Store(false)
		return true, fmt.Sprintf("Сервис проверки совместимости недоступен (ошибка: %v), товар добавлен без проверки", err), nil
	}

	switch st.Code() {
	case codes.InvalidArgument:
		log.Printf("[TraceID: %s] ❌ Некорректный запрос: %v", traceID, st.Message())
		return false, fmt.Sprintf("Ошибка запроса к сервису проверки: %s", st.Message()),
			fmt.Errorf("gRPC InvalidArgument: %s", st.Message())

	case codes.NotFound:
		log.Printf("[TraceID: %s] ⚠️ Ресурс не найден: %v", traceID, st.Message())
		return true, "Сервис проверки временно не может определить совместимость, товар добавлен", nil

	case codes.Internal:
		log.Printf("[TraceID: %s] ❌ Внутренняя ошибка Service B: %v", traceID, st.Message())
		c.healthy.Store(false)
		return true, "Сервис проверки совместимости сообщил о внутренней ошибке, товар добавлен без проверки", nil

	case codes.Unavailable:
		log.Printf("[TraceID: %s] ⚠️ Service B недоступен: %v", traceID, st.Message())
		c.healthy.Store(false)
		return true, "Сервис проверки совместимости временно недоступен, товар добавлен без проверки", nil

	default:
		log.Printf("[TraceID: %s] ❌ Неизвестная ошибка gRPC (код %v): %v", traceID, st.Code(), st.Message())
		c.healthy.Store(false)
		return true, "Неизвестная ошибка сервиса проверки, товар добавлен без проверки", nil
	}
}

func (c *CompatibilityClient) GetCategories(ctx context.Context, traceID string) ([]string, error) {
	if c == nil || !c.isAvailable() {
		log.Printf("[TraceID: %s] Service B недоступен, возвращаем fallback категории", traceID)
		return c.getFallbackCategories(), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.GetCategories(ctx, &emptypb.Empty{})
	if err != nil {
		// Обработка gRPC ошибок для GetCategories
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Internal:
				log.Printf("[TraceID: %s] ❌ Внутренняя ошибка при получении категорий: %v", traceID, st.Message())
				return c.getFallbackCategories(), nil
			case codes.NotFound:
				log.Printf("[TraceID: %s] ⚠️ Категории не найдены, используем fallback", traceID)
				return c.getFallbackCategories(), nil
			case codes.Unavailable:
				log.Printf("[TraceID: %s] ⚠️ Service B недоступен, используем fallback категории", traceID)
				c.healthy.Store(false)
				return c.getFallbackCategories(), nil
			default:
				log.Printf("[TraceID: %s] Ошибка получения категорий: %v, используем fallback", traceID, err)
				return c.getFallbackCategories(), nil
			}
		}

		log.Printf("[TraceID: %s] Не удалось получить категории: %v, используем fallback", traceID, err)
		return c.getFallbackCategories(), nil
	}

	if len(resp.Categories) == 0 {
		log.Printf("[TraceID: %s] Service B вернул пустой список категорий, используем fallback", traceID)
		return c.getFallbackCategories(), nil
	}

	return resp.Categories, nil
}

func (c *CompatibilityClient) getFallbackCategories() []string {
	return []string{"Food", "Electronics", "Chemicals"}
}

func (c *CompatibilityClient) IsServiceBAvailable() bool {
	return c != nil && c.isAvailable()
}

// Метод для ручного сброса состояния
func (c *CompatibilityClient) ResetHealth() {
	c.healthy.Store(true)
	c.failCount.Store(0)
	log.Printf("Health status reset for Service B")
}
