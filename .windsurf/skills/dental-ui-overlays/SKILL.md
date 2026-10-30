---
name: dental-ui-overlays
description: Elementos flotantes de SIGO-Dental (tooltip, popover, drawer, modal, diálogo destructivo, toast) con jerarquía visual clara. Usar al crear confirmaciones, formularios emergentes, paneles laterales, avisos, firmas o consentimientos.
---

# Elementos flotantes

Regla: cuanto más bloquea un elemento, más se separa visualmente del fondo. Combinar con `dental-ui-tokens`.

## Cuál usar

| Elemento | Cuándo | Aspecto | Bloquea |
|---|---|---|---|
| Tooltip | Etiquetar un icono | Fondo `navy-800`, texto blanco | No |
| Popover / dropdown | Opciones rápidas, estado de un diente | Blanco, borde, sombra media, flecha | No |
| Drawer (480 px, derecha) | Editar un hallazgo, detalle de presupuesto | Scrim 30% | Parcial |
| Modal centrado | Consentimiento, firma, decisiones que exigen respuesta | Radio 16 px, sombra grande, scrim 50% | Sí |
| Diálogo destructivo | Acciones irreversibles | Icono y botón `danger`, etiqueta explícita ("Anular presupuesto") | Sí |
| Toast | Confirmar una acción | Abajo a la derecha, 4 s, "Deshacer" si aplica | No |

Si dudas entre modal y drawer, usa drawer. Modal solo cuando el flujo no pueda continuar sin respuesta.

## Reglas obligatorias

1. Un solo modal a la vez. Nunca abrir un modal desde otro: usar pasos dentro del mismo.
2. Cerrar con `Esc` y con clic en el scrim (excepto modales de firma o destructivos).
3. Atrapar el foco mientras está abierto y devolverlo al elemento que lo abrió.
4. Usar los atributos ARIA correctos (`role="dialog"`, `aria-modal`, `aria-labelledby`).
5. Si hay datos sin guardar, avisar antes de cerrar.
6. Botones: acción primaria a la derecha, cancelar a su izquierda. En diálogos destructivos, el botón por defecto en foco es "Cancelar".
7. El toast no sustituye una confirmación: confirmaciones críticas usan diálogo.
8. Escala de capas fija: contenido 0, sticky 10, dropdown 20, drawer 30, modal 40, toast 50.

## Reglas de negocio

- Un hallazgo "Realizado" no se borra (RN-01): ofrecer "Corregir con adenda" en un drawer con firma del profesional.
- Consentimiento informado (RF-06): modal con paso de firma o código OTP, sin cierre por clic fuera.
- Texto de los botones destructivos: verbo + objeto, nunca "Aceptar".
