// ringEmblem.ts — what the Ring chip says (ADR 0114 owner decision 1,
// #2076). Pure functions, unit-tested; PlayerIdentity.svelte renders
// them.
//
// The Ring is an ordinary emblem (CR 114) with two extra wire fields:
// `level`, how many times the Ring has tempted its owner, and `lines`,
// every line of the emblem with the count it is gained at (CR 701.54c).
// The chip shows the level as a pip ("The Ring · 3"); its hover lists
// all the lines, the gained ones in full and the rest dimmed and marked
// "after the Nth temptation". The marking is words, not only a dimmer
// colour, so a line not yet gained reads as one without colour vision
// and to a screen reader.
//
// Nothing here is Ring-specific by name. An emblem is drawn this way
// when the server sends it lines, which only the Ring does today; any
// other emblem keeps the plain chip with its text as the hover.

import type { CardView, EmblemView, PlayerView } from "./protocol";

export interface EmblemLine {
  text: string;
  // How many temptations it takes (CR 701.54c).
  at: number;
  gained: boolean;
}

// isLevelledEmblem: the emblem gains its lines one by one (the Ring).
export function isLevelledEmblem(e: Pick<EmblemView, "lines" | "level">): boolean {
  return (e.lines?.length ?? 0) > 0 || (e.level ?? 0) > 0;
}

// emblemLines lists every line in the order the emblem gains them
// ("from top to bottom", the 2023-06-16 ruling). A line counts as
// gained once the level reaches its count. With no `lines` on the wire
// (a server from before ADR 0114 PR 2's field), the emblem's text is
// the one line, gained.
export function emblemLines(e: Pick<EmblemView, "lines" | "level" | "text">): EmblemLine[] {
  const level = e.level ?? 0;
  if (!e.lines || e.lines.length === 0) {
    return e.text ? [{ text: e.text, at: 1, gained: true }] : [];
  }
  return [...e.lines]
    .sort((a, b) => a.at - b.at)
    .map((l) => ({ text: l.text, at: l.at, gained: level >= l.at }));
}

// ordinal: 1st, 2nd, 3rd, 4th … 11th, 12th, 13th, 21st.
export function ordinal(n: number): string {
  const tens = n % 100;
  if (tens >= 11 && tens <= 13) return `${n}th`;
  switch (n % 10) {
    case 1:
      return `${n}st`;
    case 2:
      return `${n}nd`;
    case 3:
      return `${n}rd`;
    default:
      return `${n}th`;
  }
}

// pendingNote is the words beside a line not yet gained.
export function pendingNote(at: number): string {
  return `after the ${ordinal(at)} temptation`;
}

// temptedPhrase: "tempted once", "tempted 3 times".
export function temptedPhrase(level: number): string {
  return level === 1 ? "tempted once" : `tempted ${level} times`;
}

// emblemChipName is the chip's accessible name: "The Ring, tempted 3
// times". The pip on screen is a bare number, so the name says what
// it counts.
export function emblemChipName(e: Pick<EmblemView, "label" | "level">): string {
  return `${e.label}, ${temptedPhrase(e.level ?? 0)}`;
}

// emblemDescription is the chip's accessible description: every line,
// each said to be gained or when it will be. A screen reader gets the
// whole hover without having to open it.
export function emblemDescription(e: Pick<EmblemView, "lines" | "level" | "text">): string {
  return emblemLines(e)
    .map((l) => (l.gained ? `Gained: ${l.text}` : `Not yet, ${pendingNote(l.at)}: ${l.text}`))
    .join(" ");
}

// ringBearerTitle is the marker's hover title: "Alice's Ring-bearer".
// With no name (a card drawn somewhere that does not know its seat) it
// is just "Ring-bearer".
export function ringBearerTitle(controllerName: string | undefined): string {
  return controllerName ? `${controllerName}'s Ring-bearer` : "Ring-bearer";
}

// ringBearerNames maps each Ring-bearer on the battlefield to its
// controller's name, for the marker's title. Derived by PlayerPanel,
// which can see the seat list, as takenFrom.ts's chip is.
export function ringBearerNames(
  cards: readonly CardView[] | undefined,
  seats: readonly PlayerView[] | undefined,
): Record<string, string> {
  const names = new Map((seats ?? []).map((s) => [s.id, s.name]));
  const out: Record<string, string> = {};
  for (const c of cards ?? []) {
    if (!c.ring_bearer || !c.controller) continue;
    const name = names.get(c.controller);
    if (name) out[c.instance_id] = name;
  }
  return out;
}
