---
name: dental-ui-layout
description: Estructura de pantallas, navegación entre módulos, tablas y comportamiento responsive de SIGO-Dental. Usar al crear o modificar pantallas, menús, pestañas, tablas, tarjetas o paneles, y siempre que se pida una vista nueva del sistema.
---

# Layout y navegación

Objetivo: interfaz amigable con poca carga visual. Combinar con `dental-ui-tokens`.

## Estructura base (ficha de paciente)

```
┌──────────┬──────────────────────────────────┬───────────┐
│ Contexto │  Zona de trabajo (tabs + lienzo) │ Panel de  │
│ paciente │                                  │ contexto  │
│ + menú   │                                  │ (opcional)│
│ 240 px   │            flexible              │  320 px   │
└──────────┴──────────────────────────────────┴───────────┘
```

- El paciente es el contexto fijo. Al cambiar de módulo solo cambia la zona central; el shell no se recarga.
- **Alertas médicas (RF-02):** franja fija bajo el nombre del paciente, chips `danger`. Siempre visibles, nunca dentro de una pestaña.
- Etiquetas y notas: un solo bloque colapsable, cerrado si está vacío. No renderizar tarjetas vacías.

## Navegación

- Menú lateral: máximo 7 ítems, icono + texto, un solo ítem activo (`navy-800`).
- Tabs de segundo nivel: 3 a 5 por módulo, subrayado (no pastillas). Conservar el estado al cambiar de pestaña.
- Herramientas avanzadas detrás de un botón "Más". No mostrar todo a la vez.
- Buscador global de pacientes con `/` o `Ctrl+K`.
- Carga con skeleton loaders, no spinners a pantalla completa.

## Responsive (RNF-07)

| Pantalla | Comportamiento |
|---|---|
| 1920 x 1080 y 1366 x 768 | Tres zonas; panel derecho colapsable |
| Tablet >= 10.5" | Menú de iconos (72 px); panel derecho como drawer o bottom sheet |

Probar siempre en 1366 x 768 y en tablet. Nada de scroll horizontal en el cuerpo de la página.

## Tablas

- Cabecera clara (`bg-app`, texto `slate-500`); el navy se reserva para el menú activo.
- Filas de 48 px. Acciones secundarias visibles al pasar el cursor o con la fila seleccionada; en tablet, siempre visibles.
- Estado vacío con una frase y una acción, nunca una tabla en blanco.
- Cifras alineadas a la derecha; fechas y moneda con formato local.

## Densidad y jerarquía

- Una acción primaria por pantalla.
- Máximo 2 niveles de tarjetas anidadas. Separar con espacio, no con más bordes.
- Reglas de negocio visibles en la UI: un procedimiento "Realizado" no muestra papelera (RN-01), sino "Corregir con adenda".
