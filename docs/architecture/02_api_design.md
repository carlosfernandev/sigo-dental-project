# Diseño de la API y Manejo de Errores

## Autenticación y Autorización

- **Auth:** JWT (HS256) emitido en `POST /auth/login`.
- **Middleware:** `middleware.RequireAuth` inyecta el usuario y rol en el contexto de Fiber (`c.Locals`).
- **RBAC:** Validación estricta por roles. El token expira/cierra sesión tras 15 minutos de inactividad (controlado en memoria/middleware).

## Estándar de Respuestas HTTP

- **Éxito (200/201):** Retornar JSON limpio, ej: `{"message": "Éxito", "data": {...}}`.
- **Errores Cliente (400, 401, 403, 404):** Errores de validación, no autorizado o no encontrado.
- **Errores Servidor (500):** Fallos de base de datos o lógica no controlada.
- **Formato de Error:** Siempre retornar `{"error": "Descripción clara del error"}`.

## Validaciones

- Los handlers (controladores) solo parsean el JSON (`c.BodyParser`) y validan campos obligatorios básicos.
- La lógica de negocio pesada y las reglas de validación complejas pertenecen exclusivamente a la capa `service.go`.
