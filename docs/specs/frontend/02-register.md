# Frontend RF-01b: Registro de Usuarios del Sistema (Staff)
**Módulo:** `src/features/users/` y `src/pages/RegisterUser.jsx`

## 1. Ubicación y Accesibilidad
- **Ruta:** `/users/new` o `/admin/users`.
- **Layout:** `PrivateLayout` (requiere estar autenticado).
- **Control de Acceso (RBAC):** La vista solo debe ser accesible si el `role` del usuario logueado en el estado global (Zustand/Context) es `admin_ti`. Si otro rol intenta entrar, redirigir a `/dashboard` o mostrar un mensaje de "Acceso Denegado".

## 2. Pantalla de Registro de Usuario (UI)
- **Diseño:** Formulario estructurado en una tarjeta (card). Debe heredar estrictamente la paleta de colores, sombras y componentes base definidos en la arquitectura del frontend (`04_frontend_stack.md`).
- **Campos del Formulario:**
  - `email` (Input tipo email, validación de formato).
  - `password` (Input tipo password, será la contraseña temporal).
  - `role_id` (Select/Dropdown con los 5 roles fijos):
    - Administrador TI (`admin_ti`)
    - Recepcionista / Cajero (`recepcionista_cajero`)
    - Odontólogo (`odontologo`)
    - Asistente Dental (`asistente_dental`)
    - Auditor / Contador (`auditor_contador`)
- **Botón de Acción:** "Crear Usuario" (debe deshabilitarse mientras la petición HTTP esté en curso).

## 3. Lógica y Llamada HTTP
- **Endpoint:** `POST /users`.
- **Payload esperado:** `{ "email": "...", "password": "...", "role_id": "..." }`.
- **Autorización:** El interceptor de Axios/Fetch debe inyectar automáticamente el JWT (`Bearer <token>`).
- **Manejo de Respuestas:**
  - **Éxito (200/201):** Mostrar notificación de éxito (toast) indicando que el usuario fue creado y que deberá cambiar su contraseña al iniciar sesión por primera vez (`must_change_password`). Limpiar el formulario tras el éxito.
  - **Error (400/403/500):** Leer el envelope `{"error": {"message": "..."}}` y mostrar el mensaje al usuario de forma clara (toast o alerta bajo el formulario).

## 4. Flujo Futuro (No implementar en esta vista)
- Dado que el backend marca `must_change_password = true`, la lógica de sesión del Login deberá prepararse a futuro para forzar la redirección a una vista de "Actualizar Contraseña" antes de permitir el acceso normal.
