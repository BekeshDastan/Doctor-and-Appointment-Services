package http

import (
	"net/http"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DoctorHandler struct {
	createDoctorUseCase usecase.CreateDoctorUseCase
	getDoctorById       usecase.GetDoctorByIdUseCase
	getAllDoctors       usecase.GetAllDoctorsUseCase
}

func NewDoctorHandler(
	create usecase.CreateDoctorUseCase,
	getById usecase.GetDoctorByIdUseCase,
	getAll usecase.GetAllDoctorsUseCase,
) *DoctorHandler {
	return &DoctorHandler{
		createDoctorUseCase: create,
		getDoctorById:       getById,
		getAllDoctors:       getAll,
	}
}

func (h *DoctorHandler) RegisterRoutes(router *gin.Engine) {

	doctors := router.Group("/doctors")
	{
		doctors.POST("", h.Create)
		doctors.GET("", h.GetAll)
		doctors.GET("/:id", h.GetByID)
	}
}

func (h *DoctorHandler) Create(c *gin.Context) {
	var req usecase.CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	newDoctor := &model.Doctor{
		ID:             uuid.New().String(),
		FullName:       req.FullName,
		Specialization: req.Specialization,
		Email:          req.Email,
	}

	err := h.createDoctorUseCase.Execute(c.Request.Context(), newDoctor)
	if err != nil {
		if err == usecase.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == usecase.ErrEmptyFields {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newDoctor)
}

func (h *DoctorHandler) GetAll(c *gin.Context) {
	doctors, err := h.getAllDoctors.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch doctors"})
		return
	}

	c.JSON(http.StatusOK, doctors)
}

func (h *DoctorHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	doctor, err := h.getDoctorById.Execute(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "doctor not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doctor)
}
