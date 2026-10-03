// turnRules.ts — the one-line summaries of rules effects that last for
// the turn and change what everyone at the table can do (ADR 0107 §5).

/**
 * The game banner's line for the live "damage can't be prevented this
 * turn" grants (GameView.damage_cant_be_prevented, CR 615.12): empty when
 * there are none, otherwise the rule and the sources that made it, in the
 * order they resolved. A source named twice (two Skullcracks) is named
 * once.
 */
/**
 * The game banner's line for the live "if a creature would die this turn,
 * exile it instead" effects over a live set
 * (GameView.exile_if_creatures_die, ADR 0108 §1): empty when there are
 * none. A source named twice is named once.
 */
/**
 * The game banner's lines for the live damage multipliers
 * (GameView.damage_multipliers, ADR 0108 §3): the server words each one
 * ("Alice's sources deal double damage this turn — Insult"), oldest
 * first; this drops blanks. A repeated line is kept: two Insults are two
 * doublings, ×4, and the table should see both.
 */
export function damageMultiplierLines(lines: readonly string[] | undefined): string[] {
  if (!lines) return [];
  return lines.filter((l) => l.trim() !== "");
}

/**
 * The game banner's lines for the live damage redirections
 * (GameView.damage_redirections, ADR 0108 §9): the server words each one
 * ("Damage to Alice from Goblin Guide is dealt to Beacon of Destiny
 * instead, the next time — Beacon of Destiny"), oldest first; this drops
 * blanks. A repeated line is kept: two redirections are two effects.
 */
export function damageRedirectionLines(lines: readonly string[] | undefined): string[] {
  if (!lines) return [];
  return lines.filter((l) => l.trim() !== "");
}

export function exileIfCreaturesDieLine(sources: readonly string[] | undefined): string {
  if (!sources || sources.length === 0) return "";
  const names = [...new Set(sources.filter((s) => s !== ""))];
  const rule = "Creatures that would die this turn are exiled instead";
  return names.length === 0 ? rule : `${rule} — ${names.join(", ")}`;
}

/**
 * The game banner's line for the live source shields
 * (GameView.damage_shields, ADR 0108 §7): "Pay No Heed (Goblin Guide)"
 * and "Healing Grace (Lightning Bolt) — 3 left". Empty when there are
 * none. A line named twice is named once.
 */
export function damageShieldsLine(shields: readonly string[] | undefined): string {
  if (!shields || shields.length === 0) return "";
  const names = [...new Set(shields.filter((s) => s !== ""))];
  const rule = "Damage prevented this turn";
  return names.length === 0 ? rule : `${rule} — ${names.join(", ")}`;
}

export function damageCantBePreventedLine(sources: readonly string[] | undefined): string {
  if (!sources || sources.length === 0) return "";
  const names = [...new Set(sources.filter((s) => s !== ""))];
  const rule = "Damage can't be prevented this turn";
  return names.length === 0 ? rule : `${rule} — ${names.join(", ")}`;
}
