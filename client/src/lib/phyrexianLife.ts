import type { ActivatedAbilityView, AlternativeCostView, CardView } from "./protocol";

// phyrexianLife.ts — the arithmetic behind "pay N Phyrexian symbols
// with life" (CR 107.4f, #916), kept out of the modal so it can be
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

// PhyrexianLifePerSymbol is CR 107.4f's price for one symbol, hybrid
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

// phyrexianGrantedForCast is how many of those symbols are payable with
// life only because the viewer controls a grant (ADR 0131: K'rrik's "for
// each {B} in a cost, you may pay 2 life rather than pay that mana"),
// for the cost the caster is actually paying. A subset of the ceiling,
// never more than it.
export function phyrexianGrantedForCast(card: CardView, altCost: string | undefined): number {
  const symbols = phyrexianSymbolsForCast(card, altCost);
  if (altCost !== undefined) {
    const offer = (card.alternative_costs ?? []).find(
      (o: AlternativeCostView) => o.key === altCost,
    );
    return Math.min(symbols, offer?.phyrexian_granted ?? 0);
  }
  return Math.min(symbols, card.phyrexian_granted ?? 0);
}

// phyrexianGrantedForAbility is phyrexianGrantedForCast for a CR 602
// activated ability's mana component.
export function phyrexianGrantedForAbility(ability: ActivatedAbilityView): number {
  return Math.min(phyrexianSymbolsForAbility(ability), ability.phyrexian_granted ?? 0);
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
//
// ADR 0131 (owner decision 4): a printed Phyrexian symbol still always
// asks, as it did before. When EVERY symbol is a granted one (K'rrik's
// {B}), the prompt would be a click on almost every black spell, so it
// opens only if mana falls short, which the caller learns from the
// auto-tap preview and passes as `manaShort`, or if the player asked for
// it from the card menu (`asked`).
export function shouldAskPhyrexianLife(
  symbols: number,
  life: number | undefined,
  granted = 0,
  manaShort = false,
  asked = false,
): boolean {
  if (maxPhyrexianLife(symbols, life) <= 0) return false;
  if (symbols - granted > 0) return true;
  return manaShort || asked;
}

// onlyGrantedSymbols is the question the caller asks BEFORE it fetches
// the preview: is every life-payable symbol a granted one, so that
// whether to ask depends on the mana?
export function onlyGrantedSymbols(symbols: number, granted: number): boolean {
  return symbols > 0 && granted >= symbols;
}
