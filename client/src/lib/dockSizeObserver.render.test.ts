// @vitest-environment jsdom
//
// dockSizeObserver.render.test.ts — #4065's ResizeObserver finding. The
// dock reports its size to Game, whose onDockSize reads the size it last
// stored. Called from a tracked context, that read made the dock's size
// effect depend on Game's state, so every resize re-ran the effect from
// inside the observer's callback: a new observer, on the same element,
// whose first notification went undelivered ("ResizeObserver loop
// completed with undelivered notifications", after nearly every
// snapshot). The callback is untracked now, so one observer serves the
// dock's whole life.

import { afterEach, beforeEach, describe, expect, it } from "vitest";

import ActionDock from "./components/board/ActionDock.svelte";
import { _resetForTests as resetDock } from "./dock";
import { defaultSettings, settings } from "./settings";
import type { CardView, GameView, PlayerView } from "./protocol";
import { cleanup, flushSync, render } from "./test/render.svelte";
import { sizeSink } from "./test/sizeSink.svelte";

type Callback = (entries: unknown[], obs: unknown) => void;
const observers: { cb: Callback; targets: Element[]; live: boolean }[] = [];

class RecordingObserver {
  private rec: { cb: Callback; targets: Element[]; live: boolean };
  constructor(cb: Callback) {
    this.rec = { cb, targets: [], live: true };
    observers.push(this.rec);
  }
  observe(el: Element): void {
    this.rec.targets.push(el);
  }
  unobserve(): void {}
  disconnect(): void {
    this.rec.live = false;
  }
}

const saved: Record<string, unknown> = {};
beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  saved.ResizeObserver = g.ResizeObserver;
  g.ResizeObserver = RecordingObserver;
  observers.length = 0;
  settings.set(defaultSettings());
  resetDock();
});
afterEach(() => {
  cleanup();
  (globalThis as Record<string, unknown>).ResizeObserver = saved.ResizeObserver;
});

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});
const seat = (id: string, name: string, n: number): PlayerView =>
  ({
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
  }) as unknown as PlayerView;
const gameView = (): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("opp", "Opp", 1)],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      seq: 5,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "precombat_main",
      step: "precombat_main",
    },
    mulligans_open: false,
  }) as unknown as GameView;

describe("the dock's size report", () => {
  it("keeps one observer on the dock while its size changes", () => {
    const sink = sizeSink();
    const r = render(
      ActionDock as never,
      {
        view: gameView(),
        viewerHasPriority: true,
        viewerIsActive: true,
        autopassEnabled: false,
        onPassPriority: () => {},
        onPassTurn: () => {},
        onToggleAutopass: () => {},
        onSize: sink.report,
      } as never,
    );
    flushSync();
    const dock = r.container.querySelector<HTMLElement>("section.action-dock")!;
    // Two watch it: the sheet ceiling's and the size report's.
    const onDock = () => observers.filter((o) => o.live && o.targets.includes(dock));
    const watching = onDock();
    expect(watching.length).toBeGreaterThan(0);
    const before = observers.length;

    // The dock grows, as it does when a snapshot changes its text.
    let w = 348;
    let h = 143;
    Object.defineProperty(dock, "offsetWidth", { configurable: true, get: () => w });
    Object.defineProperty(dock, "offsetHeight", { configurable: true, get: () => h });
    for (const [nw, nh] of [
      [348, 143],
      [348, 165],
      [360, 198],
    ]) {
      w = nw;
      h = nh;
      for (const o of watching) o.cb([], o);
      flushSync();
      expect(sink.size).toEqual({ w: nw, h: nh });
    }

    // The same observers are still the ones watching it: no effect re-ran.
    expect(observers.length).toBe(before);
    expect(onDock()).toEqual(watching);
  });
});
