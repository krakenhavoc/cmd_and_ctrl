// turnRules.ts — the one-line summaries of rules effects that last for
// the turn and change what everyone at the table can do (ADR 0107 §5).

/**
 * The game banner's line for the live "damage can't be prevented this
 * turn" grants (GameView.damage_cant_be_prevented, CR 615.12): empty when
 * there are none, otherwise the rule and the sources that made it, in the
 * order they resolved. A source named twice (two Skullcracks) is named
 * once.
 */
export function damageCantBePreventedLine(sources: readonly string[] | undefined): string {
  if (!sources || sources.length === 0) return "";
  const names = [...new Set(sources.filter((s) => s !== ""))];
  const rule = "Damage can't be prevented this turn";
  return names.length === 0 ? rule : `${rule} — ${names.join(", ")}`;
}
