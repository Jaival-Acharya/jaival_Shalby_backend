package models

import (
	"database/sql"
	"time"
)

// AdminDashboardStats represents dashboard statistics
type AdminDashboardStats struct {
	TotalAppointments     int     `json:"totalAppointments"`
	CompletedAppointments int     `json:"completedAppointments"`
	CancelledAppointments int     `json:"cancelledAppointments"`
	ScheduledAppointments int     `json:"scheduledAppointments"`
	ActivePatients        int     `json:"activePatients"`
	TotalDoctors          int     `json:"totalDoctors"`
	TotalMedicines        int     `json:"totalMedicines"`
	LowStockAlerts        int     `json:"lowStockAlerts"`
	ExpiringMedicines     int     `json:"expiringMedicines"`
	RevenueThisMonth      float64 `json:"revenueThisMonth"`
	AppointmentsThisWeek  int     `json:"appointmentsThisWeek"`
}

type TimelineAppointment struct {
	ID              string `json:"id"`
	AppointmentTime string `json:"appointmentTime"`
	PatientName     string `json:"patientName"`
	DoctorName      string `json:"doctorName"`
	Type            string `json:"type"`
	Status          string `json:"status"`
}

type DashboardResponse struct {
	Stats              AdminDashboardStats   `json:"stats"`
	RecentAppointments []TimelineAppointment `json:"recentAppointments"`
	TopMedicines       []MedicineUsageStats  `json:"topMedicines"`
}

type MedicineUsageStats struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Usage      int    `json:"usage"`
	StockLevel int    `json:"stockLevel"`
}

// SystemSettings represents hospital configuration stored in database
type SystemSettings struct {
	SettingKey   string    `json:"settingKey"`
	SettingValue string    `json:"settingValue"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// HospitalConfig is the complete hospital configuration
type HospitalConfig struct {
	HospitalName            string   `json:"hospitalName"`
	Email                   string   `json:"email"`
	Phone                   string   `json:"phone"`
	Address                 string   `json:"address"`
	City                    string   `json:"city"`
	Country                 string   `json:"country"`
	PostalCode              string   `json:"postalCode"`
	Website                 string   `json:"website"`
	LogoURL                 string   `json:"logoUrl"`
	BannerURL               string   `json:"bannerUrl"`
	Currency                string   `json:"currency"`
	Timezone                string   `json:"timezone"`
	AppointmentSlotDuration int      `json:"appointmentSlotDuration"` // in minutes
	MaxAppointmentsPerSlot  int      `json:"maxAppointmentsPerSlot"`
	WorkingHours            string   `json:"workingHours"` // "9:00-17:00"
	Holidays                []string `json:"holidays"`     // dates closed
	NotificationsEmail      bool     `json:"notificationsEmail"`
	NotificationsSMS        bool     `json:"notificationsSms"`
	NotificationsPush       bool     `json:"notificationsPush"`
}

type SystemSettingsRequest struct {
	HospitalName  string `json:"hospitalName"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	Website       string `json:"website"`
	Currency      string `json:"currency"`
	Timezone      string `json:"timezone"`
	Language      string `json:"language"`
	Notifications struct {
		Email bool `json:"email"`
		SMS   bool `json:"sms"`
		Push  bool `json:"push"`
	} `json:"notifications"`
	Security struct {
		TwoFactor  bool   `json:"twoFactor"`
		AutoLogout string `json:"autoLogout"`
	} `json:"security"`
	Mobile struct {
		OfflineMode bool `json:"offlineMode"`
	} `json:"mobile"`
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID           string         `json:"id"`
	UserID       sql.NullString `json:"userId"`
	UserName     string         `json:"userName"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resourceType"`
	ResourceID   string         `json:"resourceId"`
	Details      sql.NullString `json:"details"`
	IPAddress    string         `json:"ipAddress"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type AuditLogListResponse struct {
	Logs  []AuditLog `json:"logs"`
	Total int        `json:"total"`
}

// ReportFilter for generating reports
type ReportFilter struct {
	StartDate string `json:"startDate" form:"startDate"`
	EndDate   string `json:"endDate" form:"endDate"`
	Type      string `json:"type" form:"type"` // appointment, medicine, revenue, etc
	Status    string `json:"status" form:"status"`
}

// Report represents generated report data
type Report struct {
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	StartDate   string      `json:"startDate"`
	EndDate     string      `json:"endDate"`
	Data        interface{} `json:"data"`
	Summary     interface{} `json:"summary"`
	GeneratedAt time.Time   `json:"generatedAt"`
}

type AppointmentReport struct {
	Total      int `json:"total"`
	Scheduled  int `json:"scheduled"`
	Completed  int `json:"completed"`
	Cancelled  int `json:"cancelled"`
	InProgress int `json:"inProgress"`
}

type MedicineReport struct {
	TotalMedicines    int     `json:"totalMedicines"`
	TotalValue        float64 `json:"totalValue"`
	LowStockItems     int     `json:"lowStockItems"`
	ExpiringItems     int     `json:"expiringItems"`
	DiscontinuedItems int     `json:"discontinuedItems"`
}

type RevenueReport struct {
	TotalRevenue     float64 `json:"totalRevenue"`
	ConsultationFees float64 `json:"consultationFees"`
	MedicineSales    float64 `json:"medicineSales"`
	OtherServices    float64 `json:"otherServices"`
	AverageFee       float64 `json:"averageFee"`
}
