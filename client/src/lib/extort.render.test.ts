// @vitest-environment jsdom
//
// #4544, the rendered half: an extort drain plays the Scales of Souls over
// the table and its `extort` sound; one, two and three drained opponents
// send one soul each; a run of drains queues (one scale at a time, the
// waiting ones merged); the three reduced-motion signals and the Short
// setting show a static scale and no flight; a click skips it; and life
// lost any other way plays nothing. The rules are pinned in extort.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: vi.fn() }));

import ExtortScales from "./components/board/ExtortScales.svelte";
import { play as playSound } from "./sounds";
import type { GameView, LogEvent } from "./protocol";
import { L } from "./labels";
import { resetSettings, updateSettings } from "./settings";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";
import { EFFECT_STILL_HOLD_MS } from "./tableEffects";
import { EXTORT_MS, darkPanAt, scaleBox } from "./extort";

// ---- the table ----------------------------------------------------------

const SEATS = ["a", "b", "c", "d"];

function viewWith(log: LogEvent[]): GameView {
  const v = dockTestView({ log } as Partial<GameView>);
  const base = v.seats[0];
  v.seats = SEATS.map((id, i) => ({ ...base, id, name: id, seat: i })) as GameView["seats"];
  return v;
}

let seq = 0;
const resolve = (batch: number, label = "Crypt Ghast — extort"): LogEvent => ({
  seq: ++seq,
  kind: "resolve",
  seat: 0,
  card_id: "ghast",
  label,
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
// An extort trigger that drained `victims` (seat indexes) for 1 each.
const extort = (batch: number, victims: number[]): LogEvent[] => [
  resolve(batch),
  ...victims.map((s) => life(batch, s, -1)),
  life(batch, 0, victims.length),
];

const boardSize = { w: 1200, h: 800 };

function rectEl(
  attr: string,
  value: string,
  r: { left: number; top: number; width: number; height: number },
): HTMLElement {
  const el = document.createElement("div");
  el.setAttribute(attr, value);
  el.getBoundingClientRect = () => new DOMRect(r.left, r.top, r.width, r.height);
  return el;
}

// The avatars: a at the bottom (the extort player), b to the left, c at the
// top, d to the right.
const AVATARS: Record<string, { left: number; top: number }> = {
  a: { left: 570, top: 700 },
  b: { left: 40, top: 370 },
  c: { left: 570, top: 40 },
  d: { left: 1100, top: 370 },
};

function makeBoard(): HTMLElement {
  const board = document.createElement("div");
  board.getBoundingClientRect = () => new DOMRect(0, 0, boardSize.w, boardSize.h);
  for (const id of SEATS) {
    board.appendChild(rectEl("data-seat-id", id, { ...AVATARS[id], width: 60, height: 60 }));
  }
  document.body.appendChild(board);
  return board;
}

function advance(ms: number) {
  flushSync();
  vi.advanceTimersByTime(ms);
  flushSync();
}

const layer = () => document.querySelector<HTMLElement>("[data-extort-scales]")!;
const scene = () => layer().querySelector<HTMLElement>("[data-extort]");
const souls = () => [...layer().querySelectorAll<HTMLElement>("[data-extort-soul]")];
const sparks = () => [...layer().querySelectorAll<HTMLElement>("[data-extort-spark]")];
const caption = () => layer().querySelector("[data-extort-caption]")?.textContent;
const extortSounds = () => vi.mocked(playSound).mock.calls.filter(([n]) => n === "extort").length;

// Mount on a primed (empty) frame, then deliver `log` as a new frame.
function deliver(log: LogEvent[]) {
  const board = makeBoard();
  const r = render(
    ExtortScales as never,
    {
      view: viewWith([]),
      boardEl: board,
      viewerID: "a",
    } as never,
  );
  advance(50);
  const set = (next: LogEvent[]) => {
    r.setProps({ view: viewWith(next) } as never);
    flushSync();
  };
  set(log);
  return { r, set };
}

beforeEach(() => {
  vi.useFakeTimers({
    toFake: ["Date", "setTimeout", "clearTimeout", "setInterval", "clearInterval"],
  });
  vi.setSystemTime(1_000_000);
  seq = 0;
  resetSettings();
  updateSettings("animations", "enabled", true);
  updateSettings("animations", "effects", "full");
  updateSettings("animations", "speed", 1);
  updateSettings("accessibility", "reduceMotion", false);
  vi.mocked(playSound).mockClear();
});

afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("the Scales of Souls", () => {
  it("draws nothing on its first frame, and nothing for an old entry", () => {
    const board = makeBoard();
    render(
      ExtortScales as never,
      {
        view: viewWith(extort(5, [1])),
        boardEl: board,
      } as never,
    );
    flushSync();
    expect(scene()).toBeNull();
    expect(playSound).not.toHaveBeenCalled();
  });

  it("is non-blocking and hidden from assistive tech", () => {
    deliver(extort(5, [1]));
    expect(layer().getAttribute("aria-hidden")).toBe("true");
  });

  it("an extort drain plays the scale and the extort sound, once", () => {
    deliver(extort(5, [1]));
    expect(scene()).not.toBeNull();
    expect(scene()!.getAttribute("data-extort-motion")).toBe("full");
    expect(extortSounds()).toBe(1);
    expect(vi.mocked(playSound)).toHaveBeenCalledWith("extort");
    expect(layer().querySelector("svg")).not.toBeNull();
    expect(caption()).toBe(L.effectExtort);
  });

  it("is a compact scale in the middle of the board", () => {
    deliver(extort(5, [1]));
    const box = scaleBox(boardSize);
    const el = layer().querySelector<HTMLElement>(".x-box")!;
    expect(el.style.left).toBe(`${box.x}px`);
    expect(el.style.top).toBe(`${box.y}px`);
    expect(el.style.width).toBe(`${box.width}px`);
    // Under a quarter of the board's width: the cards around it stay readable.
    expect(box.width).toBeLessThan(boardSize.w / 4);
  });

  it("one drained opponent sends one soul, from their avatar into the dark pan, and one spark home", () => {
    deliver(extort(5, [1]));
    expect(souls()).toHaveLength(1);
    expect(sparks()).toHaveLength(1);
    const dark = darkPanAt(scaleBox(boardSize));
    const [soul] = souls();
    // The soul sits on the pan and starts back at b's avatar centre.
    expect(soul.style.left).toBe(`${dark.x}px`);
    expect(soul.style.top).toBe(`${dark.y}px`);
    expect(soul.style.getPropertyValue("--dx")).toBe(`${Math.round(70 - dark.x)}px`);
    expect(soul.style.getPropertyValue("--dy")).toBe(`${Math.round(400 - dark.y)}px`);
    // The light goes to a's avatar: below the scale, so it flies down.
    const [spark] = sparks();
    expect(parseFloat(spark.style.getPropertyValue("--dy"))).toBeGreaterThan(0);
    expect(extortSounds()).toBe(1);
  });

  it("two drained opponents send two souls, three send three; the sound is still one", () => {
    deliver(extort(5, [1, 3]));
    expect(souls()).toHaveLength(2);
    expect(sparks()).toHaveLength(1);
    expect(extortSounds()).toBe(1);
    cleanup();
    vi.mocked(playSound).mockClear();
    deliver(extort(6, [1, 2, 3]));
    expect(souls()).toHaveLength(3);
    expect(sparks()).toHaveLength(1);
    expect(extortSounds()).toBe(1);
  });

  it("the souls leave a little apart, and the scale ends by itself", () => {
    deliver(extort(5, [1, 2, 3]));
    const delays = souls().map((s) => parseFloat(s.style.getPropertyValue("--delay")));
    expect(delays[0]).toBe(0);
    expect(delays[1]).toBeGreaterThan(delays[0]);
    expect(delays[2]).toBeGreaterThan(delays[1]);
    advance(EXTORT_MS - 1);
    expect(scene()).not.toBeNull();
    advance(1);
    expect(scene()).toBeNull();
  });

  it("scales its clock with animations.speed", () => {
    updateSettings("animations", "speed", 2);
    deliver(extort(5, [1]));
    expect(scene()!.style.getPropertyValue("--k")).toBe("2");
    advance(EXTORT_MS * 2 - 1);
    expect(scene()).not.toBeNull();
    advance(1);
    expect(scene()).toBeNull();
  });
});

describe("life lost any other way", () => {
  it("plays nothing for a drain with no extort trigger behind it", () => {
    deliver([
      resolve(5, "Gray Merchant of Asphodel — drain"),
      life(5, 1, -2),
      life(5, 0, 2),
      life(6, 2, -1),
    ]);
    expect(scene()).toBeNull();
    expect(extortSounds()).toBe(0);
  });

  it("plays nothing for combat damage, a paid life cost or a declined extort", () => {
    deliver([
      { seq: ++seq, kind: "damage", seat: 0, target_seat: 1, amount: 4, combat: true, text: "" },
      life(5, 0, -2),
      resolve(6),
    ]);
    expect(scene()).toBeNull();
    expect(extortSounds()).toBe(0);
  });
});

describe("a run of extort triggers", () => {
  it("is one scale at a time: the next waits, and everything that waits merges into one", () => {
    const { set } = deliver(extort(5, [1]));
    expect(extortSounds()).toBe(1);

    // Three more resolve while the first plays, over three frames.
    set([...extort(5, [1]), ...extort(6, [2])]);
    advance(300);
    set([...extort(5, [1]), ...extort(6, [2]), ...extort(7, [1, 2])]);
    advance(300);
    set([...extort(5, [1]), ...extort(6, [2]), ...extort(7, [1, 2]), ...extort(8, [3])]);
    // Still the first scale: one scene, one soul, one sound.
    expect(layer().querySelectorAll("[data-extort]")).toHaveLength(1);
    expect(scene()!.getAttribute("data-extort-count")).toBe("1");
    expect(souls()).toHaveLength(1);
    expect(extortSounds()).toBe(1);

    // It ends; the waiting scale (three triggers merged) takes over, once.
    advance(EXTORT_MS);
    expect(layer().querySelectorAll("[data-extort]")).toHaveLength(1);
    expect(scene()!.getAttribute("data-extort-count")).toBe("3");
    expect(caption()).toBe(L.effectExtortMany(3));
    expect(souls()).toHaveLength(3);
    expect(extortSounds()).toBe(2);

    // And then the table is clear: four triggers, two scales.
    advance(EXTORT_MS);
    expect(scene()).toBeNull();
    expect(extortSounds()).toBe(2);
  });

  it("triggers that resolve in one frame are one scale", () => {
    deliver([...extort(5, [1]), ...extort(6, [1]), ...extort(7, [2])]);
    expect(layer().querySelectorAll("[data-extort]")).toHaveLength(1);
    expect(scene()!.getAttribute("data-extort-count")).toBe("3");
    expect(souls()).toHaveLength(2);
    expect(extortSounds()).toBe(1);
  });

  it("does not play an old drain again when the log carries it on", () => {
    const { set } = deliver(extort(5, [1]));
    advance(EXTORT_MS);
    expect(scene()).toBeNull();
    set([...extort(5, [1]), { seq: ++seq, kind: "step", seat: 0, step: "end", text: "" }]);
    expect(scene()).toBeNull();
    expect(extortSounds()).toBe(1);
  });
});

describe("motion", () => {
  const expectStatic = () => {
    expect(scene()).not.toBeNull();
    expect(scene()!.getAttribute("data-extort-motion")).toBe("still");
    expect(souls()).toHaveLength(0);
    expect(sparks()).toHaveLength(0);
    expect(caption()).toBe(L.effectExtort);
    // The sound plays whatever the motion.
    expect(extortSounds()).toBe(1);
  };

  it("accessibility.reduceMotion shows a static scale: no flight, a fade, then gone", () => {
    updateSettings("accessibility", "reduceMotion", true);
    deliver(extort(5, [1, 2]));
    expectStatic();
    advance(EFFECT_STILL_HOLD_MS - 1);
    expect(scene()).not.toBeNull();
    advance(1);
    expect(scene()).toBeNull();
  });

  it("the OS's prefers-reduced-motion does the same", () => {
    vi.stubGlobal(
      "matchMedia",
      vi.fn(() => ({ matches: true, addEventListener() {}, removeEventListener() {} })),
    );
    deliver(extort(5, [1]));
    expectStatic();
  });

  it("animations off, and effects Off, do the same", () => {
    updateSettings("animations", "enabled", false);
    deliver(extort(5, [1]));
    expectStatic();
    cleanup();
    vi.mocked(playSound).mockClear();
    updateSettings("animations", "enabled", true);
    updateSettings("animations", "effects", "off");
    deliver(extort(6, [1]));
    expectStatic();
  });

  it("the Short setting shows a static scale for a shorter hold", () => {
    updateSettings("animations", "effects", "short");
    deliver(extort(5, [1]));
    expectStatic();
    advance(999);
    expect(scene()).not.toBeNull();
    advance(1);
    expect(scene()).toBeNull();
  });

  it("a click skips the scale and the one waiting", () => {
    const { set } = deliver(extort(5, [1]));
    set([...extort(5, [1]), ...extort(6, [2])]);
    expect(scene()).not.toBeNull();
    window.dispatchEvent(new Event("pointerdown"));
    flushSync();
    expect(scene()).toBeNull();
    advance(EXTORT_MS * 2);
    expect(scene()).toBeNull();
    expect(extortSounds()).toBe(1);
  });

  it("Escape skips it too", () => {
    deliver(extort(5, [1]));
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(scene()).toBeNull();
  });

  it("a prime key forgets what was playing, and cues nothing already on the wire", () => {
    const { r } = deliver(extort(5, [1]));
    expect(scene()).not.toBeNull();
    r.setProps({ beatsPrimeKey: "reconnected", view: viewWith(extort(5, [1])) } as never);
    flushSync();
    expect(scene()).toBeNull();
    expect(extortSounds()).toBe(1);
  });
});
