# Arquitectura Frontend — SIGO-Dental

La referencia visual y de interacción es [`docs/design/PLAN_DISENO_FRONTEND.md`](../design/PLAN_DISENO_FRONTEND.md).
Si este documento y el plan difieren, mantenerlos sincronizados y usar el plan como guía visual.

## Stack Core
- **Framework:** React 18 + Vite.
- **Estilos:** Tailwind CSS, con los tokens visuales centralizados.
- **Enrutamiento:** React Router DOM (v6+), con layouts público y privado y carga diferida por ruta.
- **Llamadas HTTP:** un cliente en `src/shared/lib/http.ts`; añade el JWT desde `localStorage` y normaliza el envelope de API. Los componentes no llaman `fetch` directamente.
- **Estado Global:** React Context solo para sesión y estado de UI. El estado del servidor usa TanStack Query; no se replica en el contexto.

## Diseño Visual
- **Tema general:** claro, minimalista, legible y de alto contraste. No usar un tema global dark slate/emerald.
- **Fondos y superficies:** `bg-app` (`#F4F7FA`) para el fondo general y `surface` (`#FFFFFF`) para tarjetas y paneles.
- **Texto y divisiones:** `ink-900` (`#1E293B`), `slate-500` (`#64748B`) y `border` (`#E2E8F0`).
- **Acento principal:** `primary-700` (`#0F766E`) y `primary-50` (`#F0FDFA`). El color primario se reserva para acciones y foco; no usar esmeralda como paleta decorativa.
- **Colores semánticos:** peligro, advertencia, información y éxito siguen los tokens del plan de diseño y siempre incluyen texto o símbolo además del color.
- **Componentes:** bordes limpios, tarjetas planas con radio de 12 px, formularios de una columna y una única acción primaria rellena por pantalla.
- **Accesibilidad:** contraste WCAG AA, etiquetas asociadas a inputs, foco visible y objetivos táctiles de al menos 44 px.

## Estructura de Carpetas
Organizar por feature, no por tipo de archivo:

```text
src/
  app/                         # router, layouts y providers
  features/
    auth/                      # login, sesión y schemas
    patients/                  # API, componentes y páginas de pacientes
    scheduling/
    billing/
  shared/
    ui/                        # primitivas presentacionales sin acceso a datos
    lib/                       # cliente HTTP y utilidades compartidas
    types/api.ts               # generado desde OpenAPI; no editar manualmente
```

Las llamadas a la API pertenecen al feature que las consume y usan el cliente compartido. Los
features no se importan entre sí; el código transversal vive en `shared/`.
