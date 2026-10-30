# Frontend RF-01 y RF-02: Gestión de Pacientes e Historial Médico
**Módulo:** `src/features/patients/` y `src/pages/`

## 1. Ubicación y Accesibilidad
- **Rutas:**
  - `/patients` (Listado general).
  - `/patients/new` (Formulario de creación).
  - `/patients/:id` (Vista de detalle del paciente).
- **Layout:** `PrivateLayout` (requiere estar autenticado).
- **Control de Acceso:** Accesible para recepcionistas, odontólogos y asistentes (cualquier rol logueado en Fase 1, excepto restricciones específicas que dicte el negocio).

## 2. Pantalla de Listado (`src/pages/PatientsList.jsx`)
- **Diseño:** Tabla de datos (Data Table) siguiendo los estilos globales del proyecto (`04_frontend_stack.md`).
- **Columnas clave:** Nombre, Documento/ID, Teléfono, Seguro/Convenio, Acciones (Ver detalle).
- **Acción principal:** Botón superior "Nuevo Paciente" que redirige a `/patients/new`.
- **Lógica:** `GET /patients`. Mostrar un loader mientras carga y manejar estados vacíos si no hay registros.

## 3. Pantalla de Creación (`src/pages/NewPatient.jsx`)
- **Diseño:** Formulario estructurado en tarjetas (cards) heredando el diseño base. Para no abrumar al usuario, dividir en dos secciones visuales (Datos Personales e Historial Médico).
- **Sección 1: Datos Personales (RF-01)**
  - Nombres, fecha de nacimiento, email, teléfono, dirección.
  - Contacto de emergencia.
  - Tutor legal (Checkbox para habilitar 3 campos obligatorios si es menor de edad).
  - Seguro/Convenio (Select o Input).
- **Sección 2: Historial Médico (RF-02)**
  - Selects múltiples o Checkboxes (Arrays JSON) para: Enfermedades sistémicas, alergias, antecedentes quirúrgicos, coagulopatías.
- **Lógica:**
  - Validar campos requeridos antes del envío.
  - Hacer `POST /patients` (y la petición correspondiente al historial médico si el backend lo requiere en endpoints separados, o todo en un mismo payload según el DTO).
  - Al recibir `200/201`, redirigir al listado `/patients` con un toast de éxito.

## 4. Pantalla de Detalle (`src/pages/PatientDetail.jsx`)
- **Diseño:** Panel informativo.
- **Cabecera Crítica (RF-02):** Si el paciente tiene alergias o coagulopatías reportadas en su historial, mostrar un "Banner/Alerta" visual prominente en la parte superior (usando colores de advertencia/peligro coherentes con la paleta).
- **Lógica:** `GET /patients/:id`.
