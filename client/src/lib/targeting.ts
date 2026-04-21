import { writable, type Writable } from "svelte/store";
import type { CardView } from "./protocol";

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
export type TargetingMode = "any" | "player" | "creature" | "stack_spell" | "card_in_graveyard";

// TargetingState is the active prompt. `card` is the spell being
// cast; `mode` is what the UI should accept as a click. The caller
// is responsible for calling cast_spell with the resolved target
// when a surface-level click matches.
export interface TargetingState {
  card: CardView;
  mode: TargetingMode;
}

export const targeting: Writable<TargetingState | null> = writable(null);

// begin enters a targeting prompt. Overwrites any existing prompt
// — the last cast wins. The caller has already verified the
// card's target_mode is non-empty.
export function begin(card: CardView, mode: TargetingMode): void {
  targeting.set({ card, mode });
}

// cancel clears the prompt without firing cast_spell. Wired to the
// Escape keybinding in Game.svelte.
export function cancel(): void {
  targeting.set(null);
}

// isTargetingPlayer / isTargetingCard / isTargetingStack are mode-
// membership helpers so consumers don't have to pattern-match.
export function isTargetingPlayer(mode: TargetingMode): boolean {
  return mode === "any" || mode === "player";
}

export function isTargetingCreature(mode: TargetingMode): boolean {
  return mode === "any" || mode === "creature";
}

export function isTargetingStack(mode: TargetingMode): boolean {
  return mode === "stack_spell";
}

export function isTargetingGraveyard(mode: TargetingMode): boolean {
  return mode === "card_in_graveyard";
}
