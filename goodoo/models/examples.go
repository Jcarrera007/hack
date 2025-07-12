package models

import (
	"time"
)

// User model - equivalent to res.users in Odoo
type User struct {
	BaseModel
	Name     string    `gorm:"size:255;not null" json:"name"`
	Email    string    `gorm:"size:255;unique;not null" json:"email"`
	Login    string    `gorm:"size:64;unique;not null" json:"login"`
	Password string    `gorm:"size:255" json:"-"`
	Active   bool      `gorm:"default:true" json:"active"`
	LastLogin *time.Time `json:"last_login"`
	
	// Relationships
	PartnerID *uint `gorm:"index" json:"partner_id"`
	Partner   *Partner `gorm:"foreignKey:PartnerID" json:"partner,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "res_users"
}

// Partner model - equivalent to res.partner in Odoo
type Partner struct {
	BaseModel
	Name         string  `gorm:"size:255;not null" json:"name"`
	Email        string  `gorm:"size:255" json:"email"`
	Phone        string  `gorm:"size:64" json:"phone"`
	Street       string  `gorm:"size:255" json:"street"`
	City         string  `gorm:"size:64" json:"city"`
	Zip          string  `gorm:"size:24" json:"zip"`
	CountryID    *uint   `gorm:"index" json:"country_id"`
	IsCompany    bool    `gorm:"default:false" json:"is_company"`
	CustomerRank int     `gorm:"default:0" json:"customer_rank"`
	SupplierRank int     `gorm:"default:0" json:"supplier_rank"`
	
	// Relationships
	ParentID *uint     `gorm:"index" json:"parent_id"`
	Parent   *Partner  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Partner `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	
	// One-to-many: users
	Users []User `gorm:"foreignKey:PartnerID" json:"users,omitempty"`
}

// TableName specifies the table name for Partner
func (Partner) TableName() string {
	return "res_partner"
}

// Product model - equivalent to product.product in Odoo
type Product struct {
	BaseModel
	Name           string  `gorm:"size:255;not null" json:"name"`
	DefaultCode    string  `gorm:"size:64;index" json:"default_code"`
	Barcode        string  `gorm:"size:64;unique" json:"barcode"`
	ListPrice      float64 `gorm:"type:decimal(16,2);default:0" json:"list_price"`
	StandardPrice  float64 `gorm:"type:decimal(16,2);default:0" json:"standard_price"`
	Type           string  `gorm:"size:32;default:'consu'" json:"type"` // 'consu', 'service', 'product'
	Active         bool    `gorm:"default:true" json:"active"`
	SaleOk         bool    `gorm:"default:true" json:"sale_ok"`
	PurchaseOk     bool    `gorm:"default:true" json:"purchase_ok"`
	Weight         float64 `gorm:"type:decimal(8,3);default:0" json:"weight"`
	Volume         float64 `gorm:"type:decimal(8,3);default:0" json:"volume"`
	
	// Relationships
	CategoryID *uint            `gorm:"index" json:"categ_id"`
	Category   *ProductCategory `gorm:"foreignKey:CategoryID" json:"categ,omitempty"`
	
	// Many-to-many relationships with suppliers
	SupplierIDs []uint    `gorm:"-" json:"supplier_ids"`
	Suppliers   []Partner `gorm:"many2many:product_supplierinfo;" json:"suppliers,omitempty"`
}

// TableName specifies the table name for Product
func (Product) TableName() string {
	return "product_product"
}

// ProductCategory model - equivalent to product.category in Odoo
type ProductCategory struct {
	BaseModel
	Name         string `gorm:"size:255;not null" json:"name"`
	CompleteName string `gorm:"size:512" json:"complete_name"`
	
	// Hierarchical structure
	ParentID *uint             `gorm:"index" json:"parent_id"`
	Parent   *ProductCategory  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []ProductCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	
	// One-to-many: products
	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// TableName specifies the table name for ProductCategory
func (ProductCategory) TableName() string {
	return "product_category"
}

// SaleOrder model - equivalent to sale.order in Odoo
type SaleOrder struct {
	BaseModel
	Name        string     `gorm:"size:64;not null;unique" json:"name"`
	State       string     `gorm:"size:32;default:'draft'" json:"state"` // 'draft', 'sent', 'sale', 'done', 'cancel'
	DateOrder   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"date_order"`
	AmountTotal float64    `gorm:"type:decimal(16,2);default:0" json:"amount_total"`
	AmountTax   float64    `gorm:"type:decimal(16,2);default:0" json:"amount_tax"`
	
	// Relationships
	PartnerID *uint   `gorm:"not null;index" json:"partner_id"`
	Partner   Partner `gorm:"foreignKey:PartnerID" json:"partner"`
	
	UserID *uint `gorm:"index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	
	// One-to-many: order lines
	OrderLines []SaleOrderLine `gorm:"foreignKey:OrderID" json:"order_line,omitempty"`
}

// TableName specifies the table name for SaleOrder
func (SaleOrder) TableName() string {
	return "sale_order"
}

// SaleOrderLine model - equivalent to sale.order.line in Odoo
type SaleOrderLine struct {
	BaseModel
	Name          string  `gorm:"size:512;not null" json:"name"`
	ProductQty    float64 `gorm:"type:decimal(16,3);default:1" json:"product_uom_qty"`
	PriceUnit     float64 `gorm:"type:decimal(16,2);default:0" json:"price_unit"`
	PriceSubtotal float64 `gorm:"type:decimal(16,2);default:0" json:"price_subtotal"`
	
	// Relationships
	OrderID   uint      `gorm:"not null;index" json:"order_id"`
	Order     SaleOrder `gorm:"foreignKey:OrderID" json:"order"`
	
	ProductID *uint   `gorm:"index" json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
}

// TableName specifies the table name for SaleOrderLine
func (SaleOrderLine) TableName() string {
	return "sale_order_line"
}

// Example of how to register models
func RegisterDefaultModels() {
	registry := GetRegistry()
	
	registry.Register("res.users", User{})
	registry.Register("res.partner", Partner{})
	registry.Register("product.product", Product{})
	registry.Register("product.category", ProductCategory{})
	registry.Register("sale.order", SaleOrder{})
	registry.Register("sale.order.line", SaleOrderLine{})
}