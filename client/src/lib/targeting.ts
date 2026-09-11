import { writable, type Writable } from "svelte/store";
import type {
  ActivatedAbilityView,
  AlternativeCostView,
  CardView,
  LegalTargetsView,
  ModeOptionView,
  PendingChoiceView,
  TapCostView,
} from "./protocol";

// targeting.ts is the shared-store plumbing for the S14 "cast a
// catalog card, pick a target" flow. When a player clicks a hand
// card whose CardView declares a non-empty `target_mode`, the
// caller writes a TargetingState here; receiving surfaces (player
// headers, battlefield cards, stack items) watch the store and
// re-route their click handlers to resolve the target prompt.
// Store clears on Escape key or on cast completion.
//
// This is the minimum-viable two-click pick (hand card → click a
// target → cast fires). A richer picker (hover-highlight legal
// targets, validate predicate server-side) is S20 smart-cast
// territory.

// TargetingMode mirrors effects.Spec.TargetMode on the server.
// Values:
//   "any"         — player, creature, planeswalker, or battle
//   "player"      — seated player only
//   "creature"    — battlefield creature only
//   "stack_spell" — a spell currently on the stack
//   "card_in_graveyard" — a card in any graveyard (Regrowth, Eternal
//                         Witness)
export type TargetingMode =
  | "any"
  | "player"
  | "creature"
  | "permanent"
  | "stack_spell"
  | "card_in_graveyard";

// CastChoices bundles every announce-time decision collected before
// targeting opens, so a cast rides one object instead of a growing
// tail of positional parameters (the debt ADR 0021 flagged when
// xValue, discardIDs and sacrificeIDs became three of them; S22's
// alternative cost was the fourth and paid it off).
//
// Every field is optional and every field maps to exactly one
// cast_spell param, so applyCastChoices is the single place that
// knows the wire names.
export interface CastChoices {
  // S20 sub-PR 3: the announced X for an {X} spell.
  xValue?: number;
  // S21 sub-PR 5: instance IDs paid to a "discard a card" additional
  // cost.
  discardIDs?: string[];
  // S21 sub-PR 6: the permanent paid to a "sacrifice a creature"
  // additional cost.
  sacrificeIDs?: string[];
  // S22: the key of the alternative cost being paid INSTEAD of the
  // mana cost ("overload", "evoke", "cleave"). Undefined is the
  // ordinary "pay the printed cost" case.
  altCost?: string;
  // S22: the untapped permanents tapped to help pay — convoke and
  // waterbend. Undefined and empty are the same thing to the server;
  // tapping nothing is always legal.
  tapIDs?: string[];
  // ADR 0034: which printed face of a modal DFC is being cast or
  // played. Undefined and 0 are both "the front face", which is
  // every single-faced card. The face is chosen FIRST — before the
  // alternative cost, the modes, X and the targets — because it
  // decides what the card even is: Sea Gate Restoration and Sea
  // Gate, Reborn have different types, different costs and, via
  // the composite catalog key, different rules.
  face?: number;
  // S29: the zone the cast comes out of. Undefined is the hand,
  // which is every cast the Board's own surfaces fire.
  //
  // It rides CastChoices rather than being a parameter of its own
  // because a graveyard cast has to walk the SAME prompt chain as a
  // hand cast — flashback picks a cost, escape pays an additional
  // cost, Cackling Counterpart picks a target — and threading a
  // second positional argument through eight `after*` seams is
  // exactly the debt this object was created to pay off. The two
  // pre-S29 senders of `from_zone` (the command zone and the exile
  // impulse button) fire bare payloads with no prompts at all, which
  // is why they never needed it.
  fromZone?: CastSourceZone;
}

// CastSourceZone is the `from_zone` vocabulary the server's
// castZoneFromWire accepts. "hand" is never sent — it is the server
// default and omitting it keeps every pre-S29 client's payload
// byte-identical.
export type CastSourceZone = "command" | "exile" | "graveyard";

// applyCastChoices writes a CastChoices onto a cast_spell payload.
// Undefined fields are omitted rather than sent as null — the server
// distinguishes "no additional cost paid" from "paid nothing".
export function applyCastChoices(
  params: Record<string, unknown>,
  choices: CastChoices | undefined,
): void {
  if (!choices) return;
  if (choices.xValue !== undefined) params.x_value = choices.xValue;
  if (choices.discardIDs !== undefined) params.discard_ids = choices.discardIDs;
  if (choices.sacrificeIDs !== undefined) params.sacrifice_ids = choices.sacrificeIDs;
  if (choices.altCost !== undefined) params.alternative_cost = choices.altCost;
  if (choices.tapIDs !== undefined && choices.tapIDs.length > 0) params.tap_ids = choices.tapIDs;
  // Face 0 is omitted rather than sent explicitly: it is the server
  // default, and `omitempty` on the Go side means an explicit zero
  // and an absent field are the same byte on the wire anyway.
  if (choices.face !== undefined && choices.face > 0) params.face = choices.face;
  // S29: omitted for a hand cast, for the same reason face 0 is.
  if (choices.fromZone !== undefined) params.from_zone = choices.fromZone;
}

// TargetingState is the active prompt. `card` is the spell being
// cast; `mode` is what the UI should accept as a click. The caller
// is responsible for calling cast_spell with the resolved target
// when a surface-level click matches.
export interface TargetingState {
  card: CardView;
  mode: TargetingMode;
  // S20: the server-computed legal set for this card (from
  // CardView.legal_targets). Undefined for free-form cards, where
  // legality falls back to the mode heuristics below.
  legal?: { players: Set<string>; cards: Set<string> };
  // S20 sub-PR 2: set when the prompt answers a pick_target pending
  // choice (a triggered ability choosing its target) rather than a
  // cast. The click resolves the choice instead of firing
  // cast_spell, and the prompt can't be cancelled — the trigger
  // needs a target.
  choiceID?: string;
  // S22: the announce-time payments and choices collected before
  // this prompt opened — X, additional-cost picks, the alternative
  // cost being paid. They ride the cast_spell payload verbatim via
  // applyCastChoices. Absent for pick_target and ability prompts,
  // which aren't casts.
  choices?: CastChoices;
  // S21 sub-PR 2: set when the prompt collects targets for an
  // ACTIVATED ability rather than a cast. The confirm fires
  // activate_ability with these announce-time choices.
  ability?: { index: number; sacrificeIDs: string[]; crewIDs: string[] };
  // Human-readable clause for the banner ("target artifact or
  // enchantment"); the server's TargetSpec label.
  label?: string;
  // S20 sub-PR 4: the chosen mode indexes of a modal spell; ride the
  // cast_spell payload as `modes`.
  modes?: number[];
  // S20 sub-PR 5: the clause's target count. At max 1 the first
  // click completes the prompt; otherwise clicks toggle into
  // `picked` (in click order — positional clauses read it) and the
  // banner's Done fires once at least `min` are picked. max 0 =
  // unbounded.
  min: number;
  max: number;
  picked: TargetRef[];
}

export interface TargetRef {
  kind: "player" | "card";
  id: string;
}

export const targeting: Writable<TargetingState | null> = writable(null);

// begin enters a targeting prompt. Overwrites any existing prompt
// — the last cast wins. The caller has already verified the
// card's target_mode is non-empty.
//
// `alt` is the alternative cost being paid, when one is (S22): its
// clause replaces the card's, because the spell's targets are
// whatever the cost it was cast for says they are. Wash Away hard-cast
// can only hit a spell that wasn't cast from its owner's hand;
// cleaved it can hit any spell, and the legal set differs
// accordingly.
export function begin(
  card: CardView,
  mode: TargetingMode,
  choices?: CastChoices,
  alt?: AlternativeCostView,
): void {
  const lt = alt ? alt.legal_targets : card.legal_targets;
  const legal = lt
    ? { players: new Set(lt.players ?? []), cards: new Set(lt.cards ?? []) }
    : undefined;
  targeting.set({
    card,
    mode,
    legal,
    choices,
    ...countOf(lt, choices),
    picked: [],
  });
}

// countOf reads a clause's min / max off the wire; free-form cards
// (no legal set) are single-target.
//
// S22: a clause counted by X ("Exile X target creatures you
// control") carries no usable min / max — the server can't know X
// when it builds the snapshot — so the count comes from the X the
// caster announced in the cost prompts instead. The server rejects
// any other count, so getting this wrong is a rejected cast rather
// than a wrong one.
function countOf(
  lt: LegalTargetsView | undefined,
  choices?: CastChoices,
): { min: number; max: number } {
  if (!lt) return { min: 1, max: 1 };
  if (lt.count_from_x) {
    const x = choices?.xValue ?? 0;
    return { min: x, max: x };
  }
  return { min: lt.min ?? 1, max: lt.max ?? 1 };
}

// isMultiPick reports whether the prompt accumulates picks rather
// than completing on the first click — which is also what decides
// whether the banner grows a Done button.
//
// `min < 1` is the "up to one target" case (Teferi, Time Raveler's
// −3, The Wandering Emperor's −2). Exactly one pick is allowed, but
// zero is a legal answer, so the prompt cannot complete on the first
// click: the player needs a way to say "none". Done is that way, and
// canConfirm lets it fire at zero.
export function isMultiPick(t: TargetingState): boolean {
  return t.max !== 1 || t.min < 1;
}

// isPicked reports whether a target is already in the pick list.
export function isPicked(t: TargetingState, id: string): boolean {
  return t.picked.some((p) => p.id === id);
}

// togglePick adds a target to (or removes it from) a multi-pick
// prompt. Refuses a pick beyond max. Returns the new state.
export function togglePick(t: TargetingState, ref: TargetRef): TargetingState {
  if (isPicked(t, ref.id)) {
    return { ...t, picked: t.picked.filter((p) => p.id !== ref.id) };
  }
  if (t.max > 0 && t.picked.length >= t.max) return t;
  return { ...t, picked: [...t.picked, ref] };
}

// canConfirm reports whether Done may fire: at least min picked.
export function canConfirm(t: TargetingState): boolean {
  return t.picked.length >= t.min;
}

// The banner (mounted by Game.svelte) and the cast flow (Board)
// are separate components; Board registers the handler that turns
// the pick list into a cast_spell / resolve_choice, and the banner's
// Done calls confirm().
let confirmHandler: (() => void) | null = null;
export function setConfirmHandler(fn: (() => void) | null): void {
  confirmHandler = fn;
}
export function confirm(): void {
  confirmHandler?.();
}

// beginForMode enters a targeting prompt for the targeted option of
// a modal spell: the legal set and banner clause come from the
// option, not the card. `modes` is the full chosen set (the targeted
// option plus any untargeted ones) and rides the cast.
export function beginForMode(
  card: CardView,
  option: ModeOptionView,
  modes: number[],
  choices?: CastChoices,
): void {
  const mode = (option.target_mode || "any") as TargetingMode;
  const lt = option.legal_targets;
  const legal = lt
    ? { players: new Set(lt.players ?? []), cards: new Set(lt.cards ?? []) }
    : undefined;
  targeting.set({
    card,
    mode,
    legal,
    choices,
    modes,
    label: option.label,
    ...countOf(lt, choices),
    picked: [],
  });
}

// isModal reports whether a card needs the mode picker before it
// can be cast.
export function isModal(card: CardView): boolean {
  return (card.modes?.options?.length ?? 0) > 0;
}

// discardCostOf returns how many cards a card demands as an
// additional cost to cast, or 0 for the vast majority that demand
// none.
export function discardCostOf(card: CardView): number {
  return card.additional_cost?.discard_cards ?? 0;
}

// sacrificeCostOptions returns the permanents that may pay a
// "sacrifice a creature" additional cost, or undefined when the card
// charges no such cost. An empty array means the cost is unpayable.
export function sacrificeCostOptions(card: CardView): string[] | undefined {
  const opts = card.additional_cost?.sacrifice_options;
  if (!opts) return undefined;
  return opts.cards ?? [];
}

// tapCostOf returns a card's convoke / waterbend clause, or
// undefined for the vast majority of cards that offer none (S22).
export function tapCostOf(card: CardView): TapCostView | undefined {
  return card.tap_cost;
}

// tapCostLimit is how many permanents the picker may accept: the cap
// the server sent, or — for a waterbend {X}, whose size the server
// couldn't know when it built the snapshot — the X the caster just
// announced.
export function tapCostLimit(tc: TapCostView, xValue: number | undefined): number {
  if (tc.max && tc.max > 0) return tc.max;
  return xValue ?? 0;
}

// alternativeCostsOf returns the "cast this for its overload / evoke
// / cleave cost instead" offers on a card, or an empty list for the
// vast majority that have none (S22).
export function alternativeCostsOf(card: CardView): AlternativeCostView[] {
  return card.alternative_costs ?? [];
}

// alternativeCostByKey finds the offer a cast is paying. Undefined
// for the ordinary case — `key` undefined means "paying the printed
// mana cost", which is not an offer and has no view.
export function alternativeCostByKey(
  card: CardView,
  key: string | undefined,
): AlternativeCostView | undefined {
  if (key === undefined) return undefined;
  return alternativeCostsOf(card).find((a) => a.key === key);
}

// modeOptionCastable reports whether an option can be chosen right
// now: untargeted options always can; targeted ones need at least
// one legal target.
export function modeOptionCastable(option: ModeOptionView): boolean {
  const lt = option.legal_targets;
  if (!lt) return true;
  return (lt.players?.length ?? 0) + (lt.cards?.length ?? 0) >= (lt.min ?? 1);
}

// hasXCost reports whether a cast needs the X prompt.
//
// Usually that is an {X} in the printed cost. S22 adds a second
// source: a waterbend {X} cost is a cost of its own, layered on top
// of the card's, so Waterbender's Restoration prints {U}{U} and
// still has an X to announce — and that X is also the number of
// creatures its clause targets.
export function hasXCost(card: CardView): boolean {
  if (card.tap_cost?.demands_x) return true;
  return (card.mana_cost ?? "").includes("{X}");
}

// beginForAbility enters a targeting prompt for an activated
// ability's target clause. `card` is the source permanent; the
// legal set comes from the ability, not the card.
export function beginForAbility(
  card: CardView,
  ability: ActivatedAbilityView,
  sacrificeIDs: string[],
  crewIDs: string[] = [],
): void {
  const lt = ability.legal_targets;
  targeting.set({
    card,
    mode: (ability.target_mode || "any") as TargetingMode,
    legal: lt ? { players: new Set(lt.players ?? []), cards: new Set(lt.cards ?? []) } : undefined,
    ability: { index: ability.index, sacrificeIDs, crewIDs },
    // The count comes off the wire like every other clause's. This
    // used to be a hard-coded 1 / 1 with a note explaining that
    // ActivatedAbilityView carried an inline `{players?, cards?}`
    // with no min / max, and naming the fix: give it a
    // LegalTargetsView and emit the count server-side. #334 needed
    // exactly that — Teferi's "up to one target" is Min 0 — so the
    // note is now the code.
    ...countOf(lt),
    picked: [],
  });
}

// isLegalCardTarget / isLegalPlayerTarget answer "can I click this
// right now?" for a live prompt. With a server legal set (S20
// structured targeting) it's membership; without one (free-form
// S13.1 cards) it's the mode heuristic and the caller's zone
// routing.
export function isLegalCardTarget(t: TargetingState, instanceID: string): boolean {
  if (t.legal) return t.legal.cards.has(instanceID);
  return isTargetingCreature(t.mode) || isTargetingStack(t.mode) || isTargetingGraveyard(t.mode);
}

export function isLegalPlayerTarget(t: TargetingState, playerID: string): boolean {
  if (t.legal) return t.legal.players.has(playerID);
  return isTargetingPlayer(t.mode);
}

// legalTargetCount is the banner's "N legal targets" figure; -1 when
// the prompt is free-form.
export function legalTargetCount(t: TargetingState): number {
  if (!t.legal) return -1;
  return t.legal.players.size + t.legal.cards.size;
}

// beginChoice enters a targeting prompt for a pick_target pending
// choice. `card` is the trigger's source (for the banner); the
// legal set comes from the choice itself.
export function beginChoice(choice: PendingChoiceView, card: CardView): void {
  const pt = choice.pick_target ?? {};
  targeting.set({
    card,
    mode: "any",
    legal: { players: new Set(pt.players ?? []), cards: new Set(pt.cards ?? []) },
    choiceID: choice.id,
    label: choice.reason,
    ...countOf(pt),
    picked: [],
  });
}

// cancel clears the prompt without firing cast_spell. Wired to the
// Escape keybinding in Game.svelte. A pick_target prompt is not
// cancellable — the server is waiting for a target — so cancel is a
// no-op for it (the banner hides its Cancel button too).
export function cancel(): void {
  let current: TargetingState | null = null;
  targeting.update((t) => {
    current = t;
    return t;
  });
  if (current && (current as TargetingState).choiceID) return;
  targeting.set(null);
}

// isTargetingPlayer / isTargetingCard / isTargetingStack are mode-
// membership helpers so consumers don't have to pattern-match.
export function isTargetingPlayer(mode: TargetingMode): boolean {
  return mode === "any" || mode === "player";
}

export function isTargetingCreature(mode: TargetingMode): boolean {
  return mode === "any" || mode === "creature" || mode === "permanent";
}

export function isTargetingStack(mode: TargetingMode): boolean {
  return mode === "stack_spell";
}

export function isTargetingGraveyard(mode: TargetingMode): boolean {
  return mode === "card_in_graveyard";
}
