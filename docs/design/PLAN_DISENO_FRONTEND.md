# Plan de Diseño Frontend: SIGO-Dental

Referencia: interfaz de ficha de paciente con odontograma (contexto a la izquierda, trabajo al centro, presupuesto a la derecha).
Requisitos del SRS relacionados: RF-02, RF-03, RF-09, RF-10, RNF-01, RNF-07, RN-01.

## 1. Principios

1. **Contexto clínico siempre visible, lo secundario bajo demanda.** Nombre, edad y alertas médicas no desaparecen nunca.
2. **Una acción primaria por pantalla.** Un solo botón relleno; el resto es secundario o de texto.
3. **Pensado para el sillón.** Objetivos táctiles de al menos 44 px (guantes, tablet).
4. **El color significa algo.** Rojo, azul y verde se reservan para estados clínicos y financieros; la interfaz usa neutros y un solo acento.

## 2. Paleta (tokens)

| Token | Hex | Uso |
|---|---|---|
| `ink-900` | `#1E293B` | Texto principal |
| `navy-800` | `#1E2A4A` | Ítem activo del menú, tooltips |
| `slate-500` | `#64748B` | Texto secundario |
| `bg-app` | `#F4F7FA` | Fondo general |
| `surface` | `#FFFFFF` | Tarjetas y paneles |
| `border` | `#E2E8F0` | Divisiones |
| `primary-700` | `#0F766E` | Botón primario, enlaces, foco |
| `primary-50` | `#F0FDFA` | Fondos suaves del acento |
| `danger-600` / `danger-50` | `#DC2626` / `#FEE2E2` | Alergias, caries, acciones destructivas |
| `warning-50` / `warning-800` | `#FEF3C7` / `#92400E` | Notas y avisos |
| `info-600` | `#2563EB` | Restauraciones existentes |
| `success-700` | `#15803D` | Pagado, confirmado, sellantes |

Ajustes respecto a la referencia:
- El teal claro del botón principal no da contraste suficiente con texto blanco: usar `primary-700`.
- Las tarjetas pastel (etiquetas, notas, alergias) compiten entre sí: se muestran solo si tienen contenido.

## 3. Distribución

```
┌────────────────────┬──────────────────────────────────────────────┐
│ Sidebar (20%)      │ Navbar superior (80%)                       │
│                    ├──────────────────────────────────────────────┤
│ Menú               │                                              │
│                    │ Zona de trabajo (80%):                       │
│                    │ dashboard y pantallas de gestión             │
│                    │                                              │
│                    │                                              │
├────────────────────┤                                              │
│ Perfil             │                                              │
│ Nombre de usuario  │                                              │
│ Rol                │                                              │
└────────────────────┴──────────────────────────────────────────────┘
```

- Proporción: La distribución usa un sidebar del 20% del ancho y una zona de trabajo del 80%.

- Navbar: La navbar superior ocupa el ancho de la zona de trabajo; el contenido de cada pantalla queda debajo y puede desplazarse sin mover la navegación.

- Sidebar: El sidebar ocupa la altura disponible y su perfil permanece en un contenedor anclado al fondo, con nombre de usuario y rol.

- **Escritorio (1920 y 1366):** sidebar 20% y zona de trabajo 80%; el contexto o panel derecho, cuando aplique, vive dentro de la zona de trabajo y es colapsable.
- **Tablet (>= 10.5"):** menú reducido a iconos (72 px); el panel derecho pasa a drawer o bottom sheet.
- **Alertas médicas:** franja delgada fija bajo el nombre del paciente, con chips rojos.
- **Etiquetas y Notas:** un solo bloque colapsable, cerrado si está vacío.
- **Tablas:** cabecera clara, filas de 48 px, acciones secundarias visibles solo al pasar el cursor o con la fila seleccionada.
- **Espaciado:** cuadrícula de 8 px; tarjetas con radio de 12 px, planas (sin sombra).

## 4. Navegación con poca carga visual

- Menú lateral de 6 a 7 ítems máximo (Filiación, Historia clínica, Odontograma, Ortodoncia, Estado de cuenta, Archivos).
- El paciente es el contexto fijo: al cambiar de módulo solo cambia la zona central.
- Tabs de segundo nivel: 3 a 5 por módulo, con subrayado en lugar de pastillas.
- Divulgación progresiva: herramientas avanzadas detrás de "Más".
- Skeleton loaders en lugar de spinners; conservar el estado del odontograma al cambiar de pestaña.
- Transiciones de 150 a 200 ms, solo opacidad.
- Buscador global de pacientes (`/` o `Ctrl+K`).

## 5. Elementos flotantes

| Elemento | Uso | Aspecto | Bloquea |
|---|---|---|---|
| Tooltip | Etiquetar un icono | Fondo navy, texto blanco | No |
| Popover / dropdown | Opciones rápidas, estado de un diente | Blanco, borde, sombra media | No |
| Drawer lateral (480 px) | Editar hallazgo, detalle de presupuesto | Entra por la derecha, scrim 30% | Parcial |
| Modal centrado | Consentimiento, firma, confirmaciones | Radio 16 px, sombra grande, scrim 50% | Sí |
| Diálogo destructivo | Acciones irreversibles | Icono y botón rojos, etiqueta explícita | Sí |
| Toast | Confirmar una acción | Abajo a la derecha, 4 s, con "Deshacer" si aplica | No |

Reglas:
1. Un solo modal a la vez; no abrir un modal desde otro (usar pasos dentro del mismo).
2. Todo flotante cierra con `Esc`, devuelve el foco y lo atrapa mientras está abierto.
3. Avisar antes de cerrar si hay datos sin guardar.
4. Por RN-01, los hallazgos "Realizados" no se borran: se ofrece "Corregir con adenda" (drawer con firma).
5. Capas: contenido 0, sticky 10, dropdown 20, drawer 30, modal 40, toast 50.

## 6. Odontograma

- Render en SVG (RNF-01: carga inferior a 1.5 s).
- Interruptor FDI / Universal.
- Dientes con objetivo táctil de al menos 44 px; zoom por cuadrante en tablet.
- Selección de diente y superficie en un popover; "marcado múltiple" con barra flotante inferior.
- Leyenda de colores y símbolos plegada por defecto.
- Estados con color **y** símbolo (accesibilidad para daltonismo).

## 7. Tipografía y accesibilidad

- Inter o Plus Jakarta Sans. Cuerpo y tablas 14 px, secciones 16-18 px, nombre del paciente 20 px.
- Contraste WCAG AA; anillo de foco de 2 px en `primary-700`; nunca comunicar estados solo con color.
- Español con formatos locales (fecha, moneda).

## 8. Orden de implementación

1. Sistema de diseño: tokens, tipografía, componentes base (botón, input, chip, tarjeta, tabla).
2. Shell: menú lateral, cabecera del paciente con alertas, tabs, responsive.
3. Capa flotante: tooltip, popover, drawer, modal, alerta, toast.
4. Odontograma: componente SVG, estados, marcado múltiple.
5. Panel de presupuesto: resumen y drawer de detalle.

## 9. Skills asociadas

| Skill | Cubre |
|---|---|
| `dental-ui-tokens` | Colores, tipografía, espaciado, accesibilidad |
| `dental-ui-layout` | Shell, navegación, tablas, responsive |
| `dental-ui-overlays` | Tooltip, popover, drawer, modal, alerta, toast |
| `dental-ui-odontogram` | Componente del odontograma |
