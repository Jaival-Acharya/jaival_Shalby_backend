package models

import (
	"database/sql"
	"time"
)

// Medicine represents a medicine/drug in the pharmacy
type Medicine struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	GenericName  string    `json:"genericName"`
	Category     string    `json:"category"`
	DosageForm   string    `json:"dosageForm"`
	Manufacturer string    `json:"manufacturer"`
	Price        float64   `json:"price"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}

// MedicineInventory represents stock information for a medicine
type MedicineInventory struct {
	ID              string         `json:"id"`
	MedicineID      string         `json:"medicineId"`
	StockQuantity   int            `json:"stockQuantity"`
	Unit            string         `json:"unit"`
	ReorderLevel    int            `json:"reorderLevel"`
	ExpiryDate      sql.NullString `json:"expiryDate"` // nullable DATE
	LastRestockedAt sql.NullTime   `json:"lastRestockedAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

// MedicineWithStock combines medicine info with inventory
type MedicineWithStock struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	GenericName     string    `json:"genericName"`
	Category        string    `json:"category"`
	DosageForm      string    `json:"dosageForm"`
	Manufacturer    string    `json:"manufacturer"`
	Price           float64   `json:"price"`
	IsActive        bool      `json:"isActive"`
	StockQuantity   int       `json:"stockQuantity"`
	Unit            string    `json:"unit"`
	ReorderLevel    int       `json:"reorderLevel"`
	ExpiryDate      string    `json:"expiryDate"`
	LastRestockedAt string    `json:"lastRestockedAt"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// MedicineRequest is for updating just medicine info
type MedicineRequest struct {
	Name         string  `json:"name" binding:"required"`
	GenericName  string  `json:"genericName" binding:"required"`
	Category     string  `json:"category" binding:"required"`
	DosageForm   string  `json:"dosageForm" binding:"required"`
	Manufacturer string  `json:"manufacturer"`
	Price        float64 `json:"price" binding:"required,gt=0"`
	IsActive     bool    `json:"isActive"`
}

// MedicineInventoryRequest is for inventory-only updates
type MedicineInventoryRequest struct {
	StockQuantity int    `json:"stockQuantity" binding:"required,gt=0"`
	Unit          string `json:"unit" binding:"required"`
	ReorderLevel  int    `json:"reorderLevel" binding:"required,gt=0"`
	ExpiryDate    string `json:"expiryDate"`
}

// CreateMedicineRequest combines medicine and inventory data for creation
type CreateMedicineRequest struct {
	Name          string  `json:"name" binding:"required"`
	GenericName   string  `json:"genericName" binding:"required"`
	Category      string  `json:"category" binding:"required"`
	DosageForm    string  `json:"dosageForm" binding:"required"`
	Manufacturer  string  `json:"manufacturer"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	IsActive      bool    `json:"isActive"`
	StockQuantity int     `json:"stockQuantity" binding:"required,gt=0"`
	Unit          string  `json:"unit" binding:"required"`
	ReorderLevel  int     `json:"reorderLevel" binding:"required,gt=0"`
	ExpiryDate    string  `json:"expiryDate"`
}

// MedicineResponse is returned when creating a medicine
type MedicineResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	GenericName  string    `json:"genericName"`
	Category     string    `json:"category"`
	DosageForm   string    `json:"dosageForm"`
	Manufacturer string    `json:"manufacturer"`
	Price        float64   `json:"price"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}

// MedicineListResponse contains list of medicines
type MedicineListResponse struct {
	Medicines []MedicineWithStock `json:"medicines"`
	Total     int                 `json:"total"`
}

// MedicineStatsResponse contains medicine statistics
type MedicineStatsResponse struct {
	TotalMedicines    int     `json:"totalMedicines"`
	LowStockCount     int     `json:"lowStockCount"`
	TotalValue        float64 `json:"totalValue"`
	ExpiringCount     int     `json:"expiringCount"`
	DiscontinuedCount int     `json:"discontinuedCount"`
}
