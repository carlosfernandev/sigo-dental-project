# Plan de Diseño Frontend: SIGO-Dental

Dirección visual: **papel milimetrado clínico**. La interfaz usa la cuadrícula propia del trabajo odontológico (odontograma, agenda, tablas) como lenguaje visual: ángulos rectos, líneas finas y pocos colores con significado.

Referencia: interfaz de ficha de paciente con odontograma (contexto a la izquierda, trabajo al centro, presupuesto a la derecha).
Requisitos del SRS relacionados: RF-02, RF-03, RF-09, RF-10, RNF-01, RNF-07, RN-01.

## 1. Principios

1. **Contexto clínico siempre visible, lo secundario bajo demanda.** Nombre, edad y alertas médicas no desaparecen nunca.
2. **Una acción primaria por pantalla.** Un solo botón relleno; el resto es secundario o de texto.
3. **Pensado para el sillón.** Objetivos táctiles de al menos 44 px (guantes, tablet).
4. **El color significa algo.** Rojo, verde, ciruela y amarillo se reservan para estados clínicos, financieros y de selección; la interfaz usa neutros y un solo acento.
5. **La estructura es información.** Toda superficie es rectangular (radio 0) y se separa con líneas, no con sombras ni degradados. El grosor de la línea indica jerarquía: 1 px estructura, 2 px elemento interactivo o contenedor flotante, 3 px foco.
6. **Nada se comunica solo con color.** Cada estado lleva color, trama y letra o texto.

## 2. Paleta (tokens)

| Token | Hex | Uso |
|---|---|---|
| `ink-900` | `#17231C` | Texto principal, bordes de contenedores flotantes (contraste sobre blanco superior a 15:1) |
| `ink-600` | `#46554B` | Texto secundario (7.9:1 sobre blanco) |
| `paper` | `#FBFCFA` | Fondo general |
| `surface` | `#FFFFFF` | Paneles y celdas |
| `grid-line` | `#CFD8D0` | Líneas de cuadrícula decorativas y divisiones entre filas |
| `border-strong` | `#5A6B60` | Borde de campos, botones secundarios y celdas interactivas (5.7:1, cumple 3:1 para componentes) |
| `primary-700` | `#1D6A4B` | Botón primario, enlaces, restauraciones, estados "realizado" (6.5:1 con texto blanco) |
| `primary-50` | `#E6F1EB` | Fondo suave del acento, fila con hover |
| `highlight` | `#FFD23F` | Exclusivo de selección y foco: ítem de menú activo, diente seleccionado, fila seleccionada. Siempre con texto `ink-900` |
| `danger-700` / `danger-50` | `#B3261E` / `#FBE4E2` | Alergias, caries, vencido, acciones destructivas (6.5:1) |
| `plan-700` / `plan-50` | `#6A2E6F` / `#F3E8F3` | Tratamiento planificado o presupuestado |
| `warning-800` / `warning-50` | `#6B4200` / `#FFF1C2` | Notas y avisos (7.7:1); siempre con icono y la palabra "Aviso" |

Ajustes respecto a la versión anterior:
- Se elimina el azul de la interfaz: la navegación activa pasa de azul marino a amarillo marcador con texto oscuro, y las restauraciones existentes pasan de azul a verde con trama.
- El teal del botón principal se cambia por `primary-700`, que cumple contraste con texto blanco.
- No hay degradados ni fondos pastel decorativos. Los pasteles (`danger-50`, `plan-50`, `warning-50`) solo aparecen detrás de contenido real.
- Etiquetas, notas y alergias se muestran únicamente si tienen contenido.
- `highlight` no se usa como aviso. Los avisos usan `warning-50` con icono y texto, para que el amarillo siga significando "esto está seleccionado".

## 3. Distribución

```
┌────────────────────┬──────────────────────────────────────────────┐
│ Sidebar (20%)      │ Navbar superior (80%)                        │
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

- **Proporción:** sidebar del 20% del ancho y zona de trabajo del 80%, separados por una línea de 1 px (`border-strong`).
- **Navbar:** ocupa el ancho de la zona de trabajo; el contenido de cada pantalla queda debajo y puede desplazarse sin mover la navegación. Tiene una línea inferior de 1 px.
- **Sidebar:** ocupa la altura disponible y su perfil permanece en un contenedor anclado al fondo, con nombre de usuario y rol.
- **Rejilla:** 12 columnas con base de 8 px. El contenido va alineado a la izquierda; el texto nunca se justifica y las líneas de lectura miden menos de 80 caracteres.
- **Escritorio (1920 y 1366):** sidebar 20% y zona de trabajo 80%; el contexto o panel derecho, cuando aplique, vive dentro de la zona de trabajo y es colapsable.
- **Tablet (>= 10.5"):** menú reducido a iconos (72 px) con etiqueta de texto bajo cada icono; el panel derecho pasa a drawer o bottom sheet.
- **Alertas médicas:** franja fija bajo el nombre del paciente, con chips rectangulares de fondo `danger-700`, texto blanco, icono y nombre de la alergia. Ejemplo: `[!] Alergia: penicilina`.
- **Etiquetas y Notas:** un solo bloque colapsable, cerrado si está vacío.
- **Tablas:** cabecera con borde inferior de 2 px, celdas con líneas de 1 px (`grid-line`), filas de 48 px. Fila seleccionada con fondo `highlight` y texto `ink-900`. Acciones secundarias visibles al pasar el cursor, al enfocar con teclado o con la fila seleccionada.
- **Menú lateral activo:** fondo `highlight`, texto en negrita y borde de 2 px `ink-900`. Los demás ítems muestran icono y texto siempre.
- **Espaciado:** cuadrícula de 8 px; paneles planos, con radio 0 y sin sombra.

## 4. Navegación con poca carga visual

- Menú lateral de 6 a 7 ítems máximo (Filiación, Historia clínica, Odontograma, Ortodoncia, Estado de cuenta, Archivos).
- El paciente es el contexto fijo: al cambiar de módulo solo cambia la zona central.
- Tabs de segundo nivel: 3 a 5 por módulo, como celdas rectangulares unidas en una fila. La activa lleva fondo `highlight` y borde inferior abierto hacia el contenido; las demás, fondo `surface` y borde `border-strong`.
- Divulgación progresiva: herramientas avanzadas detrás de "Más".
- Skeleton loaders en bloques rectangulares en lugar de spinners; conservar el estado del odontograma al cambiar de pestaña.
- Transiciones de 150 a 200 ms, solo opacidad. Con `prefers-reduced-motion` se eliminan.
- Buscador global de pacientes (`/` o `Ctrl+K`).
- Orden de tabulación igual al orden visual, con enlace "Saltar al contenido" al inicio.

## 5. Elementos flotantes

| Elemento | Uso | Aspecto | Bloquea |
|---|---|---|---|
| Tooltip | Etiquetar un icono | Fondo `ink-900`, texto blanco, radio 0 | No |
| Popover / dropdown | Opciones rápidas, estado de un diente | Blanco, borde 2 px `ink-900`, sin sombra | No |
| Drawer lateral (480 px) | Editar hallazgo, detalle de presupuesto | Entra por la derecha, borde izquierdo 2 px `ink-900`, scrim 30% | Parcial |
| Modal centrado | Consentimiento, firma, confirmaciones | Borde 2 px `ink-900`, radio 0, scrim 50% | Sí |
| Diálogo destructivo | Acciones irreversibles | Franja superior `danger-700` con icono y texto, botón rojo con etiqueta explícita | Sí |
| Toast | Confirmar una acción | Abajo a la derecha, 8 s, pausa al pasar el cursor o enfocar, con "Deshacer" si aplica | No |

Reglas:
1. Un solo modal a la vez; no abrir un modal desde otro (usar pasos dentro del mismo).
2. Todo flotante cierra con `Esc`, devuelve el foco y lo atrapa mientras está abierto.
3. Avisar antes de cerrar si hay datos sin guardar.
4. Por RN-01, los hallazgos "Realizados" no se borran: se ofrece "Corregir con adenda" (drawer con firma).
5. Capas: contenido 0, sticky 10, dropdown 20, drawer 30, modal 40, toast 50.
6. Los toast usan `role="status"` para que los lectores de pantalla los anuncien. Los errores que bloquean usan `role="alert"` y no desaparecen solos.

## 6. Odontograma

Es el elemento más característico del sistema: la cuadrícula clínica convertida en interfaz. El resto de la pantalla se mantiene sobrio para que este componente sea lo que se recuerde.

- Render en SVG (RNF-01: carga inferior a 1.5 s).
- Cada pieza es un cuadrado de 44 px o más, dividido en cinco superficies:

```
 ┌───┬───────┬───┐
 │   │   V   │   │    V vestibular   L lingual
 ├───┼───────┼───┤    M mesial       D distal
 │ M │   O   │ D │    O oclusal
 ├───┼───────┼───┤
 │   │   L   │   │
 └───┴───────┴───┘
```

- Las piezas se alinean en dos hileras de 16 celdas contiguas, con el número de pieza sobre cada celda y una línea central que separa los cuadrantes.
- Interruptor FDI / Universal.
- Zoom por cuadrante en tablet.
- Selección de diente y superficie en un popover; la pieza seleccionada se marca con fondo `highlight` y borde de 3 px. "Marcado múltiple" con barra flotante inferior.
- Leyenda de colores, tramas y letras plegada por defecto.
- Navegación con teclado: el odontograma es una rejilla (`role="grid"`). Las flechas mueven entre piezas, `Enter` abre el popover y `Esc` lo cierra. Cada celda tiene etiqueta para lector de pantalla, por ejemplo "Pieza 16, oclusal, caries, planificada".

Estados (color, trama y letra a la vez):

| Hallazgo | Color | Trama | Letra |
|---|---|---|---|
| Caries | `danger-700` | Relleno sólido | C |
| Restauración existente | `primary-700` | Rayas diagonales | R |
| Sellante | `primary-700` | Cuadrícula cruzada | S |
| Tratamiento planificado | `plan-700` | Puntos | P |
| Pieza ausente | `ink-900` | Cruz diagonal sobre toda la celda | X |

Los hallazgos "Realizados" y los "Planificados" se distinguen también por el contorno: línea continua para realizado y discontinua para planificado.

## 7. Tipografía y accesibilidad

- **Familias:** Archivo (variante semi-ancha) para títulos y nombres; Atkinson Hyperlegible Next para texto, tablas y formularios. Atkinson fue diseñada para baja visión y diferencia bien caracteres parecidos (I, l, 1; O, 0), algo útil en números de pieza e importes.
- **Escala:** cuerpo y formularios 18 px, tablas 16 px, secciones 20-22 px, nombre del paciente 24 px. Preferencia de usuario para escala de texto sin romper la rejilla.
- **Etiquetas de campos:** en sentence case, siempre visibles sobre el campo (no solo placeholder), sin mayúsculas sostenidas.
- **Contraste:** WCAG AA como mínimo en todo el texto; los bordes de componentes interactivos cumplen 3:1 (`border-strong`). El texto principal y los estados críticos superan AAA.
- **Foco:** contorno de 3 px `ink-900` con un anillo exterior `highlight` de 2 px, visible sobre cualquier fondo. Nunca se elimina.
- **Estados:** nunca solo con color. Cada chip incluye icono y texto: `Pagado`, `Pendiente`, `Vencido`; `Confirmada`, `Cancelada`.
- **Movimiento:** se respeta `prefers-reduced-motion`.
- **Errores:** explican qué pasó y cómo corregirlo, junto al campo y con `aria-describedby`. Ejemplo: "Falta la fecha de nacimiento. Escríbela como día/mes/año".
- **Acciones:** los botones nombran lo que hacen ("Guardar nota de evolución", "Registrar pago", "Firmar consentimiento") y mantienen el mismo verbo en la confirmación.
- **Idioma:** español con formatos locales (fecha, moneda).

## 8. Orden de implementación

1. Sistema de diseño: tokens, tipografía, componentes base (botón, input, chip, panel, tabla) con radio 0 y estados de foco.
2. Shell: menú lateral, cabecera del paciente con alertas, tabs, responsive.
3. Capa flotante: tooltip, popover, drawer, modal, alerta, toast.
4. Odontograma: componente SVG, tramas, navegación por teclado, marcado múltiple.
5. Panel de presupuesto: resumen y drawer de detalle.
6. Auditoría de accesibilidad: contraste, teclado, lector de pantalla y zoom al 200%.

## 9. Skills asociadas

| Skill | Cubre |
|---|---|
| `dental-ui-tokens` | Colores, tramas, tipografía, espaciado, accesibilidad |
| `dental-ui-layout` | Shell, navegación, tablas, responsive |
| `dental-ui-overlays` | Tooltip, popover, drawer, modal, alerta, toast |
| `dental-ui-odontogram` | Componente del odontograma |

## 10. Resumen de cambios respecto a la versión anterior

| Antes | Ahora | Motivo |
|---|---|---|
| Radio 12 px (paneles) y 16 px (modal) | Radio 0 en todo | Lenguaje de cuadrícula |
| Sombras media y grande en popover y modal | Sin sombras; borde de 2 px | Jerarquía por línea, no por profundidad |
| Azul marino en menú activo y tooltips | Amarillo marcador con texto oscuro; tooltips en `ink-900` | Sin azul; amarillo = seleccionado |
| Azul para restauraciones | Verde con trama de rayas | Sin azul y distinguible por trama |
| Teal `#0F766E` | Verde encina `#1D6A4B` | Paleta propia y mejor contraste |
| Inter o Plus Jakarta Sans a 14 px | Archivo y Atkinson Hyperlegible Next a 16-18 px | Legibilidad y personalidad |
| Pastiles de tabs | Celdas rectangulares unidas | Coherencia con la cuadrícula |
| Toast de 4 s | 8 s con pausa | Tiempo suficiente para leer y deshacer |
| Foco de 2 px | Foco de 3 px más anillo amarillo | Visibilidad en cualquier fondo |
| Estados por color y símbolo | Color, trama y letra | Accesibilidad para daltonismo y escala de grises |
