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
//
// ADR 0120 §3: a seat's board can also be drawn a second time, expanded
// over the table, so two Cards can carry the same instance ID. The open
// popover therefore records the SURFACE it was opened from, `table` or
// `expanded`, and a Card draws it only when both the ID and its own
// surface match. A Card learns its surface from a Svelte context that
// the expanded PlayerPanel sets (setPopoverSurface); everything else is
// on the table, the default.

import { getContext, setContext } from "svelte";
import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { emit as tutorialEmit } from "./tutorialBus";

/** Where a card is drawn: on the table, or in the expanded overlay. */
export type PopoverSurface = "table" | "expanded";

export interface AbilityPopoverOpen {
  /** The card whose popover is open. */
  cardID: string;
  /** The copy of the card it was opened from (ADR 0120 §3). */
  surface: PopoverSurface;
}

const SURFACE_KEY = Symbol("popover-surface");

/**
 * setPopoverSurface tells every component below the caller which
 * surface its cards are drawn on. Call it during component init.
 */
export function setPopoverSurface(surface: PopoverSurface): void {
  setContext(SURFACE_KEY, surface);
}

/**
 * popoverSurface is the surface set by an ancestor, or `table` when
 * none set one. Call it during component init.
 */
export function popoverSurface(): PopoverSurface {
  return getContext<PopoverSurface | undefined>(SURFACE_KEY) ?? "table";
}

export const abilityPopover: Writable<AbilityPopoverOpen | null> = guardedWritable(
  null,
  "abilityPopover",
);

/**
 * openAbilityPopover opens `cardID`'s popover on the copy of the card
 * drawn on `surface`. Every route into it — a left-click, a
 * right-click, a pip — tells the tutorial (`ability-menu-opened`,
 * ADR 0076 §2.5, ADR 0117 §6), so it is said here once.
 */
export function openAbilityPopover(cardID: string, surface: PopoverSurface = "table"): void {
  abilityPopover.set({ cardID, surface });
  tutorialEmit("ability-menu-opened");
}

export function closeAbilityPopover(): void {
  abilityPopover.set(null);
}

/**
 * popoverDrawnOn is whether the copy of `cardID` on `surface` draws the
 * open popover: both the card and the surface must match.
 */
export function popoverDrawnOn(
  open: AbilityPopoverOpen | null,
  cardID: string,
  surface: PopoverSurface,
): boolean {
  return open?.cardID === cardID && open.surface === surface;
}

/** Whether the popover is open on this card, on any surface, right now. */
export function abilityPopoverOpenFor(cardID: string): boolean {
  return get(abilityPopover)?.cardID === cardID;
}
