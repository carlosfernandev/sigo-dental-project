# Diseño de API

## Contrato

- OpenAPI 3.0 es el contrato público único del monolito y un entregable del SRS. El frontend
  genera `frontend/src/shared/types/api.ts` desde ese documento; no se edita a mano.
- Las rutas se agrupan por dominio en un solo servidor Fiber. Sus declaraciones pertenecen al
  subpaquete `routes` del dominio.
- Los cambios a rutas y formas de respuesta existentes son aditivos. Un cambio incompatible
  requiere una nueva versión de API.
- Las respuestas exitosas usan `{"data": ...}`. Los errores usan
  `{"error":{"code":"...","message":"...","details":...}}`.
- `code` es estable y apto para lógica cliente; `message` es texto humano y no se usa para ramificar.

## Autenticación y autorización

`POST /auth/login` emite un JWT firmado con HS256. Las solicitudes protegidas envían `Authorization:
Bearer <token>`. `middleware.RequireAuth` valida el token y coloca la identidad autenticada en el
contexto de la solicitud.

El middleware resuelve autenticación (AuthN), no permisos. El servicio de cada dominio aplica RBAC
y cualquier filtrado de datos requerido por el caso de uso. RN-02 exige que el total teórico no
forme parte de la respuesta de cierre ciego antes de ingresar el conteo físico; ocultarlo en la UI
no cumple el requisito.

RNF-05 exige cerrar el acceso tras 15 minutos de inactividad. La implementación actual mantiene el
seguimiento en memoria del proceso; conservar esta limitación explícita al cambiar despliegue o
estrategia de sesiones.

## Responsabilidades por capa

- Middleware: validar credenciales de transporte y establecer identidad.
- Handler: parsear HTTP, invocar el caso de uso y mapear el resultado; no construye errores JSON
  ad hoc ni aplica reglas de autorización del dominio.
- Servicio: reglas de negocio, RBAC y límite transaccional.
- Repositorio: persistencia mediante interfaces; no depende de HTTP y no inicia transacciones.
- Un manejador global de errores convierte errores tipados del dominio al envelope y código HTTP.

La lógica de negocio y las validaciones de políticas pertenecen al servicio. Los handlers pueden
validar sintaxis y forma de la solicitud.
