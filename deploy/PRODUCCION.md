# Despliegue remoto paso a paso

Esta guía despliega SIGO-Dental en un servidor Linux con Docker Compose v2. Usa estos valores de
ejemplo como marcadores y reemplázalos:

- `SERVIDOR`: IP o nombre DNS del servidor.
- `USUARIO_SSH`: usuario con acceso SSH y permisos para ejecutar Docker.
- `DOMINIO`: nombre DNS público que apunta al servidor.
- `/opt/sigo-dental`: ruta del proyecto en el servidor.

No subas `.env.production` ni claves privadas al repositorio.

## 1. Preparar el servidor

Instala Git, Docker Engine y el complemento Docker Compose v2 siguiendo las instrucciones oficiales
para la distribución Linux del servidor. El usuario de despliegue debe poder ejecutar Docker.

Comprueba la instalación:

```sh
ssh USUARIO_SSH@SERVIDOR
docker --version
docker compose version
git --version
```

Configura el DNS de `DOMINIO` para apuntar al servidor y permite tráfico entrante TCP en los puertos
80 y 443 en el firewall del proveedor. No publiques los puertos de PostgreSQL ni Redis.

## 2. Subir el código

Desde tu equipo local, confirma que los cambios estén en el remoto:

```sh
git status
git add .gitignore README.md deploy/PRODUCCION.md deploy/compose.production.yml deploy/.env.production.example backend/Dockerfile backend/platform/config/db.go frontend/Dockerfile frontend/nginx/default.conf
git commit -m "chore(deploy): prepare production docker setup"
git push
```

En el servidor, clona el repositorio (primera instalación):

```sh
sudo mkdir -p /opt/sigo-dental
sudo chown "$USER":"$USER" /opt/sigo-dental
git clone URL_DEL_REPOSITORIO /opt/sigo-dental
cd /opt/sigo-dental
```

Si el repositorio ya está clonado, actualízalo en vez de volver a clonarlo:

```sh
cd /opt/sigo-dental
git pull --ff-only
```

Para repositorios privados, configura previamente el acceso Git del servidor mediante una clave
SSH de despliegue o el método aprobado por tu organización.

## 3. Crear la configuración secreta

En el servidor, desde `/opt/sigo-dental`:

```sh
cp deploy/.env.production.example deploy/.env.production
nano deploy/.env.production
```

Configura al menos:

- `SIGO_CORS_ALLOWED_ORIGINS=https://DOMINIO` (sin ruta final).
- `SIGO_JWT_SECRET`, generado con `openssl rand -hex 32`.
- Contraseñas largas y únicas para `SIGO_PG_PASSWORD` y `SIGO_ADMIN_PASSWORD`.
- El correo del administrador inicial en `SIGO_ADMIN_EMAIL`.

Protege el archivo:

```sh
chmod 600 deploy/.env.production
```

## 4. Instalar los certificados TLS

Prepara las carpetas para los certificados:

```sh
install -d -m 755 deploy/tls/web deploy/tls/postgres
```

Obtén un certificado público válido para `DOMINIO` con el proveedor de certificados elegido. Copia
el certificado completo y su clave privada al servidor en estas rutas:

```text
/opt/sigo-dental/deploy/tls/web/fullchain.pem
/opt/sigo-dental/deploy/tls/web/privkey.pem
```

Deja la clave web solo legible por root:

```sh
chmod 600 deploy/tls/web/privkey.pem
```

PostgreSQL también requiere TLS. Instala en `deploy/tls/postgres/` un certificado de servidor con
`postgres` en su SAN DNS, su clave privada y el certificado CA que lo firmó:

```text
deploy/tls/postgres/server.crt
deploy/tls/postgres/server.key
deploy/tls/postgres/ca.crt
```

La clave `server.key` debe pertenecer al UID 70 y tener modo `600`; la CA debe ser legible por los
contenedores de backend. Asegura permisos de recorrido de la carpeta para que Docker pueda montar
los archivos:

```sh
sudo chown 70:70 deploy/tls/postgres/server.key
sudo chmod 600 deploy/tls/postgres/server.key
sudo chmod 644 deploy/tls/postgres/server.crt deploy/tls/postgres/ca.crt
sudo chmod 755 deploy/tls/postgres
```

No guardes estas claves en Git. Si los certificados los copiaste desde tu equipo, usa `scp` sobre
SSH desde tu equipo local (ajusta las rutas locales de origen) y luego aplica los permisos
anteriores en el servidor:

```sh
scp RUTA_LOCAL/fullchain.pem RUTA_LOCAL/privkey.pem USUARIO_SSH@SERVIDOR:/opt/sigo-dental/deploy/tls/web/
scp RUTA_LOCAL/server.crt RUTA_LOCAL/server.key RUTA_LOCAL/ca.crt USUARIO_SSH@SERVIDOR:/opt/sigo-dental/deploy/tls/postgres/
```

## 5. Construir e iniciar

Desde la raíz del repositorio en el servidor:

```sh
cd /opt/sigo-dental
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml config --quiet
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml up --build -d
```

Compose espera PostgreSQL, aplica las migraciones y crea el administrador inicial antes de iniciar
el API y el frontend. El seed del administrador es idempotente: conserva la cuenta si ya existe.

## 6. Comprobar el despliegue

```sh
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml ps
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml logs --tail=100 postgres migrate seed-admin api web
curl --fail https://DOMINIO/healthz
curl --fail https://DOMINIO/api/health
```

Abre `https://DOMINIO` e inicia sesión con el administrador configurado. Si algo falla, revisa los
logs del servicio afectado y confirma que el DNS, los certificados, los permisos y las variables de
entorno estén correctos.

## 7. Actualizar una instalación existente

Antes de actualizar, crea un respaldo de PostgreSQL. El siguiente comando genera un volcado SQL en el
servidor:

```sh
mkdir -p backups
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB"' > "backups/sigo-$(date +%Y%m%d-%H%M%S).sql"
```

Luego actualiza el código y reconstruye los servicios:

```sh
git pull --ff-only
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml up --build -d
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml ps
```

Las migraciones pendientes se aplican durante el arranque. Conserva el volumen `pgdata`; no uses
`docker compose down -v` en una instalación con datos que quieras mantener.

## 8. Detener

Para detener los contenedores y conservar la base de datos:

```sh
docker compose --env-file deploy/.env.production -f deploy/compose.production.yml down
```
