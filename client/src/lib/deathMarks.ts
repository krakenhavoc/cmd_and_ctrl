// deathMarks.ts — the chip a permanent wears when something has already
// decided what killing it does this turn (ADR 0108 §1 and §2): it will be
// exiled instead of dying (Lava Coil, Disintegrate), or it can't be
// regenerated (Incinerate, Whippoorwill; CR 701.19c).

export interface DeathMarkInput {
  exiled_if_it_dies?: string[];
  cant_be_regenerated?: boolean;
}

export interface DeathMarkBadge {
  /** The short badge text. */
  text: string;
  /** The tooltip: each rule in full, with the sources that made it. */
  title: string;
}

/**
 * The badge for a permanent's death marks, or null when it has none. A
 * source named twice (two Lava Coils) is named once.
 */
export function deathMarkBadge(card: DeathMarkInput): DeathMarkBadge | null {
  const sources = [...new Set((card.exiled_if_it_dies ?? []).filter((s) => s !== ""))];
  const exiled = (card.exiled_if_it_dies ?? []).length > 0;
  const noRegen = card.cant_be_regenerated === true;
  if (!exiled && !noRegen) return null;
  const text: string[] = [];
  const title: string[] = [];
  if (noRegen) {
    text.push("NO REGEN");
    title.push("Can't be regenerated this turn");
  }
  if (exiled) {
    text.push("EXILED IF IT DIES");
    const rule = "Exiled instead if it dies this turn";
    title.push(sources.length > 0 ? `${rule} — ${sources.join(", ")}` : rule);
  }
  return { text: text.join(" · "), title: title.join("\n") };
}
