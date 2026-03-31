package http

import (
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/usecase"
	"github.com/gin-gonic/gin"
)

type CreateAppointmentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DoctorID    string `json:"doctor_id"`
}

type UpdateStatusRequest struct {
	Status model.Status `json:"status"`
}

type AppointmentHandler struct {
	createUseCase       usecase.CreateAppointmentUseCase
	getUseCase          *usecase.GetAppointmentUseCase
	listUseCase         *usecase.ListAppointmentsUseCase
	updateStatusUseCase *usecase.UpdateStatusUseCase
	idCounter           uint64
}

func NewAppointmentHandler(
	create usecase.CreateAppointmentUseCase,
	get *usecase.GetAppointmentUseCase,
	list *usecase.ListAppointmentsUseCase,
	update *usecase.UpdateStatusUseCase,
) *AppointmentHandler {
	return &AppointmentHandler{
		createUseCase:       create,
		getUseCase:          get,
		listUseCase:         list,
		updateStatusUseCase: update,
		idCounter:           0,
	}
}

func (h *AppointmentHandler) RegisterRoutes(router *gin.Engine) {
	appointments := router.Group("/appointments")
	{
		appointments.POST("", h.Create)
		appointments.GET("", h.GetAll)
		appointments.GET("/:id", h.GetByID)
		appointments.PATCH("/:id/status", h.UpdateStatus)
	}
}

func (h *AppointmentHandler) Create(c *gin.Context) {
	var req CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	newID := atomic.AddUint64(&h.idCounter, 1)

	appointment := &model.Appointment{
		ID:          strconv.FormatUint(newID, 10),
		Title:       req.Title,
		Description: req.Description,
		DoctorID:    req.DoctorID,
	}

	err := h.createUseCase.Execute(c.Request.Context(), appointment)
	if err != nil {
		if err.Error() == "doctorID and aappointment title is required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "The doctor doesn't exist" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appointment)
}

func (h *AppointmentHandler) GetAll(c *gin.Context) {
	appointments, err := h.listUseCase.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch appointments: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, appointments)
}

func (h *AppointmentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	appointment, err := h.getUseCase.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if appointment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		return
	}

	c.JSON(http.StatusOK, appointment)
}

func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	err := h.updateStatusUseCase.Execute(c.Request.Context(), id, req.Status)
	if err != nil {
		if err.Error() == "cannot revert status from 'done' to 'new'" || err.Error() == "invalid status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}
