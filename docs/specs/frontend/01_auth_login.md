# Frontend RF-01: Autenticación y Layout Base
**Módulo:** `src/features/auth/` y `src/pages/`

## 1. Configuración Base Requerida
- Configurar el enrutador en `App.jsx` o `main.jsx` con dos layouts:
  - `PublicLayout` (para el login).
  - `PrivateLayout` (layout principal con Sidebar de navegación, requiere estar autenticado).
- Configurar un interceptor HTTP (Axios/Fetch) que lea el JWT de `localStorage` y lo inyecte en la cabecera `Authorization: Bearer <token>`.

## 2. Pantalla de Login (`src/pages/Login.jsx`)
- **UI:** Formulario centrado en pantalla sobre fondo *dark slate*. Un input para email, uno para contraseña, y un botón de submit color verde esmeralda de alto contraste.
- **Lógica:**
  - Hacer POST a `http://localhost:3000/auth/login` (ajustar puerto según `.env`).
  - Manejar el envelope de error `{"error": {"message": "..."}}` y mostrar una alerta en pantalla.
  - Al recibir HTTP 200, extraer el JWT, guardarlo en `localStorage`, y redirigir a `/dashboard`.

## 3. Estado de Sesión (`src/store/authStore.js` o Context)
- Necesitamos guardar si el usuario está autenticado y sus datos básicos (rol).
- Crear función de `logout` que borre el token y redirija al `/login`.

## 4. Layout Privado (`src/components/Sidebar.jsx`)
- Crear un menú lateral colapsable (dark theme).
- Enlaces iniciales (Fase 1):
  - Inicio (`/dashboard`)
  - Pacientes (`/patients`)
  - Agenda (`/scheduling`)
  - Presupuestos (`/billing`)
- Añadir botón de "Cerrar Sesión" al final del sidebar.
