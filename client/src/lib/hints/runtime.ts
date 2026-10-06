// runtime.ts — the one hint engine, and what the components share
// (ADR 0125 §3.5, §3.6).
//
// HintLayer.svelte (mounted once, in App.svelte) polls the engine and
// publishes the tip on screen here. HintCard draws it; HintSlot draws it
// instead when it belongs inside the Settings dialog. The shortcut layer
// calls `focusTip` for `i`, and the Help menu (ADR 0125 PR 5) calls
// `replayTips` and `showAllTipsAgain`.

import type { Readable } from "svelte/store";
import { guardedWritable } from "../guardedStore";
import { L } from "../labels";
import { settings } from "../settings";
import type { Hint, HintID, HintPlace } from "./hint";
import type { Placement, Rect } from "./place";
import { createHintEngine, placesFor, type EngineOptions, type HintEngine } from "./queue";
import { withSeen, withoutSeen } from "./seen";

/** The tip on screen, as the card draws it. */
export interface ActiveTip {
  hint: Hint;
  title: string;
  body: string;
  /** The anchor's rect, for the ring. */
  ring: Rect;
  /** Where the card goes; never "wait" (a hint that cannot be placed is not drawn). */
  placement: Exclude<Placement, { kind: "wait" }>;
  /** The phone strip's distance from the bottom edge (the dock bar at the table). */
  stripBottom: number;
  reduceMotion: boolean;
}

const tipStore = guardedWritable<ActiveTip | null>(null, "activeTip");

/** activeTip is the tip on screen, or null. */
export const activeTip: Readable<ActiveTip | null> = { subscribe: tipStore.subscribe };

/** setActiveTip is HintLayer's: what is on screen now. */
export function setActiveTip(tip: ActiveTip | null): void {
  tipStore.set(tip);
}

let engine: HintEngine = createHintEngine();

/** hintEngine is the one engine HintLayer polls. */
export function hintEngine(): HintEngine {
  return engine;
}

/** The id the tip's body carries, for the anchor's aria-describedby. */
export const TIP_BODY_ID = "hint-tip-body";

/** markHintSeen records a hint as dismissed at its current version. */
export function markHintSeen(h: Pick<Hint, "id" | "version">): void {
  settings.update((prev) => {
    const seen = withSeen(prev.help.seen, h.id, h.version);
    return seen === prev.help.seen ? prev : { ...prev, help: { ...prev.help, seen } };
  });
}

/**
 * markHintsTaught marks hints seen by id, each at its current version:
 * a tutorial step that completes has taught them (ADR 0125 §5.3). An id
 * no live hint has is skipped; tutorialSteps.test.ts keeps the script's
 * ids live.
 */
export function markHintsTaught(ids: readonly HintID[], hints: readonly Hint[]): void {
  for (const id of ids) {
    const h = hints.find((x) => x.id === id);
    if (h) markHintSeen(h);
  }
}

// ---- focus ----

let returnFocus: Element | null = null;

/** rememberReturnFocus records where focus goes back to when the tip closes. */
export function rememberReturnFocus(el: Element | null): void {
  returnFocus = el;
}

function tipElement(): HTMLElement | null {
  if (typeof document === "undefined") return null;
  return document.querySelector<HTMLElement>(`[aria-label="${L.tip}"]`);
}

/**
 * focusTip is Go to the tip (`i`, ADR 0125 §3.6): focus moves to the
 * card, and Escape there puts it back. With no tip showing it does
 * nothing and says nothing.
 */
export function focusTip(): void {
  const el = tipElement();
  if (!el) return;
  const from = document.activeElement;
  if (from && !el.contains(from)) returnFocus = from;
  el.focus();
}

/**
 * dismissTip closes the tip on screen: "Got it" and Escape mark it seen;
 * "Hide tips" does too, and turns tips off. Focus inside the card goes
 * back to where it came from.
 */
export function dismissTip(how: "got-it" | "hide", now: number = Date.now()): void {
  const card = tipElement();
  const hadFocus =
    !!card && typeof document !== "undefined" && card.contains(document.activeElement);
  const h = engine.dismiss(now);
  tipStore.set(null);
  if (h) markHintSeen(h);
  if (how === "hide") {
    settings.update((prev) => ({ ...prev, help: { ...prev.help, tipsOff: true } }));
  }
  if (hadFocus && returnFocus instanceof HTMLElement && returnFocus.isConnected) {
    returnFocus.focus();
  }
  returnFocus = null;
}

/**
 * replayableTips is what a replay of `place` shows: the place's own
 * hints and, on a site page, the "site" hints after them, as a visit
 * offers them (queue.ts placesFor). Empty for a place with none.
 */
export function replayableTips(place: HintPlace, hints: readonly Hint[]): Hint[] {
  const places = placesFor(place);
  return hints.filter((h) => places.includes(h.place));
}

/**
 * replayTips shows a place's hints again, one after another, even with
 * tips off: the Help menu's "Tips for this page" and the ⋯ menu's "Tips
 * for the table" (ADR 0125 §6). It forgets them first. On a site page
 * that includes the "site" hints, so Home, which has none of its own,
 * still shows the one about Help.
 */
export function replayTips(place: HintPlace, hints: readonly Hint[]): void {
  const ids = replayableTips(place, hints).map((h) => h.id);
  settings.update((prev) => ({
    ...prev,
    help: { ...prev.help, seen: withoutSeen(prev.help.seen, ids) },
  }));
  engine.replay(place, ids);
}

/** showAllTipsAgain forgets every dismissed hint and turns tips back on (ADR 0125 §6). */
export function showAllTipsAgain(): void {
  settings.update((prev) => ({ ...prev, help: { seen: {}, tipsOff: false } }));
}

/** _resetHintsForTests gives the next test a fresh engine and no tip. */
export function _resetHintsForTests(opts: EngineOptions = {}): void {
  engine = createHintEngine(opts);
  tipStore.set(null);
  returnFocus = null;
}
