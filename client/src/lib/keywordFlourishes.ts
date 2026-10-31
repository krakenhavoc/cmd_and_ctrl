// keywordFlourishes.ts — which keyword flourishes a battlefield card
// wears (#4556).
//
// A flourish is a quiet visual cue drawn on the card itself, on top of
// the keyword chip: flying hovers, reach leans up, deathtouch has a dark
// edge, first and double strike carry blades, lifelink glows red. The
// chips stay: a flourish is atmosphere and the chip is the fact.
//
// It reads the card's EFFECTIVE keywords (`CardView.abilities`, the
// engine's list after Layer 6, the same list KeywordBadgeRow badges), not
// the printed ones, so a keyword a lord grants shows and one a Humility
// takes away stops showing, with no extra bookkeeping. The client owns no
// keyword grammar beyond these names.

export interface Flourishes {
  // Hovers: a small lift, a soft shadow beneath and (with motion) the
  // tiniest slow sway.
  flying: boolean;
  // Leans up. Flying wins where a card has both: a creature that already
  // hovers does not also stretch.
  reach: boolean;
  // A thin dark edge.
  deathtouch: boolean;
  // 0, 1 (first strike) or 2 (double strike). Double strike implies
  // first strike (CR 702.4b), so a card with both shows two blades, not
  // three.
  blades: 0 | 1 | 2;
  // A faint red glow.
  lifelink: boolean;
}

export const NO_FLOURISHES: Readonly<Flourishes> = Object.freeze({
  flying: false,
  reach: false,
  deathtouch: false,
  blades: 0,
  lifelink: false,
});

// flourishesFor derives the flourishes from an effective keyword list.
// Matching is case-insensitive and ignores a trailing parameter ("flying"
// is what the wire carries; nothing here depends on that).
export function flourishesFor(abilities: readonly string[] | null | undefined): Flourishes {
  if (!abilities || abilities.length === 0) return NO_FLOURISHES;
  const has = new Set(abilities.map((a) => a.toLowerCase().trim()));
  const flying = has.has("flying");
  const blades = has.has("double strike") ? 2 : has.has("first strike") ? 1 : 0;
  const out: Flourishes = {
    flying,
    reach: has.has("reach") && !flying,
    deathtouch: has.has("deathtouch"),
    blades,
    lifelink: has.has("lifelink"),
  };
  return out.flying || out.reach || out.deathtouch || out.blades > 0 || out.lifelink
    ? out
    : NO_FLOURISHES;
}

// swayDelay turns a card's id into a stable negative animation delay in
// milliseconds, so a row of fliers does not bob in step. A pure function
// of the id: no timer, no random, the same card is always out of phase by
// the same amount.
export function swayDelay(id: string, spanMs = 7000): string {
  let h = 0;
  for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) >>> 0;
  return `-${h % spanMs}ms`;
}
