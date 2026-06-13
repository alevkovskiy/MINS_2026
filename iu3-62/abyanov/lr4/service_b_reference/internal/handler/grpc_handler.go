package handler

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"service_b_reference/internal/service"
	pb "warehouse_microservices/proto/compatibility"
)

type GRPCHandler struct {
	pb.UnimplementedCompatibilityServiceServer
	checker *service.CompatibilityChecker
}

func NewGRPCHandler(checker *service.CompatibilityChecker) *GRPCHandler {
	return &GRPCHandler{checker: checker}
}

func (h *GRPCHandler) CheckCompatibility(ctx context.Context, req *pb.CompatibilityRequest) (*pb.CompatibilityResponse, error) {
	traceID := req.TraceId
	if traceID == "" {
		traceID = "unknown"
	}

	if req.Category == "" {
		err := status.Error(codes.InvalidArgument, "category is required")
		log.Printf("[TraceID: %s] ❌ ОШИБКА: возвращаем gRPC статус %d (%s)",
			traceID, codes.InvalidArgument, codes.InvalidArgument.String())
		return nil, err
	}

	log.Printf("[TraceID: %s] >>> CheckCompatibility запрос: категория='%s', существующие категории=%v",
		traceID, req.Category, req.ExistingCategories)

	compatible, incompatible, msg := h.checker.Check(req.Category, req.ExistingCategories)

	log.Printf("[TraceID: %s] <<< Результат: compatible=%v, message='%s', incompatible=%v, gRPC статус=%d (%s)",
		traceID, compatible, msg, incompatible, codes.OK, codes.OK.String())

	return &pb.CompatibilityResponse{
		Compatible:             compatible,
		Message:                msg,
		IncompatibleCategories: incompatible,
	}, nil
}

func (h *GRPCHandler) GetCategories(ctx context.Context, _ *emptypb.Empty) (*pb.CategoriesResponse, error) {
	log.Println(">>> GetCategories запрос")

	categories := h.checker.GetCategories()

	if len(categories) == 0 {
		err := status.Error(codes.NotFound, "no categories found")
		log.Printf("❌ ОШИБКА: возвращаем gRPC статус %d (%s)", codes.NotFound, codes.NotFound.String())
		return nil, err
	}

	log.Printf("<<< GetCategories ответ: %v, gRPC статус=%d (%s)",
		categories, codes.OK, codes.OK.String())

	return &pb.CategoriesResponse{
		Categories: categories,
	}, nil
}

func (h *GRPCHandler) HealthCheck(ctx context.Context, _ *emptypb.Empty) (*pb.HealthResponse, error) {
	log.Println(">>> HealthCheck запрос")

	if !h.checker.IsHealthy() {
		err := status.Error(codes.Unavailable, "service is unhealthy")
		log.Printf("❌ ОШИБКА: возвращаем gRPC статус %d (%s)", codes.Unavailable, codes.Unavailable.String())
		return nil, err
	}

	resp := &pb.HealthResponse{
		Healthy:     true,
		ServiceName: "Reference Service (Service B)",
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	log.Printf("<<< HealthCheck ответ: healthy=%v, service_name=%s, timestamp=%s, gRPC статус=%d (%s)",
		resp.Healthy, resp.ServiceName, resp.Timestamp, codes.OK, codes.OK.String())

	return resp, nil
}
