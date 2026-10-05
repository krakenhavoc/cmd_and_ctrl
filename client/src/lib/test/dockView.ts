// dockView.ts — helpers for render tests of the action dock's sheets
// (ADR 0111 Delivery PR 6). A picker that is a sheet draws its body in
// the dock's sheet panel and its confirm / cancel in the dock's action
// bar, inside one non-modal dialog; these find each part.

import { flushSync } from "svelte";

import { L } from "../labels";
import type { GameView } from "../protocol";

// dockTestView is the smallest GameView the dock's header renders.
export function dockTestView(over: Partial<GameView> = {}): GameView {
  const zone = (kind: string, owner?: string) => ({ kind, owner, count: 0, cards: [] });
  const seat = (id: string, name: string, n: number) => ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  });
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("opp", "Opp", 1)],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      seq: 1,
      number: 1,
      active_seat: 0,
      priority_holder: 0,
      phase: "precombat_main",
      step: "precombat_main",
    },
    mulligans_open: false,
    ...over,
  } as unknown as GameView;
}

// The dock region ("actions").
export function dockRegion(root: ParentNode = document): HTMLElement | null {
  return root.querySelector<HTMLElement>(`section.action-dock[aria-label="${L.actions}"]`);
}

// The open request's dialog in the dock.
export function dockDialog(root: ParentNode = document): HTMLElement | null {
  return dockRegion(root)?.querySelector<HTMLElement>('.dock-request[role="dialog"]') ?? null;
}

// The open sheet panel, inside the dialog (null if none, or minimised).
export function sheetPanel(root: ParentNode = document): HTMLElement | null {
  const el = dockDialog(root)?.querySelector<HTMLElement>(".dock-sheet") ?? null;
  return el && !el.hidden ? el : null;
}

// The action bar's primary while a request holds the bar.
export function barPrimary(root: ParentNode = document): HTMLButtonElement | null {
  return dockDialog(root)?.querySelector<HTMLButtonElement>(".dock-bar .request-primary") ?? null;
}

// The action bar's secondaries while a request holds the bar.
export function barSecondaries(root: ParentNode = document): HTMLButtonElement[] {
  return [
    ...(dockDialog(root)?.querySelectorAll<HTMLButtonElement>(".dock-bar .dock-btn.secondary") ??
      []),
  ];
}

// The text a button is named by: its content without aria-hidden caps.
export function nameOf(el: Element): string {
  const label = el.getAttribute("aria-label");
  if (label) return label;
  const copy = el.cloneNode(true) as Element;
  copy.querySelectorAll('[aria-hidden="true"]').forEach((n) => n.remove());
  return (copy.textContent ?? "").replace(/\s+/g, " ").trim();
}

// barButton finds an action-bar button by its accessible name.
export function barButton(name: string, root: ParentNode = document): HTMLButtonElement | null {
  const bar = dockDialog(root)?.querySelector(".dock-bar");
  if (!bar) return null;
  return (
    [...bar.querySelectorAll<HTMLButtonElement>("button")].find((b) => nameOf(b) === name) ?? null
  );
}

// pressKey dispatches a window keydown, as a key pressed with focus on
// the page body would arrive, and flushes.
export function pressKey(key: string, target: EventTarget = document.body): KeyboardEvent {
  const ev = new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true });
  target.dispatchEvent(ev);
  flushSync();
  return ev;
}
