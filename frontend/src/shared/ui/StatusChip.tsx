type StatusChipProps = {
  label: string;
  tone: "positive" | "negative" | "neutral";
};

export function StatusChip({ label, tone }: StatusChipProps) {
  return <span className={`status-chip status-chip--${tone}`}>{label}</span>;
}
