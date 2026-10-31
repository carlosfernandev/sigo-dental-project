# SIGO-Dental

Aplicación clínica en desarrollo. El backend es un monolito Go/Fiber con PostgreSQL y Redis; el
frontend es una SPA React/Vite. Para trabajar localmente se ejecutan la infraestructura, el backend
y el frontend en terminales separadas.

## Requisitos

- Go **1.27.1** (versión declarada en `backend/go.mod`).
- Docker con Docker Compose v2.
- Node.js **20.19+** o **22.12+**, y npm.

## Primera ejecución

### 1. Preparar la configuración local

Desde la raíz del repositorio:

```sh
cp deploy/.env.example deploy/.env
```

Edita `deploy/.env` antes de iniciar los servicios:

- Cambia `SIGO_PG_PASSWORD` y `SIGO_JWT_SECRET` por valores locales propios. Puedes generar el
  secreto JWT con `openssl rand -hex 32`.
- La plantilla incluye un administrador de prueba local: `admin@example.com` con contraseña
  `clinic-user-password-2026`. El login muestra estos datos solo durante `npm run dev`.
- Estas credenciales son públicas y exclusivas para desarrollo local. Sustitúyelas antes de usar la
  configuración fuera de una máquina local.
- La contraseña del administrador debe tener al menos 12 caracteres, no ser común ni contener el
  correo electrónico de la cuenta.

No compartas ni subas `deploy/.env`; está excluido del control de versiones.

### 2. Iniciar PostgreSQL y Redis

En una terminal:

```sh
cd deploy
docker compose up -d
docker compose ps
```

PostgreSQL queda disponible en `127.0.0.1:5432` y Redis en `127.0.0.1:6379`. Compose crea la base
de datos `sigo` al inicializar el volumen por primera vez.

### 3. Aplicar migraciones, crear el administrador e iniciar la API

En otra terminal, desde la raíz:

```sh
cd backend
go run ./cmd/migrate up
go run ./cmd/seed-admin
go run cmd/api/main.go
```

La migración se ejecuta antes de iniciar la API y el seed crea el administrador de forma idempotente.
El backend carga automáticamente las variables de `../deploy/.env` cuando se inicia desde `backend/`.
La API escucha en `http://localhost:8080`; puedes comprobarla con:

```sh
curl http://localhost:8080/health
```

### 4. Iniciar el frontend

En una tercera terminal:

```sh
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

Vite muestra la dirección local al iniciar (por defecto, `http://localhost:5173`). El archivo
`frontend/.env.example` configura la URL del backend en `http://localhost:8080`; ajusta
`VITE_API_BASE_URL` en `.env.local` si usas otro puerto.

Abre la dirección local de Vite e inicia sesión con el correo y la contraseña del administrador
configurados en `deploy/.env`. En desarrollo, el formulario también muestra y puede cargar las
credenciales de prueba definidas en `frontend/.env.local`.

## Uso diario

Después de la primera ejecución, conserva `deploy/.env` y puedes iniciar los servicios de esta forma:

1. `cd deploy && docker compose up -d`
2. En otra terminal: `cd backend && go run cmd/api/main.go`
3. En otra terminal: `cd frontend && npm run dev`

Las migraciones y el seed solo necesitan repetirse cuando haya cambios de esquema o cuando debas
crear el administrador inicial. `go run ./cmd/migrate up` es seguro de ejecutar más de una vez.

## Detener el entorno

Detén la API y Vite con `Ctrl+C` en sus terminales. Para detener PostgreSQL y Redis conservando los
datos locales:

```sh
cd deploy
docker compose down
```

## Despliegue en producción con Docker

La configuración de producción vive en `deploy/compose.production.yml`. Publica únicamente HTTP y
HTTPS; PostgreSQL y Redis quedan dentro de la red privada de Compose. Nginx redirige HTTP a HTTPS,
termina TLS 1.3 y enruta `/api` al backend.

Para la guía completa de publicación en un servidor remoto, actualizaciones y respaldos, sigue
[`deploy/PRODUCCION.md`](deploy/PRODUCCION.md).

1. Copia `deploy/.env.production.example` a `deploy/.env.production`. Configura un dominio real,
   `SIGO_CORS_ALLOWED_ORIGINS` con el origen HTTPS del frontend, contraseñas únicas y un secreto JWT
   generado con `openssl rand -hex 32`. El archivo de entorno contiene secretos y no se versiona.
2. Instala certificados confiables para el dominio en `deploy/tls/web/fullchain.pem` y
   `deploy/tls/web/privkey.pem`.
3. Instala el certificado de servidor, su clave privada y el certificado de la CA en
   `deploy/tls/postgres/server.crt`, `deploy/tls/postgres/server.key` y
   `deploy/tls/postgres/ca.crt`. El certificado debe incluir `postgres` como nombre DNS SAN porque
   ese es el host interno de la API. La clave debe ser legible por el usuario PostgreSQL del
   contenedor (UID 70 en la imagen Alpine usada aquí) y tener permisos restringidos (por ejemplo,
   `0600`). Los servicios de backend montan solo la CA y verifican tanto la cadena como el nombre
   del servidor TLS.
4. Desde la raíz del repositorio, construye e inicia los servicios:

   ```sh
   docker compose --env-file deploy/.env.production -f deploy/compose.production.yml up --build -d
   ```

Compose espera a que PostgreSQL esté disponible, aplica migraciones y crea el administrador inicial
de forma idempotente antes de iniciar la API. Redis puede reiniciarse sin bloquear el backend; solo
se usa como caché. Revisa el estado y los logs con:

```sh
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml ps
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml logs -f api web
```

El volumen `pgdata` conserva la base de datos al actualizar contenedores. Para detener los servicios
sin borrar los datos, usa `docker compose ... down` con el mismo archivo de entorno y Compose.

## Comandos útiles

Desde la raíz del repositorio:

```sh
make migrate-up  # aplicar migraciones pendientes
make seed-admin  # crear el administrador configurado en deploy/.env
make lint        # formato Go y go vet
make test        # pruebas Go con detector de condiciones de carrera
make build       # compilar la API
```

## Puertos locales

| Servicio | Dirección |
|---|---|
| Frontend | `http://localhost:5173` |
| API | `http://localhost:8080` |
| Estado de la API | `http://localhost:8080/health` |
| PostgreSQL | `127.0.0.1:5432` |
| Redis | `127.0.0.1:6379` |
