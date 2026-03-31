package routes

import (
	"database/sql"

	"shalby_backend/config"
	"shalby_backend/handlers"
	"shalby_backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes
func SetupRoutes(r *gin.Engine, db *sql.DB, cfg *config.Config) {
	// Apply CORS middleware globally
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes (no authentication required)
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/login", handlers.LoginHandler(db, cfg))
		authRoutes.POST("/signup", handlers.SignupHandler(db))
		authRoutes.POST("/logout", middleware.AuthMiddleware(cfg), handlers.LogoutHandler())
		authRoutes.POST("/request-password-reset", handlers.RequestPasswordResetHandler(db))
		authRoutes.POST("/verify-password-otp", handlers.VerifyPasswordOTPHandler(db))
		authRoutes.POST("/reset-password", handlers.ResetPasswordHandler(db))
	}

	// Admin routes
	adminRoutes := r.Group("/api/admin")
	// adminRoutes.Use(middleware.AuthMiddleware(cfg))
	// adminRoutes.Use(middleware.RoleMiddleware("Admin"))
	{
		// Dashboard & Profile
		adminRoutes.GET("/dashboard", handlers.GetAdminDashboard(db))
		adminRoutes.GET("/profile", handlers.GetAdminProfile(db))
		adminRoutes.PATCH("/profile", handlers.UpdateAdminProfile(db))

		// Doctor Management
		adminRoutes.GET("/doctors", handlers.GetAllDoctors(db))
		adminRoutes.GET("/doctors/:id", handlers.GetDoctorByID(db))
		adminRoutes.POST("/doctors", handlers.CreateDoctor(db))
		adminRoutes.PUT("/doctors/:id", handlers.UpdateDoctor(db))
		adminRoutes.DELETE("/doctors/:id", handlers.DeleteDoctor(db))

		// Patient Management
		adminRoutes.GET("/patients", handlers.GetAllPatients(db))
		adminRoutes.POST("/patients", handlers.CreatePatient(db))
		adminRoutes.GET("/patients/:id", handlers.GetPatientDetails(db))
		adminRoutes.PUT("/patients/:id", handlers.UpdatePatient(db))
		adminRoutes.DELETE("/patients/:id", handlers.DeletePatient(db))

		// Medicine Management
		adminRoutes.GET("/medicines", handlers.GetAllMedicines(db))
		adminRoutes.GET("/medicines/stats", handlers.GetMedicineStats(db))
		adminRoutes.POST("/medicines", handlers.CreateMedicine(db))
		adminRoutes.PUT("/medicines/:id", handlers.UpdateMedicine(db))
		adminRoutes.DELETE("/medicines/:id", handlers.DeleteMedicine(db))

		// Settings
		adminRoutes.GET("/settings", handlers.GetSystemSettings(db))
		adminRoutes.PUT("/settings", handlers.UpdateSystemSettings(db))

		// Reports
		adminRoutes.GET("/reports", handlers.GetReports(db))

		// Staff Management
		adminRoutes.POST("/staff", handlers.CreateStaff(db))
		adminRoutes.GET("/staff", handlers.GetAllStaff(db))
		adminRoutes.GET("/staff/role/:role", handlers.GetStaffByRole(db))
		adminRoutes.PUT("/staff/:id", handlers.UpdateStaff(db))
		adminRoutes.PUT("/staff/:id/deactivate", handlers.DeactivateStaff(db))
		adminRoutes.PUT("/staff/:id/activate", handlers.ActivateStaff(db))
		adminRoutes.GET("/doctors/department/:department", handlers.GetDoctorsInDepartment(db))
	}

	// Nurse routes - placeholder, full setup below with receptionist

	// Receptionist routes
	receptionistRoutes := r.Group("/api/receptionist")
	receptionistRoutes.Use(middleware.AuthMiddleware(cfg))
	// receptionistRoutes.Use(middleware.RoleMiddleware("Receptionist"))
	{
		// Profile endpoints
		receptionistRoutes.GET("/profile", handlers.GetReceptionistProfile(db))
		receptionistRoutes.PATCH("/profile", handlers.UpdateReceptionistProfile(db))

		receptionistRoutes.POST("/patients/register", handlers.RegisterPatient(db))
		receptionistRoutes.POST("/appointments/book", handlers.BookAppointment(db))
		receptionistRoutes.POST("/patients/checkin", handlers.CheckInPatient(db))
		receptionistRoutes.GET("/appointments/pending", handlers.GetPendingAppointments(db))

		// New endpoints for enhanced receptionist features
		receptionistRoutes.GET("/appointments", handlers.GetAppointmentsByFilter(db))
		receptionistRoutes.GET("/patients", handlers.GetReceptionistPatients(db))
		receptionistRoutes.GET("/patients/:id", handlers.GetPatientDetails(db))
		receptionistRoutes.PUT("/patients/:id", handlers.UpdatePatient(db))
		receptionistRoutes.GET("/doctors", handlers.GetReceptionistDoctors(db))
		receptionistRoutes.GET("/beds", handlers.GetAllBeds(db))
		receptionistRoutes.PUT("/beds/:bedId/reassign", handlers.ReassignBed(db))
		receptionistRoutes.POST("/beds/:bedId/discharge", handlers.DischargeBedPatient(db))
		receptionistRoutes.GET("/beds/:bedId/discharge-info", handlers.GetBedDischargeInfo(db))
		receptionistRoutes.POST("/nurse-tasks", handlers.CreateNurseTask(db))
		receptionistRoutes.GET("/nurse-tasks", handlers.GetNurseTasks(db))
		receptionistRoutes.PUT("/nurse-tasks/:taskId", handlers.UpdateNurseTask(db))
		receptionistRoutes.PATCH("/nurse-tasks/:taskId/status", handlers.UpdateTaskStatus(db))
		receptionistRoutes.DELETE("/nurse-tasks/:taskId", handlers.DeleteTask(db))
	}

	// Dropdown/Lookup routes
	dropdownRoutes := r.Group("/api/dropdowns")
	{
		dropdownRoutes.GET("/departments", handlers.GetDepartments(db))
		dropdownRoutes.POST("/departments", handlers.CreateDepartment(db))
		dropdownRoutes.GET("/specializations", handlers.GetSpecializations(db))
		dropdownRoutes.GET("/allergies", handlers.GetAllergies(db))
		dropdownRoutes.GET("/conditions", handlers.GetConditions(db))
		dropdownRoutes.GET("/medicine-categories", handlers.GetMedicineCategories(db))
		dropdownRoutes.POST("/medicine-categories", handlers.CreateMedicineCategory(db))
		dropdownRoutes.GET("/medicine-generic-names", handlers.GetMedicineGenericNames(db))
		dropdownRoutes.GET("/cities", handlers.GetCities(db))
		dropdownRoutes.GET("/roles", handlers.GetRoles(db))
		dropdownRoutes.POST("/roles", handlers.CreateRole(db))
		dropdownRoutes.GET("/beds", handlers.GetBeds(db))
		dropdownRoutes.GET("/beds/occupancy-stats", handlers.GetBedsOccupancyStats(db))
	}

	// Nurse routes
	nurseRoutes := r.Group("/api/nurse")
	nurseRoutes.Use(middleware.AuthMiddleware(cfg))
	{
		// Profile endpoints
		nurseRoutes.GET("/profile", handlers.GetNurseProfile(db))
		nurseRoutes.PATCH("/profile", handlers.UpdateNurseProfile(db))

		// Existing vitals endpoints
		nurseRoutes.POST("/vitals", handlers.RecordVitals(db))
		nurseRoutes.POST("/record-vitals", handlers.RecordVitals(db))
		nurseRoutes.POST("/vitals/admitted/:patientId", handlers.RecordAdmittedPatientVitals(db))
		nurseRoutes.GET("/checked-in-patients", handlers.GetCheckedInPatients(db))
		nurseRoutes.GET("/patient-vitals/:patient_id", handlers.GetPatientVitals(db))
		nurseRoutes.GET("/recent-vitals", handlers.GetRecentVitals(db))

		// New queue, tasks, and patient monitoring endpoints
		nurseRoutes.GET("/queue", handlers.GetNurseQueue(db))
		nurseRoutes.GET("/tasks", handlers.GetMyTasks(db))
		nurseRoutes.PUT("/tasks/:id", handlers.UpdateMyTask(db))
		nurseRoutes.GET("/patients/admitted", handlers.GetAdmittedPatients(db))
		nurseRoutes.GET("/stats", handlers.GetNurseStats(db))
	}

	// Doctor routes
	doctorRoutes := r.Group("/api/doctor")
	doctorRoutes.Use(middleware.AuthMiddleware(cfg))
	{
		doctorRoutes.GET("/profile", handlers.GetDoctorProfile(db))
		doctorRoutes.GET("/dashboard/stats", handlers.GetDoctorDashboardStats(db))
		doctorRoutes.GET("/appointments", handlers.GetDoctorAppointments(db))
		doctorRoutes.GET("/appointments/:appointmentId", handlers.GetDoctorAppointmentDetail(db))
		doctorRoutes.GET("/appointments/:appointmentId/consultation", handlers.GetConsultationByAppointmentID(db))
		doctorRoutes.GET("/patients", handlers.GetDoctorPatients(db))
		doctorRoutes.GET("/patients/:patientId", handlers.GetPatientDetails(db))
		doctorRoutes.GET("/patients/:patientId/vitals", handlers.GetPatientVitals(db))
		doctorRoutes.GET("/patients/:patientId/consultations", handlers.GetConsultationHistory(db))
		doctorRoutes.POST("/consultations/:appointmentId", handlers.CreateConsultation(db))
		doctorRoutes.GET("/consultations/:id", handlers.GetConsultationByID(db))
		doctorRoutes.POST("/prescriptions", handlers.CreatePrescription(db))
		doctorRoutes.GET("/prescriptions/:id", handlers.GetPrescriptionByID(db))
		doctorRoutes.GET("/prescriptions", handlers.GetPrescriptionsByDoctor(db))
	}

	// Pharmacist routes
	pharmacistRoutes := r.Group("/api/pharmacist")
	pharmacistRoutes.Use(middleware.AuthMiddleware(cfg))
	{
		pharmacistRoutes.GET("/profile", handlers.GetPharmacistProfile(db))
		pharmacistRoutes.GET("/dashboard/stats", handlers.GetPharmacyDashboardStats(db))
		pharmacistRoutes.GET("/pending-prescriptions", handlers.GetPendingPrescriptions(db))
		pharmacistRoutes.POST("/dispense/:id", handlers.DispensePrescription(db))
		pharmacistRoutes.GET("/medicines", handlers.GetPharmacyMedicines(db))
		pharmacistRoutes.GET("/medicines/low-stock", handlers.GetLowStockMedicines(db))
		pharmacistRoutes.GET("/medicines/expiring", handlers.GetExpiringMedicines(db))
	}
}
