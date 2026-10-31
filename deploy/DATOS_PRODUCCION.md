# Datos iniciales y operación en producción

Este documento enumera los módulos disponibles en la versión actual y qué datos requieren para
empezar a operar. Se limita a la Fase 1 indicada en `REQUIREMENTS_CHECKLIST.md`. No se deben cargar
datos ficticios en una base productiva.

## Qué genera el despliegue

Al iniciar el stack de producción, Compose realiza automáticamente estas operaciones:

1. Crea la base vacía y sus tablas mediante las migraciones de `backend/migrations/`.
2. Inserta los cinco roles del sistema: `admin_ti`, `recepcionista_cajero`, `odontologo`,
   `asistente_dental` y `auditor_contador`.
3. Crea el administrador inicial usando `SIGO_ADMIN_EMAIL` y `SIGO_ADMIN_PASSWORD` de
   `deploy/.env.production`.

La migración y el seed no crean pacientes, profesionales, recintos, sillones, procedimientos,
presupuestos, citas, facturas ni pagos. Tampoco se cargan datos de ejemplo: no existe un generador
de datos ficticios para producción. El seed del administrador es idempotente y no modifica una
cuenta que ya exista.

El seed marca la cuenta inicial con cambio de contraseña requerido, pero la versión actual no tiene
una pantalla ni un endpoint de cambio de contraseña. Por eso define una contraseña inicial única,
fuerte y bajo control del administrador en el archivo de entorno; no uses una contraseña temporal
compartida.

Para ejecutar el despliegue completo, sigue [PRODUCCION.md](PRODUCCION.md).

## Módulos disponibles y origen de sus datos

| Módulo | Datos que necesita | Cómo se crean en la versión actual |
| --- | --- | --- |
| Acceso y usuarios | Cuenta inicial y cuentas del personal con su rol | El primer `admin_ti` se crea desde el seed. El administrador registra y gestiona el resto desde Usuarios. |
| Pacientes e historial médico | Identificación, contacto, emergencia, tutor cuando corresponda, seguro/convenio e historial | El personal autorizado registra cada paciente desde Pacientes. El historial se completa en la ficha del paciente. |
| Catálogo de procedimientos | Nombre/código, categoría, descripción y precio base | El personal autorizado registra y mantiene los procedimientos desde Facturación. Carga los precios aprobados por la clínica. |
| Recursos de agenda | Recintos, sillones y dentistas | El personal autorizado crea primero los recintos, luego sus sillones y los profesionales desde Recursos de agenda. |
| Citas y recordatorios | Paciente, dentista, sillón, fecha/hora, duración y nota | Recepción registra las citas desde Agenda. Los recordatorios de 48 h y 24 h se generan internamente según las citas; su gestión está en la bandeja de recordatorios. |
| Presupuestos y aprobaciones | Paciente, procedimientos, cantidades, descuentos autorizados y estado de aprobación | Se crean desde Facturación después de cargar el catálogo y registrar al paciente. La aprobación total o parcial habilita tareas agendables. |
| Facturas y pagos | Factura del paciente y pagos con método, monto y referencia | Se emiten desde Facturación. Los pagos en efectivo, tarjeta, transferencia o seguro se registran; no se procesa dinero mediante una pasarela externa. |
| Cierre de caja | Conteo físico por el responsable y fecha del cierre | El cajero registra el cierre desde Caja al final de su turno. No se precargan cierres ni saldos. |

El dashboard presenta información del sistema; no necesita una carga separada. Los módulos clínicos
de Fase 2 (odontograma, periodontograma, notas clínicas, consentimientos y recetas) y los módulos de
Fase 3 (inventario y liquidación de honorarios) aún no forman parte del conjunto marcado como hecho
en el checklist y no se deben considerar disponibles para generar datos de producción.

## Orden recomendado para cargar datos reales

Después de desplegar y verificar el administrador inicial:

1. Inicia sesión con las credenciales iniciales y comprueba que el acceso funciona. Conserva la
   contraseña bajo control del administrador y no la compartas entre personas.
2. Registra las cuentas individuales del personal y asigna únicamente los roles necesarios. No
   compartas una cuenta entre varias personas.
3. Configura los recintos, sillones y dentistas. Revisa que nombres y especialidades correspondan a
   la operación real de la clínica.
4. Carga el catálogo vigente de procedimientos y sus precios aprobados.
5. Registra pacientes reales y completa la información clínica requerida antes de atenderlos.
6. Crea las citas con el paciente, dentista y sillón correctos; verifica fecha, hora y duración.
7. A partir de la atención real, crea presupuestos, registra sus aprobaciones, emite facturas y
   registra pagos con sus referencias.
8. Al cierre de cada turno, registra el conteo físico en Caja y verifica los cierres.

Respeta las autorizaciones, el consentimiento y las políticas de privacidad de la clínica antes de
incorporar datos personales o médicos. Valida los importes, referencias y asignaciones con el
personal responsable; el sistema no trae el catálogo comercial ni los datos de pacientes de una
clínica preconfigurados.

## Carga masiva y generación de datos

La versión actual no incluye un importador de pacientes, catálogo, profesionales ni citas, y no
incluye un comando para generar registros ficticios. Crea los datos operativos desde las pantallas
del sistema para conservar validaciones y relaciones entre registros. No insertes registros
directamente en PostgreSQL como mecanismo normal de carga: se pueden omitir reglas de negocio,
auditoría o relaciones requeridas.

Si se requiere importar información de otro sistema, primero debe definirse un procedimiento de
migración con mapeo, validación, deduplicación, respaldo y ensayo en un entorno separado. No uses los
datos de pruebas locales como datos iniciales de producción.

## Respaldo antes de cargar datos

Antes de la primera carga y antes de cada actualización, crea y conserva un respaldo fuera del
servidor o en almacenamiento protegido. El comando de respaldo y el proceso de actualización están
documentados en [PRODUCCION.md](PRODUCCION.md).
