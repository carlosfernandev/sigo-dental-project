// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"strings"

	"github.com/carlosfpatino/sigo-dental/middleware"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
)

const maxRequestBytes = 16 * 1024

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Operatory handlers
func (h *Handler) CreateOperatory(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request operatoryRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	operatory := request.entity()
	created, err := h.service.CreateOperatory(c.Context(), principal, operatory)
	if err != nil {
		return schedulingError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": operatoryResponseFrom(created)})
}

func (h *Handler) ListOperatories(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	operatories, err := h.service.ListOperatories(c.Context(), principal)
	if err != nil {
		return schedulingError(err)
	}
	items := make([]operatoryResponse, len(operatories))
	for i, o := range operatories {
		items[i] = operatoryResponseFrom(o)
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handler) GetOperatory(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	operatory, err := h.service.GetOperatory(c.Context(), principal, c.Params("id"))
	if err != nil {
		return schedulingError(err)
	}
	return c.JSON(fiber.Map{"data": operatoryResponseFrom(operatory)})
}

func (h *Handler) UpdateOperatory(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request operatoryRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	operatory := request.entity()
	operatory.ID = c.Params("id")
	updated, err := h.service.UpdateOperatory(c.Context(), principal, operatory)
	if err != nil {
		return schedulingError(err)
	}
	return c.JSON(fiber.Map{"data": operatoryResponseFrom(updated)})
}

// Chair handlers
func (h *Handler) CreateChair(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request chairRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	chair := request.entity()
	created, err := h.service.CreateChair(c.Context(), principal, chair)
	if err != nil {
		return schedulingError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": chairResponseFrom(created)})
}

func (h *Handler) ListChairs(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	chairs, err := h.service.ListChairs(c.Context(), principal)
	if err != nil {
		return schedulingError(err)
	}
	items := make([]chairResponse, len(chairs))
	for i, c := range chairs {
		items[i] = chairResponseFrom(c)
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handler) GetChair(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	chair, err := h.service.GetChair(c.Context(), principal, c.Params("id"))
	if err != nil {
		return schedulingError(err)
	}
	return c.JSON(fiber.Map{"data": chairResponseFrom(chair)})
}

func (h *Handler) UpdateChair(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request chairRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	chair := request.entity()
	chair.ID = c.Params("id")
	updated, err := h.service.UpdateChair(c.Context(), principal, chair)
	if err != nil {
		return schedulingError(err)
	}
	return c.JSON(fiber.Map{"data": chairResponseFrom(updated)})
}

// Dentist handlers
func (h *Handler) CreateDentist(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request dentistRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	dentist := request.entity()
	created, err := h.service.CreateDentist(c.Context(), principal, dentist)
	if err != nil {
		return schedulingError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": dentistResponseFrom(created)})
}

func (h *Handler) ListDentists(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	dentists, err := h.service.ListDentists(c.Context(), principal)
	if err != nil {
		return schedulingError(err)
	}
	items := make([]dentistResponse, len(dentists))
	for i, d := range dentists {
		items[i] = dentistResponseFrom(d)
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handler) GetDentist(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	dentist, err := h.service.GetDentist(c.Context(), principal, c.Params("id"))
	if err != nil {
		return schedulingError(err)
	}
	return c.JSON(fiber.Map{"data": dentistResponseFrom(dentist)})
}

func (h *Handler) UpdateDentist(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request dentistRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	dentist := request.entity()
	dentist.ID = c.Params("id")
	updated, err := h.service.UpdateDentist(c.Context(), principal, dentist)
	if err != nil {
		return schedulingError(err)
	}
	return c.JSON(fiber.Map{"data": dentistResponseFrom(updated)})
}

// Appointment handlers
func (h *Handler) CreateAppointment(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request appointmentRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	appointment, err := request.entity()
	if err != nil {
		return invalidRequest()
	}
	created, err := h.service.CreateAppointment(c.Context(), principal, appointment)
	if err != nil {
		return schedulingError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": appointmentResponseFrom(created)})
}

func (h *Handler) ListAppointments(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var chairID, dentistID, patientID *string
	if v := c.Query("chair_id"); v != "" {
		chairID = &v
	}
	if v := c.Query("dentist_id"); v != "" {
		dentistID = &v
	}
	if v := c.Query("patient_id"); v != "" {
		patientID = &v
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	var startDatePtr, endDatePtr *string
	if startDate != "" {
		startDatePtr = &startDate
	}
	if endDate != "" {
		endDatePtr = &endDate
	}
	appointments, err := h.service.ListAppointments(c.Context(), principal, chairID, dentistID, patientID, startDatePtr, endDatePtr)
	if err != nil {
		return schedulingError(err)
	}
	items := make([]appointmentSummaryResponse, len(appointments))
	for i, a := range appointments {
		items[i] = appointmentSummaryResponseFrom(a)
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handler) GetAppointment(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	appointment, err := h.service.GetAppointment(c.Context(), principal, c.Params("id"))
	if err != nil {
		return schedulingError(err)
	}
	return c.JSON(fiber.Map{"data": appointmentResponseFrom(appointment)})
}

func (h *Handler) UpdateAppointment(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request appointmentRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	appointment, err := request.entity()
	if err != nil {
		return invalidRequest()
	}
	appointment.ID = c.Params("id")
	updated, err := h.service.UpdateAppointment(c.Context(), principal, appointment)
	if err != nil {
		return schedulingError(err)
	}
	return c.JSON(fiber.Map{"data": appointmentResponseFrom(updated)})
}

func decodeRequest(c fiber.Ctx, target any) error {
	contentType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || !strings.EqualFold(contentType, fiber.MIMEApplicationJSON) || len(c.Body()) == 0 || len(c.Body()) > maxRequestBytes {
		return ErrInvalidAppointment
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ErrInvalidAppointment
	}
	return nil
}

func schedulingError(err error) error {
	switch {
	case errors.Is(err, authn.ErrForbidden):
		return httpx.NewError(fiber.StatusForbidden, "forbidden", "Permiso insuficiente", err)
	case errors.Is(err, ErrInvalidOperatory), errors.Is(err, ErrInvalidChair), errors.Is(err, ErrInvalidDentist), errors.Is(err, ErrInvalidAppointment):
		return invalidRequest()
	case errors.Is(err, ErrAppointmentNotFound):
		return httpx.NewError(fiber.StatusNotFound, "appointment_not_found", "Cita no encontrada", err)
	case errors.Is(err, ErrChairOverlap):
		return httpx.NewError(fiber.StatusConflict, "chair_overlap", "El sillón ya tiene una cita en este horario", err)
	default:
		return httpx.NewError(fiber.StatusInternalServerError, "internal_error", "Error interno del servidor", err)
	}
}

func unauthorized() error {
	return httpx.NewError(fiber.StatusUnauthorized, "unauthorized", "Autenticación requerida", nil)
}

func invalidRequest() error {
	return httpx.NewError(fiber.StatusBadRequest, "invalid_request", "Solicitud inválida", nil)
}
