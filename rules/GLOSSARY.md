# Domain glossary — SRS (es) → code (en)

Single source of truth for naming. The SRS is Spanish; code, identifiers, comments, and
commit messages are English. Use these exact identifiers so the model, API, and UI stay aligned
with `docs/SRS_SIGO-Dental.md`.

## People and roles

| SRS (es) | Code (en) |
| --- | --- |
| Paciente | `Patient` |
| Tutor legal | `LegalGuardian` |
| Odontólogo / Especialista | `Dentist` |
| Asistente dental | `DentalAssistant` |
| Recepcionista / Cajero | `Receptionist` / `Cashier` |
| Auditor / Contador | `Auditor` / `Accountant` |
| Administrador TI | `SystemAdmin` |

## Clinical record

| SRS (es) | Code (en) |
| --- | --- |
| Expediente clínico | `ClinicalRecord` |
| Anamnesis / historial médico | `MedicalHistory` |
| Alergia | `Allergy` |
| Condición crítica / badge de advertencia | `CriticalAlert` |
| Evolución / nota SOEP | `ProgressNote` (fields: `subjective`, `objective`, `assessment`, `plan`) |
| Adenda | `Addendum` |
| Firma digital | `DigitalSignature` |
| Consentimiento informado | `InformedConsent` |
| Receta | `Prescription` |
| Signos vitales | `VitalSigns` |
| Seguro dental / convenio | `InsurancePlan` |

## Odontogram and periodontogram

| SRS (es) | Code (en) |
| --- | --- |
| Odontograma | `Odontogram` |
| Dentición permanente (32) | `PermanentDentition` |
| Dentición temporal / decidua (20) | `PrimaryDentition` |
| Nomenclatura FDI / Universal | `FdiNotation` / `UniversalNotation` |
| Superficie dental | `ToothSurface` |
| Mesial / distal | `mesial` / `distal` |
| Oclusal / incisal | `occlusal` / `incisal` |
| Vestibular | `buccal` |
| Lingual / palatino | `lingual` / `palatal` |
| Caries | `Caries` |
| Restauración | `Restoration` |
| Endodoncia | `EndodonticTreatment` |
| Implante | `Implant` |
| Prótesis fija / removible | `FixedProsthesis` / `RemovableProsthesis` |
| Exodoncia | `Extraction` |
| Sellante | `Sealant` |
| Periodontograma | `PeriodontalChart` |
| Profundidad al sondaje | `ProbingDepth` |
| Margen gingival | `GingivalMargin` |
| NIC (nivel de inserción clínica) | `ClinicalAttachmentLevel` |
| Furca | `Furcation` |
| Movilidad | `Mobility` |
| Sangrado al sondaje | `BleedingOnProbing` |

## Treatment and scheduling

| SRS (es) | Code (en) |
| --- | --- |
| Plan de tratamiento | `TreatmentPlan` |
| Fase / sesión | `TreatmentPhase` / `TreatmentSession` |
| Presupuesto | `TreatmentQuote` |
| Aprobación total / parcial | `full` / `partial` (on `TreatmentQuote`) |
| Honorarios | `ProfessionalFee` |
| Costo de laboratorio | `LabCost` |
| Consultorio | `Operatory` |
| Sillón dental | `Chair` |
| Turno / cita | `Appointment` |
| Recordatorio | `Reminder` |

Appointment states (`RF-10`): `scheduled`, `confirmed`, `checked_in`, `in_progress`,
`completed`, `cancelled`, `no_show`.

## Billing and inventory

| SRS (es) | Code (en) |
| --- | --- |
| Facturación electrónica | `ElectronicInvoice` |
| Comprobante fiscal | `FiscalDocument` |
| Timbrado | `InvoiceStamp` |
| Pago mixto | `PaymentSplit` |
| Pasarela de pagos | `PaymentGateway` |
| Cierre de caja | `CashClose` |
| Corte de caja | `CashReport` |
| Liquidación de honorarios | `FeeSettlement` |
| Insumo | `Supply` |
| Descargo de insumos | `SupplyConsumption` |
| Inventario | `Inventory` |
| Trazabilidad / bitácora | `AuditTrail` |

## Conventions

- Spanish is reserved for SRS text and end-user clinic copy. Everything else is English.
- Do not introduce synonyms: it is `Chair`, never `DentalUnit`; `Appointment`, never `Booking`.
- Enum values are `snake_case` strings; Go types and JSON fields follow Go/JS conventions.
