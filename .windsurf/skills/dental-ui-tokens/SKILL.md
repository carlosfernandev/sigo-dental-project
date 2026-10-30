---
name: dental-ui-tokens
description: Paleta, tipografía, espaciado y accesibilidad de SIGO-Dental. Usar siempre que se cree o modifique un componente visual, estilo, color o texto de la interfaz, aunque el usuario no mencione "diseño" o "tokens".
---

# Tokens de diseño SIGO-Dental

Todo color, tamaño y espaciado sale de tokens. Nunca escribir hex ni valores de píxel sueltos en componentes.

## Colores

| Token | Hex | Uso |
|---|---|---|
| ink-900 | #1E293B | Texto principal |
| navy-800 | #1E2A4A | Ítem activo del menú, tooltips |
| slate-500 | #64748B | Texto secundario |
| bg-app | #F4F7FA | Fondo general |
| surface | #FFFFFF | Tarjetas y paneles |
| border | #E2E8F0 | Divisiones |
| primary-700 / primary-50 | #0F766E / #F0FDFA | Botón primario, enlaces, foco / fondos suaves |
| danger-600 / danger-50 | #DC2626 / #FEE2E2 | Alergias, caries, acciones destructivas |
| warning-800 / warning-50 | #92400E / #FEF3C7 | Notas y avisos |
| info-600 | #2563EB | Restauraciones existentes |
| success-700 | #15803D | Pagado, confirmado, sellantes |

Configurar los tokens una sola vez (Tailwind `theme.extend.colors` o variables CSS en `:root`) y consumirlos por nombre.

## Reglas

- **Un solo color de acento** (primary). Rojo, azul y verde son semánticos: no usarlos como decoración.
- **Un solo botón primario relleno por pantalla.** Los demás: secundario (borde) o texto.
- **Nunca comunicar un estado solo con color:** añadir icono, símbolo o texto.
- Texto blanco solo sobre `primary-700`, `navy-800` o `danger-600`.

## Tipografía

- Fuente: Inter o Plus Jakarta Sans, con fallback `system-ui, sans-serif`.
- Cuerpo y tablas 14 px; secciones 16-18 px; nombre del paciente 20 px; etiquetas pequeñas mínimo 12 px.

## Espaciado y forma

- Cuadrícula de 8 px (4 px solo para ajustes finos).
- Tarjetas: radio 12 px, borde `border`, sin sombra. Modales: radio 16 px.
- Objetivo táctil mínimo 44 x 44 px (uso con guantes en tablet).

## Accesibilidad (obligatorio)

- Contraste WCAG AA. Anillo de foco visible de 2 px en `primary-700` en todo elemento interactivo.
- Inputs con `label` asociado; errores con texto claro y accionable, no solo borde rojo.
- Textos en español con formatos locales (fecha dd/mm/aaaa, moneda local).
- Respetar `prefers-reduced-motion`. Transiciones de 150-200 ms, solo opacidad.
