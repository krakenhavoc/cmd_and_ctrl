import type { ActivatedAbilityView, AlternativeCostView, CardView } from "./protocol";

// phyrexianLife.ts — the arithmetic behind "pay N Phyrexian symbols
// with life" (CR 107.4c/f, #916), kept out of the modal so it can be
// tested without a component harness.
//
// A Phyrexian mana symbol — {U/P}, or one of CR 107.4's ten hybrid
// Phyrexian symbols {W/U/P} … {G/U/P} — can be paid with one mana of
// its colour, or with 2 life. Which the announcer does is part of
// ANNOUNCING (CR 601.2b for a cast, CR 602.2b for an activation), so
// it is a number collected before anything is paid and sent with the
// cast_spell / activate_ability payload as `phyrexian_life`.
//
// # The client parses no mana strings
//
// How many symbols a cost prints is a fact about the mana-cost
// syntax, and the syntax is the server's to read: #787 was a whole
// symbol family the parser had never heard of, and a client that
// re-derived the count would be a second parser to keep in step. So
// the ceiling arrives as `phyrexian_symbols` on the CardView, on the
// chosen alternative cost, or on the activated ability, and
// everything below is arithmetic on that number and the player's
// life total.

// PhyrexianLifePerSymbol is CR 107.4c's price for one symbol, hybrid
// Phyrexian included (CR 107.4f). The same constant the engine
// charges; the stepper's readout is a multiple of it.
export const PhyrexianLifePerSymbol = 2;

// phyrexianSymbolsForCast is the ceiling on a cast's claim: the count
// of the cost the caster is actually paying.
//
// An alternative cost REPLACES the mana cost (CR 601.2f), so claiming
// one replaces the ceiling with its own — Force of Will's free cast
// prints no Phyrexian symbol however many the printed cost has, and a
// hypothetical offer that printed one would allow it however few the
// printed cost has. Undefined `altCost` is the ordinary "pay the
// printed cost" case.
export function phyrexianSymbolsForCast(card: CardView, altCost: string | undefined): number {
  if (altCost !== undefined) {
    const offer = (card.alternative_costs ?? []).find(
      (o: AlternativeCostView) => o.key === altCost,
    );
    return offer?.phyrexian_symbols ?? 0;
  }
  return card.phyrexian_symbols ?? 0;
}

// phyrexianSymbolsForAbility is the same ceiling for a CR 602
// activated ability's mana component.
export function phyrexianSymbolsForAbility(ability: ActivatedAbilityView): number {
  return ability.phyrexian_symbols ?? 0;
}

// maxPhyrexianLife is how many symbols this announcer may actually
// claim: the symbols the cost prints, capped by CR 119.4 — a player
// may pay life only down to 0, so the payment may not exceed the life
// total.
//
// This is the SERVER's bound, deliberately, not a stricter one.
// Paying down to exactly 0 is a legal announcement the engine accepts
// (a state-based action then applies, which is the player's business,
// not the picker's), and a client that refused it would hide a legal
// play; a client that allowed more would collect a value the announce
// gate rejects. Either way the two must agree, so there is one rule
// and this is it.
//
// `life` may arrive undefined from a snapshot mid-update; that is
// read as "no life to spend" rather than as no limit.
export function maxPhyrexianLife(symbols: number, life: number | undefined): number {
  if (symbols <= 0) return 0;
  const affordable = Math.floor((life ?? 0) / PhyrexianLifePerSymbol);
  return Math.max(0, Math.min(symbols, affordable));
}

// clampPhyrexianLife holds a stepper's value inside [0, max] — used
// on every change, and again when a snapshot moves the life total
// under an open prompt (a Sulfuric Vortex trigger while the modal is
// up must not leave an unpayable claim sitting in the input).
export function clampPhyrexianLife(value: number, max: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.max(0, Math.min(Math.floor(value), max));
}

// phyrexianLifeCost is what a claim of `n` symbols costs in life.
export function phyrexianLifeCost(n: number): number {
  return Math.max(0, n) * PhyrexianLifePerSymbol;
}

// shouldAskPhyrexianLife reports whether the prompt is worth opening
// at all. There is nothing to ask when the cost prints no Phyrexian
// symbol, and nothing to ask when CR 119.4 leaves the player unable
// to buy even one — a modal whose only answer is 0 is a click, not a
// choice, so the announcement goes out claiming nothing.
export function shouldAskPhyrexianLife(symbols: number, life: number | undefined): boolean {
  return maxPhyrexianLife(symbols, life) > 0;
}
