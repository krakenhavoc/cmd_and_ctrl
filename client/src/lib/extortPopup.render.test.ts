// @vitest-environment jsdom
//
// #4544: the life popup over an avatar plays damage for a loss and heal for
// a gain, except for an extort drain, which the Scales of Souls cue sounds
// once (ExtortScales.svelte). Every other loss and gain keeps its sound.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: vi.fn() }));

import LifeCauseHost from "./test/LifeCauseHost.svelte";
import { play as playSound } from "./sounds";
import type { GameView, LogEvent, PlayerView } from "./protocol";
import { resetSettings, updateSettings } from "./settings";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";

let seq = 0;
const resolve = (batch: number): LogEvent => ({
  seq: ++seq,
  kind: "resolve",
  seat: 0,
  card_id: "ghast",
  label: "Crypt Ghast — extort",
  batch,
  text: "",
});
const life = (batch: number, seat: number, amount: number): LogEvent => ({
  seq: ++seq,
  kind: "life",
  seat,
  card_id: "ghast",
  amount,
  batch,
  text: "",
});

// The seat at `index` (1, the opponent, unless told otherwise) with `changes` in its history.
function frame(
  log: LogEvent[],
  changes: number[],
  index = 1,
): { view: GameView; seat: PlayerView } {
  const view = dockTestView({ log } as Partial<GameView>);
  const seat = {
    ...view.seats[index],
    life_history: changes.map((delta, i) => ({
      delta,
      new_total: 40 + delta,
      at: "2026-10-29T12:00:00Z",
      seq: i + 1,
    })),
  } as PlayerView;
  view.seats = view.seats.map((s, i) => (i === index ? seat : s)) as GameView["seats"];
  return { view, seat };
}

const sounds = () => vi.mocked(playSound).mock.calls.map(([n]) => n);

beforeEach(() => {
  seq = 0;
  resetSettings();
  updateSettings("animations", "enabled", true);
  vi.mocked(playSound).mockClear();
});
afterEach(cleanup);

// Mount on a primed frame, then deliver a loss as a new one.
function change(log: LogEvent[], delta: number, index = 1) {
  const r = render(LifeCauseHost as never, frame([], [], index) as never);
  flushSync();
  r.setProps(frame(log, [delta], index) as never);
  flushSync();
}

describe("the life popup's sound", () => {
  it("is damage for an ordinary loss", () => {
    change([life(5, 1, -3)], -3);
    expect(sounds()).toEqual(["damage"]);
  });

  it("is damage for life lost to a drain that is not extort", () => {
    change([life(5, 1, -1), life(5, 0, 1)], -1);
    expect(sounds()).toEqual(["damage"]);
  });

  it("is silent for an extort drain: the Scales' cue sounds it", () => {
    change([resolve(5), life(5, 1, -1), life(5, 0, 1)], -1);
    expect(sounds()).toEqual([]);
  });

  it("is silent for an extort drain's gain, too", () => {
    change([resolve(5), life(5, 1, -1), life(5, 0, 1)], 1, 0);
    expect(sounds()).toEqual([]);
  });

  it("keeps heal for a gain that is not extort's", () => {
    change([life(5, 0, 2)], 2, 0);
    expect(sounds()).toEqual(["heal"]);
  });
});
