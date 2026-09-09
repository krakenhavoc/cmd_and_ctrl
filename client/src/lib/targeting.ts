import { writable, type Writable } from "svelte/store";
import type { ActivatedAbilityView, CardView, ModeOptionView, PendingChoiceView } from "./protocol";

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
  // S20 sub-PR 3: the announced X for an {X} spell, chosen in the X
  // prompt before targeting; rides the cast_spell payload.
  xValue?: number;
  // S21 sub-PR 2: set when the prompt collects targets for an
  // ACTIVATED ability rather than a cast. The confirm fires
  // activate_ability with these announce-time choices.
  ability?: { index: number; sacrificeIDs: string[] };
  // Human-readable clause for the banner ("target artifact or
  // enchantment"); the server's TargetSpec label.
  label?: string;
  // S20 sub-PR 4: the chosen mode indexes of a modal spell; ride the
  // cast_spell payload as `modes`.
  modes?: number[];
  // S21 sub-PR 5: instance IDs of the cards paid to an additional
  // cost ("discard a card"), collected before this prompt opened;
  // ride the cast_spell payload as `discard_ids`.
  discardIDs?: string[];
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
export function begin(
  card: CardView,
  mode: TargetingMode,
  xValue?: number,
  discardIDs?: string[],
): void {
  const lt = card.legal_targets;
  const legal = lt
    ? { players: new Set(lt.players ?? []), cards: new Set(lt.cards ?? []) }
    : undefined;
  targeting.set({ card, mode, legal, xValue, discardIDs, ...countOf(lt), picked: [] });
}

// countOf reads a clause's min / max off the wire; free-form cards
// (no legal set) are single-target.
function countOf(lt: { min?: number; max?: number } | undefined): { min: number; max: number } {
  if (!lt) return { min: 1, max: 1 };
  return { min: lt.min ?? 1, max: lt.max ?? 1 };
}

// isMultiPick reports whether the prompt accumulates picks rather
// than completing on the first click.
export function isMultiPick(t: TargetingState): boolean {
  return t.max !== 1;
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
  xValue?: number,
  discardIDs?: string[],
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
    xValue,
    discardIDs,
    modes,
    label: option.label,
    ...countOf(lt),
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

// modeOptionCastable reports whether an option can be chosen right
// now: untargeted options always can; targeted ones need at least
// one legal target.
export function modeOptionCastable(option: ModeOptionView): boolean {
  const lt = option.legal_targets;
  if (!lt) return true;
  return (lt.players?.length ?? 0) + (lt.cards?.length ?? 0) >= (lt.min ?? 1);
}

// hasXCost reports whether a card's printed cost includes {X} — the
// cue to open the X prompt before casting.
export function hasXCost(card: CardView): boolean {
  return (card.mana_cost ?? "").includes("{X}");
}

// beginForAbility enters a targeting prompt for an activated
// ability's target clause. `card` is the source permanent; the
// legal set comes from the ability, not the card.
export function beginForAbility(
  card: CardView,
  ability: ActivatedAbilityView,
  sacrificeIDs: string[],
): void {
  const lt = ability.legal_targets;
  targeting.set({
    card,
    mode: (ability.target_mode || "any") as TargetingMode,
    legal: lt ? { players: new Set(lt.players ?? []), cards: new Set(lt.cards ?? []) } : undefined,
    ability: { index: ability.index, sacrificeIDs },
    // Single-target by contract, not by omission. Unlike CardView,
    // ModeOptionView and pick_target — which all carry a
    // LegalTargetsView and read their count via countOf — S21 sub-PR
    // 2 gave ActivatedAbilityView its own inline
    // `{players?, cards?}`, with no min/max on the wire. So there is
    // no count to read here, and an ability clause behaves as
    // exactly one pick (isMultiPick is false, the first click
    // completes the prompt), which is what the ability menu expects.
    //
    // If a multi-target activated ability is ever wanted, the fix is
    // to make ActivatedAbilityView use LegalTargetsView and emit the
    // count server-side, then switch these two lines to
    // `...countOf(lt)` like the three siblings.
    min: 1,
    max: 1,
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
