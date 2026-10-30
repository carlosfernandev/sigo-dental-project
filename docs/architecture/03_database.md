# Persistencia y transacciones

## PostgreSQL

- PostgreSQL es la única fuente de verdad. Todos los dominios usan una base de datos y el esquema
  `public`; los límites se aplican en paquetes y mediante coordinación explícita entre servicios.
- Las migraciones son SQL versionado, forward-only, en `backend/migrations/`. No modificar una
  migración aplicada: agregar otra para evolucionar el esquema.
- Cada consulta SQL enumera columnas explícitamente; no usar `SELECT *`, interpolar valores ni
  construir SQL concatenando datos.
- Persistir fechas como `timestamptz` en UTC. Convertir a `America/Guayaquil` en el borde de
  presentación.
- Representar dinero como `int64` en unidades menores (centavos), en USD y sin campo de moneda;
  no usar `float64` ni tipos decimales para cálculos monetarios.
- Los métodos de servicios y repositorios reciben `ctx context.Context` primero. No usar
  `context.Background()` en código de negocio.

## Transacciones y concurrencia

El servicio abre y controla la transacción, y pasa un querier/ejecutor a los repositorios. Los
repositorios no inician ni confirman transacciones. Los flujos que cruzan dominios coordinan una
sola transacción cuando la atomicidad sea necesaria.

- RN-03: impedir solapamientos en el mismo sillón con una restricción de exclusión PostgreSQL; la
  validación de aplicación por sí sola no protege contra concurrencia.
- RN-02: el cierre de caja usa bloqueo pesimista (`SELECT ... FOR UPDATE`) y no revela el total
  teórico antes del conteo físico.
- RF-14: descontar existencias con bloqueo pesimista, salvo que una única sentencia condicional
  exprese por completo la invariante.
- RN-01: los procedimientos realizados son inmutables; las correcciones se registran como adendas
  firmadas, no mediante borrado o actualización destructiva.

## Auditoría clínica

RNF-06 requiere un registro inmutable para cada operación CRUD clínica, con actor, timestamp
ISO-8601 e IP de origen. RF-05 exige notas SOEP firmadas y bloqueadas después del plazo
configurable (24 horas por defecto). Estas reglas deben reflejarse en el modelo y el servicio.

## Redis

Redis es únicamente caché. Toda lectura tiene un camino de fallo hacia PostgreSQL. Los errores de
Redis se registran y se ignoran como fallos de caché; nunca producen un 5xx ni hacen que Redis sea
fuente de verdad.

`GET /health` comprueba PostgreSQL y Redis cuando se usa: PostgreSQL no disponible produce 503;
Redis no disponible produce una respuesta 200 degradada.
