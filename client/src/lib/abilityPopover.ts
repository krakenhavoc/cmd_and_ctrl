// abilityPopover — which card's light ability popover is open (ADR 0117
// §1). Keyed by instance ID, in the shape of manaSourcePicker.ts.
//
// The popover itself is drawn by Card (ManaAbilityMenu, anchored inside
// the card), but since ADR 0117 a LEFT-click can open it, and that
// click lands in PlayerPanel's router, several components above the
// Card. So the open card lives here: the router writes it, a right-click
// or a pip writes it, and the one Card with that instance ID draws the
// popover. A token group's drawn card stands in for whichever member
// the popover is open on (BattlefieldRow), which is how "Use this one"
// in the member list opens it at the group.
//
// One popover at a time: opening another card's replaces it.

import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { emit as tutorialEmit } from "./tutorialBus";

export interface AbilityPopoverOpen {
  /** The card whose popover is open. */
  cardID: string;
}

export const abilityPopover: Writable<AbilityPopoverOpen | null> = guardedWritable(
  null,
  "abilityPopover",
);

/**
 * openAbilityPopover opens `cardID`'s popover. Every route into it — a
 * left-click, a right-click, a pip — tells the tutorial
 * (`ability-menu-opened`, ADR 0076 §2.5, ADR 0117 §6), so it is said
 * here once.
 */
export function openAbilityPopover(cardID: string): void {
  abilityPopover.set({ cardID });
  tutorialEmit("ability-menu-opened");
}

export function closeAbilityPopover(): void {
  abilityPopover.set(null);
}

/** Whether the popover is open on this card right now. */
export function abilityPopoverOpenFor(cardID: string): boolean {
  return get(abilityPopover)?.cardID === cardID;
}
