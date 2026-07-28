package models

type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Shop struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type InvoiceItem struct {
	ID          int     `json:"id,omitempty"`
	InvoiceID   int     `json:"invoice_id,omitempty"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Quantity    float64 `json:"quantity"`
}

type InvoiceShare struct {
	PersonID   int     `json:"person_id"`
	PersonName string  `json:"person_name,omitempty"`
	Share      float64 `json:"share"`
}

type Invoice struct {
	ID         int            `json:"id,omitempty"`
	PersonID   int            `json:"person_id"`
	ShopID     int            `json:"shop_id"`
	Date       string         `json:"date"`
	Total      float64        `json:"total"`
	PersonName string         `json:"person_name,omitempty"`
	ShopName   string         `json:"shop_name,omitempty"`
	Items      []InvoiceItem  `json:"items"`
	Shares     []InvoiceShare `json:"shares"`
}

type InvoiceShareInput struct {
	PersonID int     `json:"person_id" binding:"required"`
	Share    float64 `json:"share"`
}

type CreateInvoiceItem struct {
	Description string  `json:"description" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	Quantity    float64 `json:"quantity"`
}

type CreateInvoiceRequest struct {
	PersonID int                 `json:"person_id" binding:"required"`
	ShopID   int                 `json:"shop_id" binding:"required"`
	Date     string              `json:"date" binding:"required"`
	Items    []CreateInvoiceItem `json:"items" binding:"required,min=1,dive"`
	Shares   []InvoiceShareInput `json:"shares" binding:"required,min=1,dive"`
}

type UpdateInvoiceRequest struct {
	PersonID int                 `json:"person_id" binding:"required"`
	ShopID   int                 `json:"shop_id" binding:"required"`
	Date     string              `json:"date" binding:"required"`
	Items    []CreateInvoiceItem `json:"items" binding:"required,min=1,dive"`
	Shares   []InvoiceShareInput `json:"shares" binding:"required,min=1,dive"`
}

type CreateShopRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateShopRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateItemRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateItemRequest struct {
	Name string `json:"name" binding:"required"`
}

type PersonTotal struct {
	PersonID   int     `json:"person_id"`
	PersonName string  `json:"person_name"`
	Total      float64 `json:"total"`
}

type ShopTotal struct {
	ShopID   int     `json:"shop_id"`
	ShopName string  `json:"shop_name"`
	Total    float64 `json:"total"`
}

type Stats struct {
	Total    float64       `json:"total"`
	ByPerson []PersonTotal `json:"by_person"`
	ByShop   []ShopTotal   `json:"by_shop"`
}
