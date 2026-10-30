package patient

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

const maxPatientRequestBytes = 16 * 1024

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return patientUnauthorized()
	}
	var request patientRequest
	if err := decodePatientRequest(c, &request); err != nil {
		return invalidPatientRequest()
	}
	patient, err := request.entity()
	if err != nil {
		return invalidPatientRequest()
	}
	created, err := h.service.Create(c.Context(), principal, c.IP(), patient)
	if err != nil {
		return patientError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": response(created)})
}

func (h *Handler) List(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return patientUnauthorized()
	}
	patients, err := h.service.List(c.Context(), principal, c.IP())
	if err != nil {
		return patientError(err)
	}
	items := make([]patientSummaryResponse, 0, len(patients))
	for _, patient := range patients {
		items = append(items, patientSummaryResponse{
			ID: patient.ID, FirstName: patient.FirstName, LastName: patient.LastName,
			DateOfBirth: patient.DateOfBirth.Format("2006-01-02"), Phone: patient.Phone,
		})
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handler) Get(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return patientUnauthorized()
	}
	patient, err := h.service.Get(c.Context(), principal, c.IP(), c.Params("id"))
	if err != nil {
		return patientError(err)
	}
	return c.JSON(fiber.Map{"data": response(patient)})
}

func (h *Handler) Update(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return patientUnauthorized()
	}
	var request patientRequest
	if err := decodePatientRequest(c, &request); err != nil {
		return invalidPatientRequest()
	}
	patient, err := request.entity()
	if err != nil {
		return invalidPatientRequest()
	}
	patient.ID = c.Params("id")
	updated, err := h.service.Update(c.Context(), principal, c.IP(), patient)
	if err != nil {
		return patientError(err)
	}
	return c.JSON(fiber.Map{"data": response(updated)})
}

func decodePatientRequest(c fiber.Ctx, target any) error {
	contentType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || !strings.EqualFold(contentType, fiber.MIMEApplicationJSON) || len(c.Body()) == 0 || len(c.Body()) > maxPatientRequestBytes {
		return ErrInvalidPatient
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ErrInvalidPatient
	}
	return nil
}

func patientError(err error) error {
	switch {
	case errors.Is(err, authn.ErrForbidden):
		return httpx.NewError(fiber.StatusForbidden, "forbidden", "Permiso insuficiente", err)
	case errors.Is(err, ErrInvalidPatient):
		return invalidPatientRequest()
	case errors.Is(err, ErrPatientNotFound):
		return httpx.NewError(fiber.StatusNotFound, "patient_not_found", "Paciente no encontrado", err)
	default:
		return httpx.NewError(fiber.StatusInternalServerError, "internal_error", "Error interno del servidor", err)
	}
}

func patientUnauthorized() error {
	return httpx.NewError(fiber.StatusUnauthorized, "unauthorized", "Autenticación requerida", nil)
}

func invalidPatientRequest() error {
	return httpx.NewError(fiber.StatusBadRequest, "invalid_request", "Solicitud inválida", nil)
}
