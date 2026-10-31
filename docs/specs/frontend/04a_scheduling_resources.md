# Frontend RF-09 (Parte 1): Gestión de Recursos Médicos
**Módulo:** `src/features/scheduling/resources/` y `src/pages/`

## 1. Ubicación y Accesibilidad
- **Rutas (Vistas de configuración):**
  - `/admin/operatories` (Recintos/Clínicas)
  - `/admin/chairs` (Sillones)
  - `/admin/dentists` (Profesionales)
- **Layout:** `PrivateLayout`.
- **Control de Acceso (RBAC):** Accesible solo para `admin_ti` o roles gerenciales (según definición de negocio). Los demás roles solo los consumen en modo lectura para agendar.

## 2. Pantallas de Listado (Data Tables)
- **Diseño:** Tablas simples heredando el estilo de `04_frontend_stack.md`.
- **Lógica HTTP (Endpoints a consumir):**
  - `GET /operatories`
  - `GET /chairs` (Debe mostrar a qué operatory pertenece)
  - `GET /dentists` (Mostrar nombre, especialidad y estado activo/inactivo)

## 3. Formularios de Creación/Edición
- **Diseño:** Formularios en Modal o en tarjetas pequeñas, ya que son pocos campos.
- **Lógica HTTP:**
  - `POST /operatories` y `PUT /operatories/:id`
  - `POST /chairs` y `PUT /chairs/:id` (Requiere un select para elegir el `operatory_id`).
  - `POST /dentists` y `PUT /dentists/:id`
