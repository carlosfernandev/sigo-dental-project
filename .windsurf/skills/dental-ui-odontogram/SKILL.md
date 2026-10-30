---
name: dental-ui-odontogram
description: Componente de odontograma interactivo de SIGO-Dental (RF-03). Usar al construir o modificar el odontograma, marcado de dientes y superficies, leyenda de estados, o el selector FDI/Universal.
---

# Odontograma (RF-03)

Es el componente más crítico en rendimiento y claridad. Combinar con `dental-ui-tokens` y `dental-ui-overlays`.

## Render y rendimiento

- SVG, un componente por diente, memoizado. Meta: carga del expediente y odontograma en menos de 1.5 s (RNF-01).
- Dentición permanente (32) y temporal (20), con interruptor adulto/niño.
- Interruptor de nomenclatura **FDI / Universal** junto al título. Solo cambia las etiquetas, no el modelo de datos (guardar siempre en FDI).

## Superficies

Cada diente tiene 5 superficies seleccionables: mesial, distal, oclusal/incisal, vestibular, lingual/palatino. Objetivo táctil mínimo 44 px; en tablet, zoom por cuadrante.

## Estados: color y símbolo

| Estado | Color | Símbolo |
|---|---|---|
| Caries | `danger-600` | Relleno en la superficie |
| Restauración existente | `info-600` | Relleno sólido |
| Endodoncia | violeta `#7C3AED` | Línea en la raíz |
| Implante | `slate-500` | Tornillo en la raíz |
| Prótesis fija / removible | `slate-500` | Barra sobre la corona |
| Exodoncia | `#475569` | Aspa (X) sobre el diente |
| Sellante | `success-700` | Punto en la oclusal |

Nunca depender solo del color (daltonismo). Leyenda plegada por defecto, siempre accesible.

## Interacción

1. Clic en un diente: popover con las superficies y los estados disponibles.
2. "Marcado múltiple": barra flotante inferior con el estado a aplicar; el usuario toca varios dientes y confirma.
3. Cada cambio se guarda como **hallazgo nuevo** (modelo append-only). Nunca sobrescribir ni borrar.
4. Un hallazgo "Realizado" es de solo lectura (RN-01): corregir con adenda firmada.
5. Estados vacíos y de carga con skeleton, no con spinner.

## Accesibilidad

- Cada diente es un botón con `aria-label` ("Diente 16, oclusal: caries").
- Navegación con flechas entre dientes, `Enter` abre el popover, `Esc` lo cierra.
