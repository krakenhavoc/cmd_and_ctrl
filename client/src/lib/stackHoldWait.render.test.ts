// @vitest-environment jsdom
//
// stackHoldWait.render.test.ts — #2853. While the ADR 0119 stack hold
// counts down to an automatic pass on an opponent's item, the action
// dock shows the countdown and a one-click "wait": it arms the hold
// toggle, so the pass does not go and the player can respond.

import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { get } from "svelte/store";
import { flushSync } from "svelte";

import ActionDock from "./components/board/ActionDock.svelte";
import { holdPriority, _resetForTests as resetHold } from "./holdPriority";
import { L } from "./labels";
import { defaultSettings, settings } from "./settings";
import { setStackHoldStatus, _resetForTests as resetStackHold } from "./stackHold";
import type { GameView, PlayerView, TurnView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

beforeEach(() => {
  settings.set(defaultSettings());
  resetHold();
  resetStackHold();
});
afterEach(() => {
  cleanup();
  resetStackHold();
});

const seats = [
  { id: "a", name: "Alice", seat: 0 },
  { id: "b", name: "Bob", seat: 1 },
] as unknown as PlayerView[];

const turn: TurnView = {
  seq: 5,
  number: 3,
  active_seat: 1,
  priority_holder: 0,
  phase: "precombat_main",
  step: "precombat_main",
};

function mount() {
  return render(
    ActionDock as never,
    {
      view: { turn, seats, mulligans_open: false } as unknown as GameView,
      viewerHasPriority: true,
      viewerIsActive: false,
      autopassEnabled: false,
      onPassPriority: () => {},
      onPassTurn: () => {},
      onToggleAutopass: () => {},
    } as never,
  ).container;
}

const waitButton = (c: HTMLElement) =>
  c.querySelector<HTMLButtonElement>(`button[aria-label="${L.waitToRespond}"]`);

describe("the stack hold's wait button", () => {
  it("is not shown without a countdown", () => {
    expect(waitButton(mount())).toBeNull();
  });

  it("shows with the countdown and arms the hold when clicked", () => {
    setStackHoldStatus({ passesAt: Date.now() + 1500 });
    const c = mount();
    flushSync();
    expect(c.querySelector('[role="timer"]')?.textContent).toMatch(/auto-pass in/);
    const b = waitButton(c);
    expect(b).not.toBeNull();
    b!.click();
    flushSync();
    expect(get(holdPriority)).toBe(true);
  });
});
