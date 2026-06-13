package domain

type Product interface {
	GetID() string
	GetName() string
	GetCategory() string
	GetQuantity() int
	SetQuantity(int)
	GetPrice() float64
	GetDetails() map[string]interface{}
}

type BaseProduct struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func (p *BaseProduct) GetID() string       { return p.ID }
func (p *BaseProduct) GetName() string     { return p.Name }
func (p *BaseProduct) GetCategory() string { return p.Category }
func (p *BaseProduct) GetQuantity() int    { return p.Quantity }
func (p *BaseProduct) SetQuantity(q int)   { p.Quantity = q }
func (p *BaseProduct) GetPrice() float64   { return p.Price }
func (p *BaseProduct) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"id":       p.ID,
		"name":     p.Name,
		"category": p.Category,
		"quantity": p.Quantity,
		"price":    p.Price,
	}
}

type Food struct {
	BaseProduct
	ExpiryDate string `json:"expiry_date"`
}

func (f *Food) GetDetails() map[string]interface{} {
	details := f.BaseProduct.GetDetails()
	details["expiry_date"] = f.ExpiryDate
	return details
}

type Electronics struct {
	BaseProduct
	WarrantyMonths int `json:"warranty_months"`
}

func (e *Electronics) GetDetails() map[string]interface{} {
	details := e.BaseProduct.GetDetails()
	details["warranty_months"] = e.WarrantyMonths
	return details
}

type Chemicals struct {
	BaseProduct
	HazardLevel int `json:"hazard_level"`
}

func (c *Chemicals) GetDetails() map[string]interface{} {
	details := c.BaseProduct.GetDetails()
	details["hazard_level"] = c.HazardLevel
	return details
}

type ProductType string

const (
	FoodType        ProductType = "food"
	ElectronicsType ProductType = "electronics"
	ChemicalsType   ProductType = "chemicals"
)

type SerializableProduct struct {
	Type       ProductType `json:"type"`
	Base       BaseProduct `json:"base"`
	ExpiryDate string      `json:"expiry_date,omitempty"`
	Warranty   int         `json:"warranty_months,omitempty"`
	Hazard     int         `json:"hazard_level,omitempty"`
}

func (sp *SerializableProduct) ToProduct() Product {
	switch sp.Type {
	case FoodType:
		return &Food{
			BaseProduct: sp.Base,
			ExpiryDate:  sp.ExpiryDate,
		}
	case ElectronicsType:
		return &Electronics{
			BaseProduct:    sp.Base,
			WarrantyMonths: sp.Warranty,
		}
	case ChemicalsType:
		return &Chemicals{
			BaseProduct: sp.Base,
			HazardLevel: sp.Hazard,
		}
	}
	return nil
}

func FromProduct(p Product) SerializableProduct {
	sp := SerializableProduct{
		Base: BaseProduct{
			ID:       p.GetID(),
			Name:     p.GetName(),
			Category: p.GetCategory(),
			Quantity: p.GetQuantity(),
			Price:    p.GetPrice(),
		},
	}

	switch v := p.(type) {
	case *Food:
		sp.Type = FoodType
		sp.ExpiryDate = v.ExpiryDate
	case *Electronics:
		sp.Type = ElectronicsType
		sp.Warranty = v.WarrantyMonths
	case *Chemicals:
		sp.Type = ChemicalsType
		sp.Hazard = v.HazardLevel
	}

	return sp
}
