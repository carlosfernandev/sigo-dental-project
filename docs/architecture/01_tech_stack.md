# Stack y arquitectura

## Decisiones

| Área | Decisión |
| --- | --- |
| Frontend | React SPA; PWA en Fase 2 |
| Backend | Go + Fiber, monolito modular |
| Base de datos | PostgreSQL, una base y esquema `public` para todos los dominios |
| Caché | Redis; nunca fuente de verdad |
| Mensajería | Sin procesamiento asíncrono; llamadas síncronas dentro del monolito |
| Almacenamiento de objetos | MinIO, diferido hasta que documentos clínicos entren en alcance |

Revisar versiones disponibles de las dependencias con `go list -m -versions` y `docker pull`
antes de fijarlas. Las versiones de referencia son Fiber v3, PostgreSQL 16, pgx v5 (`pgxpool`),
go-redis v9, `github.com/golang-jwt/jwt/v5` (HS256) y `golang.org/x/crypto/bcrypt`.

## Forma del sistema

Una aplicación React se comunica por HTTPS y JWT Bearer con un solo servidor Fiber. El servidor
atiende los dominios y usa PostgreSQL como fuente de verdad; Redis es opcional y solo acelera
lecturas. La aplicación debe seguir disponible si Redis falla.

```text
frontend/ ──HTTPS/JWT Bearer──▶ monolito Go + Fiber ──▶ PostgreSQL
                                               └─────▶ Redis (caché)
```

Los módulos de dominio se incorporan cuando un flujo real los necesita; no crear módulos vacíos
por anticipado. El punto de partida es `identity`, `patient` y `scheduling`. `billing` y `clinical`
se agregan al iniciar sus primeros flujos. MinIO se implementa al final; no crear antes cargas,
contenedores ni URLs firmadas.

## Límites y dependencias

- Organización por dominio en `backend/internal/<domain>/`; no usar carpetas globales por capa.
- Dirección de dependencias: `api → handler → service → repository interface`.
- El servicio es dueño de autorización, casos de uso y límites transaccionales; el repositorio
  implementa interfaces y no abre transacciones.
- Las llamadas entre dominios son síncronas. Cuando un flujo entre dominios requiere atomicidad,
  el servicio coordina una transacción compartida.
- La autenticación se establece en middleware HTTP; la autorización se comprueba en el servicio
  propietario del caso de uso.
- El código de dominio no importa Fiber, pgx ni Redis. Solo la implementación PostgreSQL importa
  el driver.
- Las rutas HTTP viven en `routes/` dentro del dominio. Los handlers traducen HTTP a llamadas de
  servicio y no contienen reglas de negocio.

Cada dominio puede contener `entity.go`, `service.go`, `repository.go`, `handler.go`, `dto.go` y
`routes/`, según lo necesite. Los tipos de dominio no dependen de frameworks ni drivers.

## Paquetes compartidos y frontend

Los paquetes transversales viven en `backend/platform`, `backend/cache` y `backend/middleware`;
`backend/cmd/api` es la raíz de composición. Las migraciones están en `backend/migrations/`.

El frontend se organiza por funcionalidad en `frontend/src/features/` y componentes compartidos en
`frontend/src/shared/`. Cada funcionalidad es dueña de sus páginas, componentes, API y esquemas.
Los imports entre funcionalidades pasan por `shared/`. El cliente web consume el contrato público
del monolito y no conoce su topología interna. No crear carpetas globales `components/`, `hooks/`,
`services/` ni `pages/`.

## Jurisdicción y alcance externo

La jurisdicción aplicable es Ecuador y la LOPDP; no declarar cumplimiento HIPAA. El MVP opera sin
servicios de terceros: pagos, mensajería, autoridad tributaria y PACS/DICOM no tienen integraciones
externas. Las facturas se emiten internamente y los pagos electrónicos se registran con referencia.
Las fórmulas de comisión de RF-13 siguen pendientes de decisión.
