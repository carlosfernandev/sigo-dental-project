# Frontend RF-09 y RF-10 (Parte 2): Calendario y Citas
**Módulo:** `src/features/scheduling/appointments/` y `src/pages/`

## 1. Ubicación y Accesibilidad
- **Ruta:** `/scheduling` (Dashboard de la agenda).
- **Layout:** `PrivateLayout`.
- **Control de Acceso:** Recepcionistas, odontólogos y asistentes.

## 2. Pantalla Principal: Calendario (RF-09)
- **Diseño (`SchedulingDashboard.jsx`):**
  - Vista diaria (timeline) o lista ordenada por horas, aplicando el diseño *dark slate / emerald* de `04_frontend_stack.md`.
  - **Filtros requeridos:** Selectores en la parte superior consumiendo `GET /operatories`, `GET /chairs` y `GET /dentists`. Al cambiar un filtro, se debe recargar la lista de citas.
- **Lógica HTTP:**
  - `GET /appointments` (pasar parámetros de fecha para ver el día actual).

## 3. Formulario de Nueva Cita (Modal o Vista)
- **Diseño:** Formulario claro con validación de fechas.
- **Campos y Lógica HTTP:**
  - Paciente: Buscador/Select consumiendo `GET /patients`.
  - Profesional: Select consumiendo `GET /dentists`.
  - Sillón: Select consumiendo `GET /chairs`.
  - `start_time` y `end_time` (Inputs de fecha/hora).
  - `POST /appointments`. Manejar error 400/409 (solapamiento de sillón) mostrando un toast de alerta.

## 4. Gestión de Estados de Cita (RF-10)
- **Diseño:** Al hacer clic en una cita del calendario, abrir un panel o Modal con sus detalles.
- **Transición de Estado:** Incluir botones o un Select para cambiar el `status` (scheduled, confirmed, checked_in, in_progress, completed, cancelled, no_show).
- **Lógica HTTP:** `PUT /appointments/:id`. Actualizar el indicador visual de la cita en el frontend según el nuevo estado.
