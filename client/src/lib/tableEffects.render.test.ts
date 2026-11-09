// @vitest-environment jsdom
//
// ADR 0209 §1 and §3 (#4396, #4416), the rendered half: a battlefield
// exit, a counterspell, a planeswalker, a created token (#4441), counters
// put on a permanent (#4442), a commander cast from the command zone or
// going home (#4443), an Equipment attached or an Aura enchanting a
// permanent (#4444), cards milled or discarded into a graveyard (#4445),
// a card returned from a graveyard (#4446), a library searched (#4447), a
// double-faced card transforming or day and night changing (#4448) and
// a board wipe each draw their cue
// over the board and play their sound; a busy frame plays its cues shorter; reduced motion (all three signals) and the
// Off setting draw a badge and no motion; a click or Escape skips what is
// playing; and the first frame cues nothing. The rules are pinned in
// tableEffects.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: vi.fn() }));
vi.mock("./animations", async (importOriginal) => {
  const real = await importOriginal<typeof import("./animations")>();
  return { ...real, crumble: vi.fn(() => Promise.resolve()) };
});
vi.mock("./motion", async (importOriginal) => {
  const real = await importOriginal<typeof import("./motion")>();
  return {
    ...real,
    run: vi.fn(() => ({ cancel() {}, done: Promise.resolve(true) })),
  };
});

import TableEffects from "./components/board/TableEffects.svelte";
import { crumble } from "./animations";
import { run } from "./motion";
import { play as playSound } from "./sounds";
import type { CardView, GameView, LogEvent } from "./protocol";
import { L } from "./labels";
import { resetSettings, updateSettings } from "./settings";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";
import {
  EFFECT_BASE_MS,
  EFFECT_CAP_MS,
  EFFECT_SHORT_CAP_MS,
  EFFECT_STILL_HOLD_MS,
  DAY_NIGHT_FLIP_DELAY_MS,
  FIZZLE_MOTES,
  TOKEN_STAGGER_MS,
  WIPE_FIRST_CARD_MS,
  REANIMATE_HAND_VOLUME,
  busyScale,
  pileBaseMs,
  shatterPlan,
} from "./tableEffects";

// ---- the table ----------------------------------------------------------

const permanent = (id: string, controller: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name: id,
    owner: controller,
    controller,
    known_by_you: true,
    type_line: "Creature — Ogre",
    scryfall_id: `sf-${id}`,
    ...extra,
  }) as unknown as CardView;

function viewWith(log: LogEvent[], cards: CardView[]): GameView {
  const v = dockTestView({ log } as Partial<GameView>);
  v.battlefield = { kind: "battlefield", count: cards.length, cards } as GameView["battlefield"];
  return v;
}

let seq = 100;
const zone = (
  id: string,
  newZone: string,
  extra: Partial<LogEvent> = {},
  oldZone = "battlefield",
): LogEvent => ({
  seq: ++seq,
  kind: "zone",
  turn: 3,
  seat: -1,
  card_id: id,
  old_zone: oldZone,
  new_zone: newZone,
  text: "",
  ...extra,
});

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

const TILE = { width: 80, height: 112 };

// A board with four permanents in a row, the opponent's avatar and one
// stack item.
function makeBoard(ids: string[]): HTMLElement {
  const board = document.createElement("div");
  board.getBoundingClientRect = () => new DOMRect(0, 0, boardSize.w, boardSize.h);
  ids.forEach((id, i) => {
    const el = rectEl("data-instance-id", id, { left: 100 + i * 100, top: 300, ...TILE });
    el.dataset.tapped = "false";
    el.appendChild(Object.assign(document.createElement("img"), { src: `/${id}.jpg` }));
    board.appendChild(el);
  });
  board.appendChild(rectEl("data-seat-id", "opp", { left: 560, top: 20, width: 60, height: 60 }));
  board.appendChild(
    rectEl("data-stack-item-id", "item-1", { left: 40, top: 200, width: 150, height: 210 }),
  );
  document.body.appendChild(board);
  return board;
}

function advance(ms: number) {
  // Effects first (a mount's animation-frame measure is scheduled by one),
  // then the clock.
  flushSync();
  vi.advanceTimersByTime(ms);
  flushSync();
}

function layer(): HTMLElement {
  return document.querySelector<HTMLElement>("[data-table-effects]")!;
}
const effects = (kind?: string) => [
  ...layer().querySelectorAll<HTMLElement>(kind ? `[data-effect="${kind}"]` : "[data-effect]"),
];
const badges = () => [...layer().querySelectorAll<HTMLElement>("[data-effect-badge]")];
// When each animation started ends: its delay plus its length.
const runEnds = () =>
  vi.mocked(run).mock.calls.map((c) => ((c[2].delay as number) ?? 0) + (c[2].duration as number));

const IDS = ["a", "b", "c", "d"];
const cardsAll = IDS.map((id) => permanent(id, "opp"));

// Mount on a primed frame, then deliver `next` as a new frame in which the
// cards in `gone` have left the board.
function deliver(next: LogEvent[], gone: string[] = [], over: Partial<CardView>[] = []) {
  const board = makeBoard(IDS);
  const before = cardsAll.map((c, i) => ({ ...c, ...(over[i] ?? {}) }) as CardView);
  const r = render(TableEffects as never, { view: viewWith([], before), boardEl: board } as never);
  // Let the layer measure the tiles as they stand before the frame.
  advance(50);
  r.setProps({
    view: viewWith(
      next,
      before.filter((c) => !gone.includes(c.instance_id)),
    ),
  } as never);
  flushSync();
  return { r, board };
}

beforeEach(() => {
  vi.useFakeTimers({
    toFake: [
      "Date",
      "setTimeout",
      "clearTimeout",
      "setInterval",
      "clearInterval",
      "requestAnimationFrame",
      "cancelAnimationFrame",
    ],
  });
  vi.setSystemTime(1_000_000);
  seq = 100;
  resetSettings();
  updateSettings("animations", "enabled", true);
  updateSettings("animations", "effects", "full");
  updateSettings("animations", "speed", 1);
  updateSettings("accessibility", "reduceMotion", false);
  vi.mocked(run).mockClear();
  vi.mocked(crumble).mockClear();
  vi.mocked(playSound).mockClear();
});

afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("the effects layer", () => {
  it("draws nothing on its first frame, and nothing for an old entry", () => {
    const board = makeBoard(IDS);
    render(
      TableEffects as never,
      {
        view: viewWith([zone("a", "graveyard", { cause: "destroy" })], cardsAll),
        boardEl: board,
      } as never,
    );
    flushSync();
    expect(effects()).toHaveLength(0);
    expect(playSound).not.toHaveBeenCalled();
  });

  it("is non-blocking and hidden from assistive tech", () => {
    deliver([zone("a", "graveyard", { cause: "destroy" })], ["a"]);
    expect(layer().getAttribute("aria-hidden")).toBe("true");
  });

  it("a destroy draws an art copy of the card, where it stood, and plays its sound", () => {
    deliver([zone("b", "graveyard", { cause: "destroy", batch: 4 })], ["b"]);
    const [cue] = effects("destroy");
    expect(cue).toBeTruthy();
    expect(cue.getAttribute("data-effect-card")).toBe("b");
    // b's tile centre: left 200 + 40, top 300 + 56.
    expect(cue.style.left).toBe("240px");
    expect(cue.style.top).toBe("356px");
    expect(cue.querySelector("img")?.getAttribute("src")).toBe("/b.jpg");
    // D1: it shatters from its centre outward, every piece its own art.
    const shards = cue.querySelectorAll(".shatter .shard");
    expect(shards.length).toBeGreaterThan(10);
    for (const sh of shards) expect(sh.querySelector("img")?.getAttribute("src")).toBe("/b.jpg");
    expect(cue.querySelectorAll(".crk path").length).toBeGreaterThan(0);
    expect(cue.querySelector(".slash")).toBeNull();
    expect(playSound).toHaveBeenCalledWith("destroy", { volume: 1, batch: "g1:4" });
    // Its motion was started, and every part of it ends inside the 2 s cap.
    const ends = vi
      .mocked(run)
      .mock.calls.map((c) => ((c[2].delay as number) ?? 0) + (c[2].duration as number));
    expect(ends.length).toBeGreaterThan(0);
    for (const ms of ends) expect(ms).toBeLessThanOrEqual(EFFECT_CAP_MS + 1e-6);
    // And it goes by itself.
    advance(EFFECT_CAP_MS + 200);
    expect(effects()).toHaveLength(0);
  });

  it("a destroy replaced by exile looks like an exile, with the exile sound", () => {
    deliver([zone("a", "exile", { cause: "destroy", batch: 5 })], ["a"]);
    expect(effects("destroy")).toHaveLength(0);
    expect(effects("exile")).toHaveLength(1);
    const [cue] = effects("exile");
    // D3's white void, and D4's slight crumble: it is drawn in pieces.
    expect(cue.querySelector(".void")).toBeTruthy();
    expect(cue.hasAttribute("data-effect-from-destroy")).toBe(true);
    expect(cue.querySelectorAll(".shatter .shard").length).toBeGreaterThan(0);
    expect(playSound).toHaveBeenCalledWith("exile", expect.anything());
    expect(playSound).not.toHaveBeenCalledWith("destroy", expect.anything());
  });

  it("a death by state-based action reuses the crumble, gated by the effects setting", () => {
    deliver([zone("c", "graveyard", { cause: "sba", batch: 6 })], ["c"]);
    expect(effects("dies")).toHaveLength(1);
    expect(crumble).toHaveBeenCalledTimes(1);
    expect(vi.mocked(crumble).mock.calls[0][2]).toMatchObject({ gate: "effects" });
    // No new sound for a plain death.
    expect(playSound).not.toHaveBeenCalled();
  });

  it("a sacrifice and a bounce draw their own cues, and the bounce flies to its owner's seat", () => {
    const sac: LogEvent = {
      seq: ++seq,
      kind: "sacrifice",
      seat: 0,
      card_id: "a",
      batch: 2,
      text: "",
    };
    deliver([sac, zone("b", "hand", { cause: "bounce", batch: 3 })], ["a", "b"]);
    expect(effects("sacrifice")).toHaveLength(1);
    // D5: an altar, a knife, and the destroy's shatter after it.
    const [sacCue] = effects("sacrifice");
    expect(sacCue.querySelector(".altar")).toBeTruthy();
    expect(sacCue.querySelector(".knife")).toBeTruthy();
    expect(sacCue.querySelectorAll(".shatter .shard").length).toBeGreaterThan(0);
    expect(effects("bounce")).toHaveLength(1);
    expect(playSound).toHaveBeenCalledWith("sacrifice", expect.anything());
    expect(playSound).toHaveBeenCalledWith("bounce", expect.anything());
    // The knife's own sound plays as it lands, not at once.
    expect(playSound).not.toHaveBeenCalledWith("sacrifice_knife", expect.anything());
    advance(EFFECT_CAP_MS / 2);
    expect(playSound).toHaveBeenCalledWith("sacrifice_knife", expect.anything());
    // D6: it turns over (its back is the card back) and goes home face up.
    const [bounceCue] = effects("bounce");
    expect(bounceCue.querySelector(".flip .face.back")).toBeTruthy();
    expect(bounceCue.querySelector(".flip .face:not(.back) img")).toBeTruthy();
    // The copy's flight (and its after-images') ends at the avatar of seat
    // "opp" (centre 590, 50) from b's centre (240, 356).
    const flights = vi.mocked(run).mock.calls.filter((c) => {
      const last = c[1][c[1].length - 1] as Record<string, unknown>;
      return last.translate === "350px -306px";
    });
    expect(flights.length).toBeGreaterThanOrEqual(1);
  });

  it("a counterspell shatters over the stack item's last box, with the counter sound", () => {
    const c: LogEvent = {
      seq: ++seq,
      kind: "counter",
      seat: 1,
      target: "bolt",
      stack_item_id: "item-1",
      batch: 8,
      text: "",
    };
    deliver([c]);
    const [cue] = effects("counter");
    expect(cue).toBeTruthy();
    expect(cue.style.left).toBe("115px");
    expect(cue.style.top).toBe("305px");
    expect(cue.querySelectorAll(".cracks path").length).toBeGreaterThan(0);
    expect(cue.querySelectorAll(".glass").length).toBeGreaterThan(0);
    // D7: a palm stopped it.
    expect(cue.querySelector(".palm")).toBeTruthy();
    expect(playSound).toHaveBeenCalledWith("counter", { volume: 1, batch: "g1:8" });
  });

  it("draws a permanent and a spell that go before an animation frame has measured them (#4410)", () => {
    const board = makeBoard([]);
    const r = render(TableEffects as never, { view: viewWith([], []), boardEl: board } as never);
    advance(50);
    // A creature enters and a spell is cast, in one frame...
    const tile = rectEl("data-instance-id", "z", { left: 700, top: 300, ...TILE });
    tile.appendChild(Object.assign(document.createElement("img"), { src: "/z.jpg" }));
    const spell = rectEl("data-stack-item-id", "item-2", {
      left: 900,
      top: 200,
      width: 150,
      height: 210,
    });
    board.append(tile, spell);
    const enters = zone("z", "battlefield", {}, "stack");
    r.setProps({ view: viewWith([enters], [permanent("z", "opp")]) } as never);
    flushSync();
    // ...and both go in the next, with no animation frame in between.
    tile.remove();
    spell.remove();
    const counter: LogEvent = {
      seq: ++seq,
      kind: "counter",
      seat: 1,
      target: "bolt",
      stack_item_id: "item-2",
      batch: 9,
      text: "",
    };
    r.setProps({
      view: viewWith(
        [enters, zone("z", "graveyard", { cause: "destroy", batch: 10 }), counter],
        [],
      ),
    } as never);
    flushSync();
    const [destroy] = effects("destroy");
    expect(destroy).toBeTruthy();
    expect(destroy.style.left).toBe("740px");
    expect(destroy.style.top).toBe("356px");
    expect(destroy.querySelector("img")?.getAttribute("src")).toBe("/z.jpg");
    const [countered] = effects("counter");
    expect(countered).toBeTruthy();
    expect(countered.style.left).toBe("975px");
  });

  it("still takes a box measured as its frame landed again on the next animation frame", () => {
    const board = makeBoard(["a"]);
    const r = render(
      TableEffects as never,
      { view: viewWith([], [permanent("a", "opp")]), boardEl: board } as never,
    );
    advance(50);
    // The board reflows and a moves; the next frame's animation frame
    // measures it where it is now.
    const tile = board.querySelector<HTMLElement>('[data-instance-id="a"]')!;
    tile.getBoundingClientRect = () => new DOMRect(500, 300, TILE.width, TILE.height);
    r.setProps({ view: viewWith([], [permanent("a", "opp")]) } as never);
    advance(50);
    r.setProps({
      view: viewWith([zone("a", "graveyard", { cause: "destroy", batch: 11 })], []),
    } as never);
    flushSync();
    const [destroy] = effects("destroy");
    expect(destroy).toBeTruthy();
    expect(destroy.style.left).toBe("540px");
  });

  it("a planeswalker that enters glows, and one that dies cracks", async () => {
    const board = makeBoard(IDS);
    const jace = permanent("a", "opp", { type_line: "Legendary Planeswalker — Jace" });
    const r = render(
      TableEffects as never,
      { view: viewWith([], [jace]), boardEl: board } as never,
    );
    advance(50);
    // It enters (it was on the battlefield only in this frame).
    r.setProps({ view: viewWith([zone("a", "battlefield", {}, "stack")], [jace]) } as never);
    flushSync();
    // Its box is read once its entrance pop has settled.
    await vi.advanceTimersByTimeAsync(0);
    flushSync();
    expect(effects("pw_enter")).toHaveLength(1);
    expect(playSound).toHaveBeenCalledWith("pw_enter", expect.anything());
    // Later it dies.
    advance(EFFECT_CAP_MS + 200);
    advance(50);
    r.setProps({
      view: viewWith([zone("a", "graveyard", { cause: "destroy" })], []),
    } as never);
    flushSync();
    expect(effects("pw_leave")).toHaveLength(1);
    expect(effects("pw_leave")[0].querySelector(".crack")).toBeTruthy();
    // D9: and it leaves the plane through a rift.
    expect(effects("pw_leave")[0].querySelector(".rift")).toBeTruthy();
    expect(playSound).toHaveBeenCalledWith("pw_leave", expect.anything());
  });

  it("a board wipe is one sweep, one board_wipe sound and lighter per-card cues", () => {
    const wipe = IDS.slice(0, 4).map((id) =>
      zone(id, "graveyard", { cause: "destroy", batch: 9, source: "damnation" }),
    );
    deliver(wipe, IDS);
    expect(effects("board_wipe")).toHaveLength(1);
    // D10: quick cuts slice across the board.
    expect(effects("board_wipe")[0].querySelectorAll(".cut .blade").length).toBeGreaterThan(3);
    const cards = effects("destroy");
    expect(cards).toHaveLength(4);
    for (const c of cards) {
      expect(c.hasAttribute("data-effect-light")).toBe(true);
      // Every card plays the destroy's shatter, in fewer pieces.
      const pieces = c.querySelectorAll(".shatter .shard").length;
      expect(pieces).toBeGreaterThan(0);
      expect(pieces).toBeLessThan(shatterPlan("x", 80, 112).shards.length);
    }
    const wipes = vi.mocked(playSound).mock.calls.filter((c) => c[0] === "board_wipe");
    expect(wipes).toHaveLength(1);
    // The per-card sound is quieter than a lone destroy's.
    const destroys = vi.mocked(playSound).mock.calls.filter((c) => c[0] === "destroy");
    for (const c of destroys) expect(c[1]?.volume).toBeLessThan(1);
    // The whole wipe is over within the cap.
    advance(EFFECT_CAP_MS + 200);
    expect(effects()).toHaveLength(0);
  });

  it("scales a cue's length by animations.speed, inside the cap times the speed", () => {
    updateSettings("animations", "speed", 2);
    deliver([zone("b", "graveyard", { cause: "destroy" })], ["b"]);
    const ends = runEnds();
    expect(Math.max(...ends)).toBeCloseTo(EFFECT_BASE_MS.destroy * 2, 6);
    for (const ms of ends) expect(ms).toBeLessThanOrEqual(EFFECT_CAP_MS * 2 + 1e-6);
  });

  it("Short cuts a cue to 1 s", () => {
    updateSettings("animations", "effects", "short");
    deliver([zone("a", "exile", { cause: "exile_effect" })], ["a"]);
    expect(Math.max(...runEnds())).toBeCloseTo(EFFECT_SHORT_CAP_MS, 6);
  });

  it("a busy frame plays its cues shorter, and a lone one in full", () => {
    deliver([zone("a", "graveyard", { cause: "destroy", batch: 1 })], ["a"]);
    const lone = Math.max(...runEnds());
    expect(lone).toBeCloseTo(EFFECT_BASE_MS.destroy, 6);
    cleanup();
    document.body.innerHTML = "";
    vi.mocked(run).mockClear();
    // Four destroys from four sources at once: none is a wipe, but the
    // table is busy, so each plays shorter.
    deliver(
      IDS.map((id, i) => zone(id, "graveyard", { cause: "destroy", batch: 20 + i })),
      IDS,
    );
    expect(effects("destroy")).toHaveLength(4);
    const busy = Math.max(...runEnds());
    expect(busy).toBeCloseTo(EFFECT_BASE_MS.destroy * busyScale(4), 6);
    expect(busy).toBeLessThan(lone);
  });

  it("a wipe's cards stay whole and still from the frame until their own cue starts", () => {
    // A death wipe (Blasphemous Act without its scene): each card's
    // crumble waits for its stagger, so it is still there when its turn
    // (or a fireball) comes, instead of crumbling as the frame lands.
    const deaths = IDS.map((id) =>
      zone(id, "graveyard", { cause: "sba", batch: 11, source: "act" }),
    );
    deliver(deaths, IDS);
    expect(effects("dies")).toHaveLength(4);
    const waits = vi.mocked(crumble).mock.calls.map((c) => c[2]?.delay ?? 0);
    expect(waits).toHaveLength(4);
    expect(waits[0]).toBeCloseTo(WIPE_FIRST_CARD_MS, 6);
    for (let i = 1; i < waits.length; i++) expect(waits[i]).toBeGreaterThan(waits[i - 1]);
    cleanup();
    document.body.innerHTML = "";
    vi.mocked(run).mockClear();
    // A destroy wipe: every card's copy is mounted at once, and each of
    // its animations waits for the card's turn, holding its first frame
    // (the card at rest) until then.
    const destroys = IDS.map((id) =>
      zone(id, "graveyard", { cause: "destroy", batch: 12, source: "damnation" }),
    );
    deliver(destroys, IDS);
    expect(effects("destroy")).toHaveLength(4);
    for (const c of effects("destroy")) expect(c.querySelector(".face img")).toBeTruthy();
    const onCards = vi
      .mocked(run)
      .mock.calls.filter((c) => (c[0] as Element).closest('[data-effect="destroy"]'));
    expect(onCards.length).toBeGreaterThan(0);
    for (const c of onCards) {
      expect(c[2].fill).toBe("both");
      expect(c[2].delay as number).toBeGreaterThanOrEqual(WIPE_FIRST_CARD_MS - 1e-6);
    }
  });

  it("an exile wipe swirls into one white void, and its cards are drawn into it", () => {
    const wipe = IDS.slice(0, 4).map((id) =>
      zone(id, "exile", { cause: "exile_effect", batch: 9, source: "farewell" }),
    );
    deliver(wipe, IDS);
    const [w] = effects("board_wipe");
    expect(w.getAttribute("data-tone")).toBe("cold");
    expect(w.querySelector(".swirl path")).toBeTruthy();
    expect(w.querySelector(".whitevoid")).toBeTruthy();
    expect(w.querySelector(".core")).toBeTruthy();
    const cards = effects("exile");
    expect(cards).toHaveLength(4);
    // Each card goes into the board's one void, not a void of its own.
    for (const c of cards) expect(c.querySelector(".void")).toBeNull();
    advance(EFFECT_CAP_MS + 200);
    expect(effects()).toHaveLength(0);
  });
});

describe("a spell or ability fizzles (#4450)", () => {
  const fizzle = (extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "fizzle",
    seat: 1,
    card_id: "bolt",
    stack_item_id: "item-1",
    batch: 8,
    text: "Lightning Bolt fizzled: no legal targets",
    ...extra,
  });

  it("dissolves over the stack item's last box, with no palm and no glass, and plays one sound", () => {
    deliver([fizzle()]);
    const [cue] = effects("fizzle");
    expect(cue).toBeTruthy();
    expect(effects("counter")).toHaveLength(0);
    // item-1's centre: left 40 + 75, top 200 + 105.
    expect(cue.style.left).toBe("115px");
    expect(cue.style.top).toBe("305px");
    expect(cue.querySelector(".fmist")).toBeTruthy();
    expect(cue.querySelector(".fedge")).toBeTruthy();
    expect(cue.querySelectorAll(".fmote")).toHaveLength(FIZZLE_MOTES);
    expect(cue.querySelector(".palm")).toBeNull();
    expect(cue.querySelector(".glass")).toBeNull();
    expect(playSound).toHaveBeenCalledTimes(1);
    expect(playSound).toHaveBeenCalledWith("fizzle", { volume: 1, batch: "g1:8" });
    // Every part of it ends inside its own 0.8 s.
    const ends = runEnds();
    expect(ends.length).toBeGreaterThan(0);
    for (const ms of ends) expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.fizzle + 1e-6);
    advance(EFFECT_BASE_MS.fizzle + 200);
    expect(effects()).toHaveLength(0);
  });

  it("Short plays the same dissolve, and speed scales it", () => {
    updateSettings("animations", "effects", "short");
    updateSettings("animations", "speed", 2);
    deliver([fizzle()]);
    expect(effects("fizzle")).toHaveLength(1);
    expect(Math.max(...runEnds())).toBeLessThanOrEqual(EFFECT_BASE_MS.fizzle * 2 + 1e-6);
    expect(Math.max(...runEnds())).toBeGreaterThan(EFFECT_BASE_MS.fizzle);
  });

  it("an optional trigger that lapsed draws nothing and plays nothing", () => {
    deliver([fizzle({ call: "condition" })]);
    expect(effects()).toHaveLength(0);
    expect(playSound).not.toHaveBeenCalled();
  });

  it.each([
    ["reduced motion", () => updateSettings("accessibility", "reduceMotion", true)],
    ["Off", () => updateSettings("animations", "effects", "off")],
  ])("still (%s): nothing is drawn, no badge of its own, and the sound plays once", (_, still) => {
    still();
    deliver([fizzle()]);
    // The stack linger's "Fizzled" badge says it.
    expect(effects()).toHaveLength(0);
    expect(badges()).toHaveLength(0);
    expect(run).not.toHaveBeenCalled();
    expect(playSound).toHaveBeenCalledTimes(1);
    expect(playSound).toHaveBeenCalledWith("fizzle", expect.anything());
  });

  it("a skip clears the dissolve", () => {
    deliver([fizzle()]);
    expect(effects("fizzle")).toHaveLength(1);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(effects()).toHaveLength(0);
  });
});

describe("a token is created (#4441)", () => {
  const TOKENS = ["t1", "t2", "t3"];
  const tokenEntry = (id: string, batch?: number): LogEvent => ({
    seq: ++seq,
    kind: "token",
    turn: 3,
    seat: 1,
    card_id: id,
    batch,
    text: "",
  });
  const tile = (board: HTMLElement, id: string) =>
    board.querySelector<HTMLElement>(`[data-instance-id="${id}"]`)!;
  const tokenSounds = () => vi.mocked(playSound).mock.calls.filter(([n]) => n === "token_create");
  // The pop-in run on a token's own tile, if one was started.
  const popOf = (el: HTMLElement) => vi.mocked(run).mock.calls.find((c) => c[0] === el);

  // A board where "a" stood before the frame and the tokens are new in it.
  function create(log: LogEvent[], made = TOKENS) {
    const board = makeBoard(["a", ...TOKENS]);
    const a = permanent("a", "opp");
    const r = render(TableEffects as never, { view: viewWith([], [a]), boardEl: board } as never);
    advance(50);
    const tokens = made.map((id) => permanent(id, "opp", { is_token: true }));
    r.setProps({ view: viewWith(log, [a, ...tokens]) } as never);
    flushSync();
    return { r, board };
  }

  it("pops each token of a batch in, a stagger apart, with one sound for the batch", () => {
    const { board } = create(TOKENS.map((id) => tokenEntry(id, 7)));
    // A glow over each new token, at its tile's centre (t1: 200 + 40, 300 + 56).
    const cues = effects("token_create");
    expect(cues.map((c) => c.getAttribute("data-effect-card"))).toEqual(TOKENS);
    expect(cues[0].style.left).toBe("240px");
    expect(cues[0].style.top).toBe("356px");
    for (const c of cues) expect(c.querySelector(".tglow")).toBeTruthy();
    // Each token's own tile pops in from about 80%, hidden until its turn.
    const pops = TOKENS.map((id) => popOf(tile(board, id)));
    for (const p of pops) {
      expect(p).toBeTruthy();
      const first = p![1][0] as Record<string, unknown>;
      expect(first.scale).toBe("0.8");
      expect(first.opacity).toBe(0);
      expect(p![2].fill).toBe("backwards");
      expect(p![2].duration).toBeCloseTo(EFFECT_BASE_MS.token_create, 6);
    }
    expect(pops.map((p) => p![2].delay)).toEqual([0, TOKEN_STAGGER_MS, 2 * TOKEN_STAGGER_MS]);
    // One sound for the batch, at once, however many tokens it made.
    expect(tokenSounds()).toEqual([["token_create", { volume: 1, batch: "g1:7" }]]);
    advance(EFFECT_CAP_MS + 200);
    expect(tokenSounds()).toHaveLength(1);
    expect(effects()).toHaveLength(0);
  });

  it("two batches in one frame are two sounds", () => {
    create([tokenEntry("t1", 7), tokenEntry("t2", 8)], ["t1", "t2"]);
    expect(effects("token_create")).toHaveLength(2);
    expect(tokenSounds().map((c) => c[1]?.batch)).toEqual(["g1:7", "g1:8"]);
  });

  it("Short plays the same small pop, and speed scales it and its stagger", () => {
    updateSettings("animations", "effects", "short");
    updateSettings("animations", "speed", 2);
    const { board } = create(TOKENS.map((id) => tokenEntry(id, 7)));
    const last = popOf(tile(board, "t3"))!;
    expect(last[2].delay).toBeCloseTo(2 * TOKEN_STAGGER_MS * 2, 6);
    expect(last[2].duration).toBeCloseTo(EFFECT_BASE_MS.token_create * 2, 6);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS * 2 + 1e-6);
  });

  it("still (reduced motion, Off): no pop and no glow, and the sound still plays once", () => {
    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      still();
      create(TOKENS.map((id) => tokenEntry(id, 7)));
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(tokenSounds()).toHaveLength(1);
      cleanup();
      document.body.innerHTML = "";
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(playSound).mockClear();
    }
  });

  it("a skip cancels the pop-ins, so every token is at rest", () => {
    const cancelled: Element[] = [];
    vi.mocked(run).mockImplementation((el) => ({
      cancel: () => cancelled.push(el),
      done: Promise.resolve(true),
    }));
    try {
      const { board } = create(TOKENS.map((id) => tokenEntry(id, 7)));
      window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
      flushSync();
      expect(effects()).toHaveLength(0);
      for (const id of TOKENS) expect(cancelled).toContain(tile(board, id));
    } finally {
      vi.mocked(run).mockImplementation(() => ({ cancel() {}, done: Promise.resolve(true) }));
    }
  });

  it("a token already gone draws nothing, and its sound still plays", () => {
    create([tokenEntry("t1", 7)], []);
    expect(effects()).toHaveLength(0);
    expect(run).not.toHaveBeenCalled();
    expect(tokenSounds()).toHaveLength(1);
  });
});

describe("counters are put on a permanent (#4442)", () => {
  const countersEntry = (
    id: string,
    label: string,
    delta: number,
    batch: number | undefined = 7,
  ): LogEvent => ({
    seq: ++seq,
    kind: "counters",
    turn: 3,
    seat: -1,
    card_id: id,
    label,
    amount: Math.max(0, delta),
    delta,
    batch,
    text: "",
  });
  const counterSounds = () =>
    vi.mocked(playSound).mock.calls.filter(([n]) => n === "counter_add" || n === "counter_remove");
  // The run started on the element of the cue for `card` matching `sel`.
  const runOn = (card: string, sel: string) => {
    const el = layer().querySelector(`[data-effect-card="${card}"] ${sel}`);
    return vi.mocked(run).mock.calls.find((c) => c[0] === el);
  };

  // "a" shows its power/toughness, "b" a charge counter's pip; "c" has
  // neither (a token group's member).
  function put(log: LogEvent[]) {
    const board = makeBoard(["a", "b", "c"]);
    const tile = (id: string) => board.querySelector<HTMLElement>(`[data-instance-id="${id}"]`)!;
    const pt = rectEl("class", "badge pt", { left: 150, top: 390, width: 26, height: 16 });
    tile("a").appendChild(pt);
    tile("b").appendChild(
      rectEl("data-counter", "charge", { left: 255, top: 305, width: 30, height: 14 }),
    );
    const cards = ["a", "b", "c"].map((id) => permanent(id, "opp"));
    const r = render(TableEffects as never, { view: viewWith([], cards), boardEl: board } as never);
    advance(50);
    r.setProps({ view: viewWith(log, cards) } as never);
    flushSync();
    return { r, board };
  }

  it("rings each card's power/toughness, green or purple with its tick, and plays one sound for the batch", () => {
    put([countersEntry("a", "+1/+1", 2), countersEntry("c", "-1/-1", 1)]);
    const cues = effects("counters");
    expect(cues.map((c) => [c.dataset.effectCard, c.dataset.effectTone])).toEqual([
      ["a", "plus"],
      ["c", "minus"],
    ]);
    // Over a's badge (150 + 13, 390 + 8).
    expect(cues[0].style.left).toBe("163px");
    expect(cues[0].style.top).toBe("398px");
    // c has no badge: its tile's bottom-right corner (100 + 2 * 100 + 80 - 36 - 4 + 18, 300 + 112 - 22 + 9).
    expect(cues[1].style.left).toBe("358px");
    expect(cues[1].style.top).toBe("399px");
    for (const c of cues) {
      expect(c.querySelector(".cpulse")).toBeTruthy();
      expect(c.querySelector(".ctick")).toBeTruthy();
    }
    // The tick rises for +1/+1 and falls for -1/-1.
    const tickA = runOn("a", ".ctick")!;
    const tickC = runOn("c", ".ctick")!;
    const ys = (call: NonNullable<typeof tickA>) =>
      (call[1] as Record<string, unknown>[])
        .map((k) => k.translate)
        .filter(Boolean)
        .map((t) => Number(String(t).split(" ")[1].replace("px", "")));
    expect(ys(tickA)[0]).toBeGreaterThan(ys(tickA).at(-1)!);
    expect(ys(tickC)[0]).toBeLessThan(ys(tickC).at(-1)!);
    expect(runOn("a", ".cpulse")![2].duration).toBeCloseTo(EFFECT_BASE_MS.counters, 6);
    // One sound for the batch, the first of its cues to have one.
    expect(counterSounds()).toEqual([["counter_add", { volume: 1, batch: "g1:7" }]]);
    advance(EFFECT_CAP_MS + 200);
    expect(counterSounds()).toHaveLength(1);
    expect(effects()).toHaveLength(0);
  });

  it("a -1/-1 batch plays counter_remove, and two batches are two sounds", () => {
    put([countersEntry("a", "-1/-1", 1, 7), countersEntry("c", "+1/+1", 1, 8)]);
    expect(counterSounds().map((c) => [c[0], c[1]?.batch])).toEqual([
      ["counter_remove", "g1:7"],
      ["counter_add", "g1:8"],
    ]);
  });

  it("another kind rings its own pip, with no tick and no sound", () => {
    put([countersEntry("b", "charge", 1)]);
    const [cue] = effects("counters");
    expect(cue.dataset.effectTone).toBe("other");
    expect(cue.style.left).toBe("270px");
    expect(cue.style.top).toBe("312px");
    expect(cue.querySelector(".ctick")).toBeNull();
    expect(playSound).not.toHaveBeenCalled();
  });

  it("a counter taken off is no cue", () => {
    put([countersEntry("a", "+1/+1", -1), countersEntry("a", "-1/-1", -1)]);
    expect(effects()).toHaveLength(0);
    expect(playSound).not.toHaveBeenCalled();
  });

  it("Short plays the same pulse inside its cap", () => {
    updateSettings("animations", "effects", "short");
    put([countersEntry("a", "+1/+1", 1)]);
    expect(effects("counters")).toHaveLength(1);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS + 1e-6);
  });

  it("still (reduced motion, Off): nothing is drawn, and the sound still plays once", () => {
    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      still();
      put([countersEntry("a", "+1/+1", 1), countersEntry("c", "+1/+1", 1)]);
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(counterSounds()).toHaveLength(1);
      cleanup();
      document.body.innerHTML = "";
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(playSound).mockClear();
    }
  });

  it("a skip clears the pulse", () => {
    put([countersEntry("a", "+1/+1", 1)]);
    expect(effects("counters")).toHaveLength(1);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(effects()).toHaveLength(0);
  });
});

describe("a commander is cast from the command zone, or goes home (#4443)", () => {
  const cmdr = permanent("cmdr", "opp", {
    is_commander: true,
    type_line: "Legendary Creature — Elf Druid",
  });
  // The opponent's command zone card, the spell on the stack and their
  // graveyard pile, each where the board draws it.
  const CMD = { left: 900, top: 650, width: 50, height: 70 };
  const STACK = { left: 40, top: 200, width: 150, height: 210 };
  const PILE = { left: 1000, top: 700, width: 40, height: 56 };
  const centre = (r: typeof CMD) => ({ x: r.left + r.width / 2, y: r.top + r.height / 2 });

  function tableView(
    log: LogEvent[],
    at: {
      battlefield?: CardView[];
      command?: CardView[];
      graveyard?: CardView[];
      stack?: CardView[];
    },
  ): GameView {
    const v = viewWith(log, at.battlefield ?? []);
    const opp = v.seats[1];
    const z = (kind: string, cards: CardView[]) =>
      ({ kind, owner: "opp", count: cards.length, cards }) as GameView["stack"];
    opp.command = z("command", at.command ?? []);
    opp.graveyard = z("graveyard", at.graveyard ?? []);
    v.stack = z("stack", at.stack ?? []);
    return v;
  }
  const commandCard = () => rectEl("data-instance-id", "cmdr", CMD);
  const sounds = (name: string) => vi.mocked(playSound).mock.calls.filter(([n]) => n === name);
  const runsOn = (el: Element) => vi.mocked(run).mock.calls.filter((c) => c[0] === el);
  const cast = (batch: number): LogEvent => ({
    seq: ++seq,
    kind: "cast",
    turn: 3,
    seat: 1,
    card_id: "cmdr",
    old_zone: "command",
    batch,
    text: "",
  });

  // The commander sits in its command zone; then it is cast.
  function castIt() {
    const board = makeBoard([]);
    const inCommand = commandCard();
    board.appendChild(inCommand);
    board.appendChild(rectEl("data-stack-item-id", "cmdr", STACK));
    const r = render(
      TableEffects as never,
      {
        view: tableView([], { command: [cmdr] }),
        boardEl: board,
      } as never,
    );
    advance(50);
    inCommand.remove();
    r.setProps({ view: tableView([cast(4)], { stack: [cmdr] }) } as never);
    flushSync();
    return board;
  }

  // The commander is somewhere else (`before`); then it goes home.
  function goHome(entry: LogEvent, before: { battlefield?: CardView[]; graveyard?: CardView[] }) {
    const board = makeBoard((before.battlefield ?? []).map((c) => c.instance_id));
    const pile = rectEl("data-pile", "graveyard", PILE);
    pile.setAttribute("data-pile-owner", "opp");
    board.appendChild(pile);
    const r = render(
      TableEffects as never,
      {
        view: tableView([], before),
        boardEl: board,
      } as never,
    );
    advance(50);
    for (const c of before.battlefield ?? []) {
      board.querySelector(`[data-instance-id="${c.instance_id}"]`)?.remove();
    }
    const home = commandCard();
    board.appendChild(home);
    r.setProps({ view: tableView([entry], { command: [cmdr] }) } as never);
    flushSync();
    return { board, home };
  }

  it("a cast rises in gold from where it sat to its spell on the stack, with one sound", () => {
    castIt();
    const cues = effects("commander_cast");
    expect(cues).toHaveLength(1);
    const from = centre(CMD);
    const to = centre(STACK);
    expect(cues[0].style.left).toBe(`${from.x}px`);
    expect(cues[0].style.top).toBe(`${from.y}px`);
    const crown = cues[0].querySelector<HTMLElement>(".cmd-crown")!;
    expect(crown.style.left).toBe(`${to.x - from.x}px`);
    expect(crown.style.top).toBe(`${to.y - from.y}px`);
    for (const part of [".cmd-bloom", ".cmd-beam", ".cmd-crown"]) {
      const el = cues[0].querySelector(part)!;
      expect(runsOn(el)).toHaveLength(1);
    }
    expect(cues[0].querySelectorAll(".cmd-mote").length).toBeGreaterThan(0);
    // The light runs the cue's length, and nothing waits on it.
    for (const ms of runEnds())
      expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.commander_cast + 1e-6);
    expect(sounds("commander_cast")).toEqual([["commander_cast", { volume: 1, batch: "g1:4" }]]);
    advance(EFFECT_CAP_MS + 200);
    expect(sounds("commander_cast")).toHaveLength(1);
    expect(effects()).toHaveLength(0);
  });

  it("a return from the graveyard (CR 903.9a) drifts from the pile to its card, which waits hidden", () => {
    const { home } = goHome(zone("cmdr", "command", { batch: 8 }, "graveyard"), {
      graveyard: [cmdr],
    });
    const cues = effects("commander_return");
    expect(cues).toHaveLength(1);
    const from = centre(PILE);
    const to = centre(CMD);
    expect(cues[0].style.left).toBe(`${from.x}px`);
    expect(cues[0].style.top).toBe(`${from.y}px`);
    expect(cues[0].getAttribute("data-effect-lands")).toBe("card");
    const homeRing = cues[0].querySelector<HTMLElement>(".cmd-home")!;
    expect(homeRing.style.left).toBe(`${to.x - from.x}px`);
    expect(homeRing.style.top).toBe(`${to.y - from.y}px`);
    // The copy and its two after-images show the card, at its size.
    expect(cues[0].querySelectorAll(".cmd-ghost")).toHaveLength(2);
    const face = cues[0].querySelector<HTMLElement>(".cmd-copy .face")!;
    expect(face.querySelector("img")).toBeTruthy();
    expect(face.style.width).toBe(`${CMD.width}px`);
    // The card in the command zone is hidden until the copy lands.
    const held = runsOn(home);
    expect(held).toHaveLength(1);
    expect((held[0][1][0] as Record<string, unknown>).opacity).toBe(0);
    expect(held[0][2].fill).toBe("backwards");
    expect(held[0][2].duration).toBeCloseTo(EFFECT_BASE_MS.commander_return, 6);
    expect(sounds("commander_return")).toEqual([
      ["commander_return", { volume: 1, batch: "g1:8" }],
    ]);
  });

  it("a bounce its owner sent home instead (CR 903.9b) is the return from its tile, not a bounce", () => {
    goHome(zone("cmdr", "command", { cause: "bounce", batch: 9 }), {
      battlefield: [cmdr],
    });
    expect(effects("bounce")).toHaveLength(0);
    expect(sounds("bounce")).toHaveLength(0);
    const cues = effects("commander_return");
    expect(cues).toHaveLength(1);
    // makeBoard's first tile: 100 + 40, 300 + 56.
    expect(cues[0].style.left).toBe("140px");
    expect(cues[0].style.top).toBe("356px");
    expect(cues[0].querySelector<HTMLElement>(".cmd-copy .face")!.style.width).toBe(
      `${TILE.width}px`,
    );
    expect(sounds("commander_return")).toHaveLength(1);
  });

  it("a card put into the command zone that is not a commander gets no cue", () => {
    const attraction = permanent("cmdr", "opp", { type_line: "Artifact — Attraction" });
    const board = makeBoard([]);
    const r = render(
      TableEffects as never,
      {
        view: tableView([], { battlefield: [attraction] }),
        boardEl: board,
      } as never,
    );
    advance(50);
    const v = tableView([zone("cmdr", "command", { batch: 3 })], {});
    v.seats[1].command = {
      kind: "command",
      owner: "opp",
      count: 1,
      cards: [attraction],
    } as GameView["stack"];
    r.setProps({ view: v } as never);
    flushSync();
    expect(effects("commander_return")).toHaveLength(0);
    expect(sounds("commander_return")).toHaveLength(0);
  });

  it("a skip shows the waiting card at once", () => {
    const cancelled: Element[] = [];
    vi.mocked(run).mockImplementation((el) => ({
      cancel: () => cancelled.push(el),
      done: Promise.resolve(true),
    }));
    try {
      const { home } = goHome(zone("cmdr", "command", { batch: 8 }, "graveyard"), {
        graveyard: [cmdr],
      });
      window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
      flushSync();
      expect(effects()).toHaveLength(0);
      expect(cancelled).toContain(home);
    } finally {
      vi.mocked(run).mockImplementation(() => ({ cancel() {}, done: Promise.resolve(true) }));
    }
  });

  it("Short plays the same, and still (reduced motion, Off) draws nothing and plays each sound once", () => {
    updateSettings("animations", "effects", "short");
    castIt();
    expect(effects("commander_cast")).toHaveLength(1);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS + 1e-6);
    cleanup();
    document.body.innerHTML = "";

    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(playSound).mockClear();
      vi.mocked(run).mockClear();
      still();
      castIt();
      cleanup();
      document.body.innerHTML = "";
      goHome(zone("cmdr", "command", { batch: 8 }, "graveyard"), { graveyard: [cmdr] });
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(sounds("commander_cast")).toHaveLength(1);
      expect(sounds("commander_return")).toHaveLength(1);
      cleanup();
      document.body.innerHTML = "";
    }
  });
});

describe("an Equipment is attached, or an Aura enchants a permanent (#4444)", () => {
  const attachEntry = (id: string, host: string, label: string, batch = 7): LogEvent => ({
    seq: ++seq,
    kind: "attach",
    turn: 3,
    seat: 1,
    card_id: id,
    target: host,
    label,
    batch,
    text: "",
  });
  const tokenEntry = (id: string, batch = 7): LogEvent => ({
    seq: ++seq,
    kind: "token",
    turn: 3,
    seat: 1,
    card_id: id,
    batch,
    text: "",
  });
  const sounds = (name: string) => vi.mocked(playSound).mock.calls.filter(([n]) => n === name);
  const runsOn = (el: Element) => vi.mocked(run).mock.calls.filter((c) => c[0] === el);
  // The host "a" is makeBoard's first tile (100, 300); an attachment is
  // drawn tucked under it, out to its left and a little lower.
  const UNDER = { left: 73, top: 308, ...TILE };
  const at = (el: HTMLElement, r: { left: number; top: number; width: number; height: number }) => {
    el.getBoundingClientRect = () => new DOMRect(r.left, r.top, r.width, r.height);
  };
  const tile = (board: HTMLElement, id: string) =>
    board.querySelector<HTMLElement>(`[data-instance-id="${id}"]`)!;
  // The slide started on a tile: its first keyframe's translate.
  const slideFrom = (el: Element) => {
    const call = runsOn(el).find((c) => (c[1] as Keyframe[])[0].translate !== undefined)!;
    return { call, from: String((call[1] as Keyframe[])[0].translate) };
  };

  // "a" and "sword" stand on the board; then the sword is attached to "a",
  // and the board draws it under "a".
  function equip(log: LogEvent[]) {
    const board = makeBoard(["a", "sword"]);
    const cards = [
      permanent("a", "opp"),
      permanent("sword", "opp", { type_line: "Artifact — Equipment" }),
    ];
    const r = render(TableEffects as never, { view: viewWith([], cards), boardEl: board } as never);
    advance(50);
    at(tile(board, "sword"), UNDER);
    r.setProps({ view: viewWith(log, cards) } as never);
    flushSync();
    return board;
  }

  // "a" stands alone; then `ids` are new on the board, under "a".
  function arrive(log: LogEvent[], ids: string[], extra: Partial<CardView> = {}) {
    const board = makeBoard(["a"]);
    const a = permanent("a", "opp");
    const r = render(TableEffects as never, { view: viewWith([], [a]), boardEl: board } as never);
    advance(50);
    for (const id of ids) {
      const el = rectEl("data-instance-id", id, UNDER);
      el.dataset.tapped = "false";
      board.appendChild(el);
    }
    const made = ids.map((id) =>
      permanent(id, "opp", { type_line: "Enchantment — Aura", ...extra }),
    );
    r.setProps({ view: viewWith(log, [a, ...made]) } as never);
    flushSync();
    return board;
  }

  it("an Equipment slides from where it stood onto its host, clasps with a spark, and plays one sound", () => {
    const board = equip([attachEntry("sword", "a", "Equipment", 7)]);
    const cues = effects("attach_equip");
    expect(cues).toHaveLength(1);
    expect(cues[0].getAttribute("data-effect-card")).toBe("sword");
    // Centred on the host (100 + 40, 300 + 56).
    expect(cues[0].style.left).toBe("140px");
    expect(cues[0].style.top).toBe("356px");
    // The flash and the sparks sit at the host's left edge.
    const seam = cues[0].querySelector<HTMLElement>(".att-seam")!;
    expect(seam.style.left).toBe(`${-TILE.width / 2 + 2}px`);
    expect(cues[0].querySelector(".att-flash")).toBeTruthy();
    expect(cues[0].querySelectorAll(".att-spark").length).toBeGreaterThan(0);
    expect(runsOn(cues[0].querySelector(".att-rim")!)).toHaveLength(1);
    // The sword's own tile slides from its old box (200, 300) to its rest
    // under the host, held there from the frame on.
    const { call, from } = slideFrom(tile(board, "sword"));
    expect(from).toBe(`${200 - UNDER.left}px ${300 - UNDER.top}px`);
    expect(String((call[1] as Keyframe[]).at(-1)!.translate)).toBe("0px 0px");
    expect(call[2].fill).toBe("backwards");
    expect(call[2].duration).toBeCloseTo(EFFECT_BASE_MS.attach_equip, 6);
    // The host itself does not move.
    expect(runsOn(tile(board, "a"))).toHaveLength(0);
    expect(sounds("attach_equip")).toEqual([["attach_equip", { volume: 1, batch: "g1:7" }]]);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.attach_equip + 1e-6);
    advance(EFFECT_CAP_MS + 200);
    expect(sounds("attach_equip")).toHaveLength(1);
    expect(effects()).toHaveLength(0);
  });

  it("an Aura new on the battlefield slides in from just above its rest, and a soft light rings its host", () => {
    const board = arrive([attachEntry("aura", "a", "Aura", 8)], ["aura"]);
    const cues = effects("attach_aura");
    expect(cues).toHaveLength(1);
    expect(cues[0].style.left).toBe("140px");
    expect(cues[0].style.top).toBe("356px");
    const halo = cues[0].querySelector<HTMLElement>(".att-halo")!;
    expect(halo.style.width).toBe(`${TILE.width + 10}px`);
    expect(runsOn(halo)).toHaveLength(1);
    const glints = [...cues[0].querySelectorAll(".att-glint")];
    expect(glints.length).toBeGreaterThan(0);
    for (const g of glints) expect(runsOn(g)).toHaveLength(1);
    const { call, from } = slideFrom(tile(board, "aura"));
    expect(from).toBe(`${-0.3 * TILE.width}px ${-0.45 * TILE.height}px`);
    expect(call[2].duration).toBeCloseTo(EFFECT_BASE_MS.attach_aura, 6);
    expect(sounds("attach_aura")).toEqual([["attach_aura", { volume: 1, batch: "g1:8" }]]);
  });

  it("a token created attached pops in and slides onto its host, with the attach sound alone", () => {
    // A Role, or a token copy of an Aura spell: `token` and `attach` share
    // the batch.
    const board = arrive([tokenEntry("role", 9), attachEntry("role", "a", "Aura", 9)], ["role"], {
      is_token: true,
    });
    expect(effects("token_create")).toHaveLength(1);
    expect(effects("attach_aura")).toHaveLength(1);
    // Two runs on its tile: the pop (scale, opacity) and the slide
    // (translate, rotate), which play together.
    const onTile = runsOn(tile(board, "role"));
    expect(onTile).toHaveLength(2);
    const firsts = onTile.map((c) => (c[1] as Keyframe[])[0]);
    expect(firsts.some((k) => k.scale !== undefined && k.translate === undefined)).toBe(true);
    expect(firsts.some((k) => k.translate !== undefined && k.scale === undefined)).toBe(true);
    expect(sounds("attach_aura")).toHaveLength(1);
    expect(sounds("token_create")).toHaveLength(0);
  });

  it("two Equipment attached together each slide, with one sound", () => {
    const board = makeBoard(["a", "s1", "s2"]);
    const cards = ["a", "s1", "s2"].map((id) => permanent(id, "opp"));
    const r = render(TableEffects as never, { view: viewWith([], cards), boardEl: board } as never);
    advance(50);
    at(tile(board, "s1"), UNDER);
    at(tile(board, "s2"), { ...UNDER, left: UNDER.left - 10 });
    r.setProps({
      view: viewWith(
        [attachEntry("s1", "a", "Equipment", 10), attachEntry("s2", "a", "Equipment", 10)],
        cards,
      ),
    } as never);
    flushSync();
    expect(effects("attach_equip")).toHaveLength(2);
    expect(sounds("attach_equip")).toHaveLength(1);
  });

  it("an attachment already gone draws nothing, and its sound still plays", () => {
    const board = makeBoard(["a"]);
    const a = permanent("a", "opp");
    const r = render(TableEffects as never, { view: viewWith([], [a]), boardEl: board } as never);
    advance(50);
    r.setProps({ view: viewWith([attachEntry("aura", "a", "Aura", 11)], [a]) } as never);
    flushSync();
    expect(effects()).toHaveLength(0);
    expect(run).not.toHaveBeenCalled();
    expect(sounds("attach_aura")).toHaveLength(1);
  });

  it("a skip puts the attachment at rest at once", () => {
    const cancelled: Element[] = [];
    vi.mocked(run).mockImplementation((el) => ({
      cancel: () => cancelled.push(el),
      done: Promise.resolve(true),
    }));
    try {
      const board = equip([attachEntry("sword", "a", "Equipment", 7)]);
      window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
      flushSync();
      expect(effects()).toHaveLength(0);
      expect(cancelled).toContain(tile(board, "sword"));
    } finally {
      vi.mocked(run).mockImplementation(() => ({ cancel() {}, done: Promise.resolve(true) }));
    }
  });

  it("Short plays the same inside its cap, and still (reduced motion, Off) draws nothing and plays each sound once", () => {
    updateSettings("animations", "effects", "short");
    equip([attachEntry("sword", "a", "Equipment", 7)]);
    expect(effects("attach_equip")).toHaveLength(1);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS + 1e-6);
    cleanup();
    document.body.innerHTML = "";

    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(playSound).mockClear();
      vi.mocked(run).mockClear();
      still();
      equip([attachEntry("sword", "a", "Equipment", 7)]);
      cleanup();
      document.body.innerHTML = "";
      arrive([attachEntry("aura", "a", "Aura", 8)], ["aura"]);
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(sounds("attach_equip")).toHaveLength(1);
      expect(sounds("attach_aura")).toHaveLength(1);
      cleanup();
      document.body.innerHTML = "";
    }
  });
});

describe("cards are milled or discarded into a graveyard (#4445)", () => {
  // The opponent's library and graveyard piles (each a button with its top
  // card's thumbnail) and their hand, where the board draws them.
  const LIB = { left: 1000, top: 560, width: 40, height: 56 };
  const GY = { left: 1060, top: 560, width: 40, height: 56 };
  const HAND = { left: 500, top: 20, width: 200, height: 60 };
  const centre = (r: typeof LIB) => ({ x: r.left + r.width / 2, y: r.top + r.height / 2 });
  const card = (id: string, owner = "opp") => permanent(id, owner, { type_line: "Sorcery" });

  function pile(board: HTMLElement, zone: string, owner: string, r: typeof LIB) {
    const button = rectEl("data-pile", zone, r);
    button.setAttribute("data-pile-owner", owner);
    const thumb = rectEl("class", "thumb", r);
    button.appendChild(thumb);
    board.appendChild(button);
  }

  function tableView(
    log: LogEvent[],
    graveyards: { me?: CardView[]; opp?: CardView[] },
    battlefield: CardView[] = [],
  ): GameView {
    const v = viewWith(log, battlefield);
    for (const [i, id] of ["me", "opp"].entries()) {
      const cards = graveyards[id as "me" | "opp"] ?? [];
      v.seats[i].graveyard = { kind: "graveyard", owner: id, count: cards.length, cards };
    }
    return v;
  }

  // The piles and the hand are on the board; then `log` lands, with the
  // cards it names in their graveyards.
  function frame(log: LogEvent[], graveyards: { me?: CardView[]; opp?: CardView[] }) {
    const board = makeBoard([]);
    pile(board, "library", "opp", LIB);
    pile(board, "graveyard", "opp", GY);
    board.appendChild(rectEl("data-hand-owner", "opp", HAND));
    const r = render(TableEffects as never, { view: tableView([], {}), boardEl: board } as never);
    advance(50);
    r.setProps({ view: tableView(log, graveyards) } as never);
    flushSync();
    return { board, r };
  }

  const sounds = (name: string) => vi.mocked(playSound).mock.calls.filter(([n]) => n === name);
  const runsOn = (el: Element) => vi.mocked(run).mock.calls.filter((c) => c[0] === el);
  const mill = (ids: string[], extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "mill",
    turn: 3,
    seat: 1,
    amount: ids.length,
    card_ids: ids,
    batch: 4,
    text: "",
    ...extra,
  });
  const discard = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "discard",
    turn: 3,
    seat: 1,
    card_id: id,
    old_zone: "hand",
    new_zone: "graveyard",
    cause: "effect",
    batch: 6,
    text: "",
    ...extra,
  });

  it("a mill slides each card off the library onto the graveyard pile, face up, a stagger apart, with one sound", () => {
    const milled = ["m1", "m2", "m3"];
    frame([mill(milled)], { opp: milled.map((id) => card(id)) });
    const cues = effects("mill");
    expect(cues).toHaveLength(1);
    const from = centre(LIB);
    const to = centre(GY);
    expect(cues[0].style.left).toBe(`${from.x}px`);
    expect(cues[0].style.top).toBe(`${from.y}px`);
    expect(cues[0].getAttribute("data-effect-count")).toBe("3");
    expect(cues[0].hasAttribute("data-effect-sweep")).toBe(false);
    const dust = cues[0].querySelector<HTMLElement>(".pile-dust")!;
    expect(dust.style.left).toBe(`${to.x - from.x}px`);
    expect(dust.style.top).toBe(`${to.y - from.y}px`);
    // Each card is its own art, at the graveyard's top card's size, and
    // turns over on the way.
    const copies = [...cues[0].querySelectorAll<HTMLElement>(".pile-card")];
    expect(copies).toHaveLength(3);
    for (const c of copies) {
      expect(c.querySelector(".flip .face img")?.getAttribute("src")).toBeTruthy();
      expect(c.querySelector(".face.back")).toBeTruthy();
      expect(c.querySelector<HTMLElement>(".face")!.style.width).toBe(`${GY.width}px`);
      expect(runsOn(c.querySelector(".flip")!)).toHaveLength(1);
    }
    // One after another: each copy sets off later than the one before.
    const starts = copies.map((c) => runsOn(c)[0][2].delay as number);
    expect(starts[0]).toBe(0);
    expect(starts[1]).toBeGreaterThan(starts[0]);
    expect(starts[2]).toBeGreaterThan(starts[1]);
    // The whole run fits its length, and nothing waits on it.
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(pileBaseMs("mill", 3) + 1e-6);
    expect(sounds("mill")).toEqual([["mill", { volume: 1, batch: "g1:4" }]]);
    advance(EFFECT_CAP_MS + 200);
    expect(sounds("mill")).toHaveLength(1);
    expect(effects()).toHaveLength(0);
  });

  it("a mill of five or more is one sweep of a handful, with a streak", () => {
    const milled = ["a", "b", "c", "d", "e"];
    frame([mill(milled, { amount: 8 })], { opp: milled.map((id) => card(id)) });
    const cues = effects("mill");
    expect(cues).toHaveLength(1);
    expect(cues[0].hasAttribute("data-effect-sweep")).toBe(true);
    expect(cues[0].getAttribute("data-effect-count")).toBe("8");
    expect(cues[0].querySelectorAll(".pile-card")).toHaveLength(3);
    expect(runsOn(cues[0].querySelector(".pile-streak")!)).toHaveLength(1);
    // The handful moves as one.
    const starts = [...cues[0].querySelectorAll(".pile-card")].map(
      (c) => runsOn(c)[0][2].delay as number,
    );
    expect(new Set(starts).size).toBe(1);
    expect(sounds("mill")).toHaveLength(1);
  });

  it("a discard tosses each card from the hand onto the graveyard pile, with one sound for the batch", () => {
    frame([discard("d1"), discard("d2")], { opp: [card("d1"), card("d2")] });
    const cues = effects("discard");
    expect(cues).toHaveLength(1);
    const from = centre(HAND);
    expect(cues[0].style.left).toBe(`${from.x}px`);
    expect(cues[0].style.top).toBe(`${from.y}px`);
    const copies = [...cues[0].querySelectorAll<HTMLElement>(".pile-card")];
    expect(copies).toHaveLength(2);
    // Face up the whole way: a discarded card is shown as it goes.
    for (const c of copies) {
      expect(c.querySelector(".flip")).toBeNull();
      expect(c.querySelector(".face img")?.getAttribute("src")).toBeTruthy();
    }
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(pileBaseMs("discard", 2) + 1e-6);
    expect(sounds("discard")).toEqual([["discard", { volume: 1, batch: "g1:6" }]]);
  });

  it("a discard from the viewer's own hand starts at its card there, not at the pile's thumbnail of it", () => {
    const board = makeBoard([]);
    pile(board, "graveyard", "me", GY);
    // The pile's top card is the discarded card: the board draws it there too.
    board
      .querySelector('[data-pile="graveyard"] .thumb')!
      .appendChild(rectEl("data-instance-id", "mine", GY));
    const hand = rectEl("data-hand-owner", "me", HAND);
    const inHand = { left: 520, top: 600, width: 60, height: 84 };
    hand.appendChild(rectEl("data-instance-id", "mine", inHand));
    board.appendChild(hand);
    const r = render(TableEffects as never, { view: tableView([], {}), boardEl: board } as never);
    advance(50);
    r.setProps({
      view: tableView([discard("mine", { seat: 0, cause: "cleanup" })], {
        me: [card("mine", "me")],
      }),
    } as never);
    flushSync();
    const cues = effects("discard");
    expect(cues).toHaveLength(1);
    expect(cues[0].style.left).toBe(`${centre(inHand).x}px`);
    expect(cues[0].style.top).toBe(`${centre(inHand).y}px`);
  });

  it("each player milling at once is one sound, each into their own graveyard", () => {
    const board = makeBoard([]);
    pile(board, "graveyard", "opp", GY);
    pile(board, "graveyard", "me", { ...GY, left: 100 });
    const r = render(TableEffects as never, { view: tableView([], {}), boardEl: board } as never);
    advance(50);
    r.setProps({
      view: tableView([mill(["o1"], { seat: 1 }), mill(["m1"], { seat: 0 })], {
        opp: [card("o1")],
        me: [card("m1", "me")],
      }),
    } as never);
    flushSync();
    const cues = effects("mill");
    expect(cues).toHaveLength(2);
    // No library pile on this board: each drops onto its pile from above.
    const dropX = cues.map((c) => c.style.left);
    expect(dropX).toEqual([`${centre(GY).x}px`, `${100 + GY.width / 2}px`]);
    expect(sounds("mill")).toHaveLength(1);
  });

  it("a creature dying in the same frame keeps its own cue, and is not a card going to the graveyard without dying", () => {
    const board = makeBoard(IDS);
    pile(board, "library", "opp", LIB);
    pile(board, "graveyard", "opp", GY);
    const r = render(
      TableEffects as never,
      { view: tableView([], {}, cardsAll), boardEl: board } as never,
    );
    advance(50);
    board.querySelector('[data-instance-id="a"]')?.remove();
    r.setProps({
      view: tableView(
        [zone("a", "graveyard", { cause: "sba", batch: 4 }), mill(["m1"])],
        { opp: [permanent("a", "opp"), card("m1")] },
        cardsAll.slice(1),
      ),
    } as never);
    flushSync();
    expect(effects("dies")).toHaveLength(1);
    expect(effects("mill")).toHaveLength(1);
    expect(effects("mill")[0].getAttribute("data-effect-card")).toBe("m1");
    expect(effects("discard")).toHaveLength(0);
  });

  it("a milled commander its owner sends home plays both cues: into the graveyard, then home", () => {
    const cmdr = permanent("cmdr", "opp", { is_commander: true, type_line: "Legendary Creature" });
    const board = makeBoard([]);
    pile(board, "library", "opp", LIB);
    pile(board, "graveyard", "opp", GY);
    const r = render(TableEffects as never, { view: tableView([], {}), boardEl: board } as never);
    advance(50);
    const v = tableView(
      [mill(["cmdr"]), zone("cmdr", "command", { seat: 1, batch: 4 }, "graveyard")],
      {},
    );
    v.seats[1].command = { kind: "command", owner: "opp", count: 1, cards: [cmdr] };
    board.appendChild(
      rectEl("data-instance-id", "cmdr", { left: 900, top: 650, width: 50, height: 70 }),
    );
    r.setProps({ view: v } as never);
    flushSync();
    expect(effects("mill")).toHaveLength(1);
    expect(effects("commander_return")).toHaveLength(1);
    expect(sounds("mill")).toHaveLength(1);
    expect(sounds("commander_return")).toHaveLength(1);
  });

  it("Short plays the same, and still (reduced motion, Off) draws nothing and plays the sound once", () => {
    updateSettings("animations", "effects", "short");
    frame([mill(["m1", "m2"])], { opp: [card("m1"), card("m2")] });
    expect(effects("mill")).toHaveLength(1);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(pileBaseMs("mill", 2) + 1e-6);
    cleanup();
    document.body.innerHTML = "";

    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(playSound).mockClear();
      vi.mocked(run).mockClear();
      still();
      frame([mill(["m1"]), discard("d1", { batch: 4 })], { opp: [card("m1"), card("d1")] });
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(sounds("mill")).toHaveLength(1);
      expect(sounds("discard")).toHaveLength(1);
      cleanup();
      document.body.innerHTML = "";
    }
  });
});

describe("a card returns from a graveyard to the battlefield or a hand (#4446)", () => {
  // Each seat's graveyard pile (a button with its top card's thumbnail)
  // and the opponent's hand, where the board draws them.
  const GY = { left: 1060, top: 560, width: 40, height: 56 };
  const MY_GY = { left: 100, top: 700, width: 40, height: 56 };
  const HAND = { left: 500, top: 20, width: 200, height: 60 };
  const centre = (r: typeof GY) => ({ x: r.left + r.width / 2, y: r.top + r.height / 2 });
  // makeBoard's tiles: the i-th at left 100 + i * 100, top 300.
  const tileCentre = (i: number) => ({
    x: 100 + i * 100 + TILE.width / 2,
    y: 300 + TILE.height / 2,
  });

  function pile(board: HTMLElement, owner: string, r: typeof GY) {
    const button = rectEl("data-pile", "graveyard", r);
    button.setAttribute("data-pile-owner", owner);
    button.appendChild(rectEl("class", "thumb", r));
    board.appendChild(button);
  }

  function tableView(
    log: LogEvent[],
    graveyards: { me?: CardView[]; opp?: CardView[] },
    battlefield: CardView[] = [],
    hands: { me?: CardView[] } = {},
  ): GameView {
    const v = viewWith(log, battlefield);
    for (const [i, id] of ["me", "opp"].entries()) {
      const cards = graveyards[id as "me" | "opp"] ?? [];
      v.seats[i].graveyard = { kind: "graveyard", owner: id, count: cards.length, cards };
    }
    if (hands.me)
      v.seats[0].hand = { kind: "hand", owner: "me", count: hands.me.length, cards: hands.me };
    return v;
  }

  // The board as it stands after the frame (`ids` the battlefield tiles),
  // a first frame with `before` in the graveyards, then `log`.
  function frame(
    ids: string[],
    before: { me?: CardView[]; opp?: CardView[] },
    log: LogEvent[],
    after: {
      battlefield?: CardView[];
      graveyards?: { me?: CardView[]; opp?: CardView[] };
      hands?: { me?: CardView[] };
    } = {},
  ) {
    const board = makeBoard(ids);
    pile(board, "opp", GY);
    pile(board, "me", MY_GY);
    board.appendChild(rectEl("data-hand-owner", "opp", HAND));
    const r = render(
      TableEffects as never,
      { view: tableView([], before), boardEl: board } as never,
    );
    advance(50);
    r.setProps({
      view: tableView(log, after.graveyards ?? {}, after.battlefield ?? [], after.hands ?? {}),
    } as never);
    flushSync();
    return { board, r };
  }

  const sounds = (name: string) => vi.mocked(playSound).mock.calls.filter(([n]) => n === name);
  const runsOn = (el: Element) => vi.mocked(run).mock.calls.filter((c) => c[0] === el);
  const back = (id: string, to: string, extra: Partial<LogEvent> = {}): LogEvent =>
    zone(id, to, { seat: 1, batch: 7, ...extra }, "graveyard");
  const card = (id: string, owner = "opp", extra: Partial<CardView> = {}) =>
    permanent(id, owner, extra);

  it("a reanimated creature rises out of its graveyard pile to its tile, which waits for it, with one sound", () => {
    const { board } = frame(["z"], { opp: [card("z")] }, [back("z", "battlefield")], {
      battlefield: [card("z")],
    });
    const cues = effects("reanimate");
    expect(cues).toHaveLength(1);
    const from = centre(GY);
    const to = tileCentre(0);
    expect(cues[0].style.left).toBe(`${from.x}px`);
    expect(cues[0].style.top).toBe(`${from.y}px`);
    expect(cues[0].getAttribute("data-effect-to")).toBe("battlefield");
    expect(cues[0].getAttribute("data-effect-count")).toBe("1");
    const copy = cues[0].querySelector<HTMLElement>(".pile-card")!;
    expect(copy.querySelector(".face img")?.getAttribute("src")).toBeTruthy();
    expect(copy.querySelector(".rise-aura")).toBeTruthy();
    expect(cues[0].querySelectorAll(".rise-glow")).toHaveLength(1);
    // It goes from the pile to the tile, growing to the tile's size.
    const path = runsOn(copy)[0][1] as Keyframe[];
    expect(path.at(-1)?.translate).toBe(`${to.x - from.x}px ${to.y - from.y}px`);
    expect(path.at(-1)?.scale).toBe(String(TILE.width / GY.width));
    // Up out of the pile first.
    expect(path[1].translate).toMatch(/^0px -\d+px$/);
    // The tile is held hidden until the copy lands, then shown.
    const tile = board.querySelector('[data-instance-id="z"]')!;
    const held = runsOn(tile);
    expect(held).toHaveLength(1);
    expect((held[0][1] as Keyframe[])[0].opacity).toBe(0);
    expect(held[0][2].fill).toBe("backwards");
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(pileBaseMs("reanimate", 1) + 1e-6);
    expect(sounds("reanimate")).toEqual([["reanimate", { volume: 1, batch: "g1:7" }]]);
    advance(EFFECT_CAP_MS + 200);
    expect(sounds("reanimate")).toHaveLength(1);
    expect(effects()).toHaveLength(0);
  });

  it("one effect's cards from two graveyards are one cue, each from its own pile, a stagger apart, one sound", () => {
    // Living Death: the opponent's two creatures and the viewer's one.
    // `seat` on each line is who returned it; the pile is its owner's.
    const ids = ["o1", "o2", "m1"];
    frame(
      ids,
      { opp: [card("o1"), card("o2")], me: [card("m1", "me")] },
      [back("o1", "battlefield"), back("o2", "battlefield"), back("m1", "battlefield")],
      { battlefield: [card("o1"), card("o2"), card("m1", "me")] },
    );
    const cues = effects("reanimate");
    expect(cues).toHaveLength(1);
    expect(cues[0].getAttribute("data-effect-count")).toBe("3");
    const copies = [...cues[0].querySelectorAll<HTMLElement>(".pile-card")];
    expect(copies).toHaveLength(3);
    // The first two rise from the opponent's pile, the third from the viewer's.
    const o = centre(GY);
    const m = centre(MY_GY);
    expect(copies.map((c) => [c.style.left, c.style.top])).toEqual([
      ["0px", "0px"],
      ["0px", "0px"],
      [`${m.x - o.x}px`, `${m.y - o.y}px`],
    ]);
    const starts = copies.map((c) => runsOn(c)[0][2].delay as number);
    expect(starts[1]).toBeGreaterThan(starts[0]);
    expect(starts[2]).toBeGreaterThan(starts[1]);
    expect(sounds("reanimate")).toHaveLength(1);
  });

  it("a creature reanimated under another player's control rises from its owner's graveyard", () => {
    // Reanimate on the opponent's creature: the viewer controls it now.
    frame(["theirs"], { opp: [card("theirs")] }, [back("theirs", "battlefield", { seat: 0 })], {
      battlefield: [card("theirs", "opp", { controller: "me" })],
    });
    const cues = effects("reanimate");
    expect(cues).toHaveLength(1);
    expect(cues[0].style.left).toBe(`${centre(GY).x}px`);
    expect(cues[0].style.top).toBe(`${centre(GY).y}px`);
  });

  it("five or more rise as one sweep of three", () => {
    const ids = ["a", "b", "c", "d", "e"];
    frame(
      ids,
      { opp: ids.map((id) => card(id)) },
      ids.map((id) => back(id, "battlefield")),
      { battlefield: ids.map((id) => card(id)) },
    );
    const cues = effects("reanimate");
    expect(cues).toHaveLength(1);
    expect(cues[0].hasAttribute("data-effect-sweep")).toBe(true);
    expect(cues[0].getAttribute("data-effect-count")).toBe("5");
    const copies = [...cues[0].querySelectorAll(".pile-card")];
    expect(copies).toHaveLength(3);
    expect(new Set(copies.map((c) => runsOn(c)[0][2].delay as number)).size).toBe(1);
    expect(sounds("reanimate")).toHaveLength(1);
  });

  it("a card regrown into an opponent's hand rises with its face from the graveyard, turns over, and is quieter", () => {
    frame([], { opp: [card("g", "opp", { type_line: "Sorcery" })] }, [back("g", "hand")]);
    const cues = effects("reanimate");
    expect(cues).toHaveLength(1);
    expect(cues[0].getAttribute("data-effect-to")).toBe("hand");
    const copy = cues[0].querySelector<HTMLElement>(".pile-card")!;
    const flip = copy.querySelector<HTMLElement>('.flip[data-flip="down"]')!;
    expect(flip.querySelector(".face img")?.getAttribute("src")).toBeTruthy();
    expect(flip.querySelector(".face.back")).toBeTruthy();
    expect(runsOn(flip)).toHaveLength(1);
    const path = runsOn(copy)[0][1] as Keyframe[];
    const from = centre(GY);
    const to = centre(HAND);
    expect(path.at(-1)?.translate).toBe(`${to.x - from.x}px ${to.y - from.y}px`);
    expect(sounds("reanimate")).toEqual([
      ["reanimate", { volume: REANIMATE_HAND_VOLUME, batch: "g1:7" }],
    ]);
  });

  it("a card regrown into the viewer's own hand lands on its card there, face up", () => {
    const board = makeBoard([]);
    pile(board, "me", MY_GY);
    const hand = rectEl("data-hand-owner", "me", HAND);
    const inHand = { left: 520, top: 600, width: 60, height: 84 };
    hand.appendChild(rectEl("data-instance-id", "mine", inHand));
    board.appendChild(hand);
    const mine = card("mine", "me", { type_line: "Sorcery" });
    const r = render(
      TableEffects as never,
      { view: tableView([], { me: [mine] }), boardEl: board } as never,
    );
    advance(50);
    r.setProps({
      view: tableView([back("mine", "hand", { seat: 0 })], {}, [], { me: [mine] }),
    } as never);
    flushSync();
    const cues = effects("reanimate");
    expect(cues).toHaveLength(1);
    expect(cues[0].style.left).toBe(`${centre(MY_GY).x}px`);
    const copy = cues[0].querySelector<HTMLElement>(".pile-card")!;
    expect(copy.querySelector(".flip")).toBeNull();
    expect(copy.querySelector(".face img")?.getAttribute("src")).toBeTruthy();
    const path = runsOn(copy)[0][1] as Keyframe[];
    const from = centre(MY_GY);
    const to = centre(inHand);
    expect(path.at(-1)?.translate).toBe(`${to.x - from.x}px ${to.y - from.y}px`);
  });

  it("a reanimated planeswalker rises with the sound, and its arrival follows silently", async () => {
    const pw = card("z", "opp", { type_line: "Legendary Planeswalker — Liliana" });
    frame(["z"], { opp: [pw] }, [back("z", "battlefield")], { battlefield: [pw] });
    expect(effects("reanimate")).toHaveLength(1);
    // Its arrival waits for the rise to mount, then for its tile to settle.
    await vi.advanceTimersByTimeAsync(0);
    flushSync();
    await vi.advanceTimersByTimeAsync(0);
    flushSync();
    expect(effects("pw_enter")).toHaveLength(1);
    expect(sounds("reanimate")).toHaveLength(1);
    expect(sounds("pw_enter")).toHaveLength(0);
  });

  it("a token copy of a card in a graveyard is a token, not a reanimation", () => {
    const token: LogEvent = {
      seq: ++seq,
      kind: "token",
      turn: 3,
      seat: 1,
      card_id: "copy",
      batch: 7,
      text: "",
    };
    frame(["copy"], { opp: [card("orig")] }, [token], {
      battlefield: [card("copy", "opp", { is_token: true })],
      graveyards: { opp: [card("orig")] },
    });
    expect(effects("reanimate")).toHaveLength(0);
    expect(sounds("reanimate")).toHaveLength(0);
    expect(effects("token_create")).toHaveLength(1);
  });

  it("a commander going home from its graveyard is commander_return only", () => {
    const cmdr = card("cmdr", "opp", { is_commander: true, type_line: "Legendary Creature" });
    const board = makeBoard([]);
    pile(board, "opp", GY);
    board.appendChild(
      rectEl("data-instance-id", "cmdr", { left: 900, top: 650, width: 50, height: 70 }),
    );
    const r = render(
      TableEffects as never,
      { view: tableView([], { opp: [cmdr] }), boardEl: board } as never,
    );
    advance(50);
    const v = tableView([back("cmdr", "command")], {});
    v.seats[1].command = { kind: "command", owner: "opp", count: 1, cards: [cmdr] };
    r.setProps({ view: v } as never);
    flushSync();
    expect(effects("commander_return")).toHaveLength(1);
    expect(effects("reanimate")).toHaveLength(0);
    expect(sounds("reanimate")).toHaveLength(0);
  });

  it("Short plays within its cap, and still (reduced motion, Off) draws nothing and plays the sound once", () => {
    updateSettings("animations", "effects", "short");
    const ids = ["a", "b", "c", "d"];
    frame(
      ids,
      { opp: ids.map((id) => card(id)) },
      ids.map((id) => back(id, "battlefield")),
      {
        battlefield: ids.map((id) => card(id)),
      },
    );
    expect(effects("reanimate")).toHaveLength(1);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS + 1e-6);
    cleanup();
    document.body.innerHTML = "";

    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(playSound).mockClear();
      vi.mocked(run).mockClear();
      still();
      frame(["z"], { opp: [card("z")] }, [back("z", "battlefield")], { battlefield: [card("z")] });
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(sounds("reanimate")).toHaveLength(1);
      cleanup();
      document.body.innerHTML = "";
    }
  });
});

describe("a player searches a library (#4447)", () => {
  // Each seat's library pile (a button with its top card's back).
  const LIB = { left: 1000, top: 560, width: 40, height: 56 };
  const MY_LIB = { left: 100, top: 700, width: 40, height: 56 };
  const centre = (r: typeof LIB) => ({ x: r.left + r.width / 2, y: r.top + r.height / 2 });

  function pile(board: HTMLElement, owner: string, r: typeof LIB) {
    const button = rectEl("data-pile", "library", r);
    button.setAttribute("data-pile-owner", owner);
    button.appendChild(rectEl("class", "thumb", r));
    board.appendChild(button);
  }

  // Both library piles are on the board; then `log` lands.
  function frame(log: LogEvent[], cards: CardView[] = []) {
    const board = makeBoard([]);
    pile(board, "opp", LIB);
    pile(board, "me", MY_LIB);
    const r = render(TableEffects as never, { view: viewWith([], []), boardEl: board } as never);
    advance(50);
    r.setProps({ view: viewWith(log, cards) } as never);
    flushSync();
    return { board, r };
  }

  const sounds = (name: string) => vi.mocked(playSound).mock.calls.filter(([n]) => n === name);
  const runsOn = (el: Element) => vi.mocked(run).mock.calls.filter((c) => c[0] === el);
  const search = (extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "search",
    turn: 3,
    seat: 1,
    target_seat: 1,
    library_seat: 1,
    label: "search_shuffled",
    batch: 4,
    text: "P2 searched their library and shuffled it",
    ...extra,
  });

  it("a search riffles card backs off the searched library's pile, one after another, with one sound", () => {
    frame([search()]);
    const cues = effects("library_search");
    expect(cues).toHaveLength(1);
    const at = centre(LIB);
    expect(cues[0].style.left).toBe(`${at.x}px`);
    expect(cues[0].style.top).toBe(`${at.y}px`);
    expect(cues[0].getAttribute("data-effect-library")).toBe("opp");
    const backs = [...cues[0].querySelectorAll<HTMLElement>(".riffle-card")];
    expect(backs.length).toBeGreaterThanOrEqual(3);
    for (const c of backs) {
      // Backs only, at the pile's top card's size: no face, no art.
      const face = c.querySelector<HTMLElement>(".face")!;
      expect(face.classList.contains("back-only")).toBe(true);
      expect(c.querySelector("img")).toBeNull();
      expect(face.style.width).toBe(`${LIB.width}px`);
      expect(face.style.height).toBe(`${LIB.height}px`);
    }
    // One after another: each sets off later than the one before.
    const starts = backs.map((c) => runsOn(c)[0][2].delay as number);
    expect(starts[0]).toBe(0);
    for (let i = 1; i < starts.length; i++) expect(starts[i]).toBeGreaterThan(starts[i - 1]);
    expect(runsOn(cues[0].querySelector(".riffle-glow")!)).toHaveLength(1);
    // About half a second, and nothing waits on it.
    for (const ms of runEnds())
      expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.library_search + 1e-6);
    expect(sounds("library_search")).toEqual([["library_search", { volume: 1, batch: "g1:4" }]]);
    advance(EFFECT_CAP_MS + 200);
    expect(sounds("library_search")).toHaveLength(1);
    expect(effects()).toHaveLength(0);
  });

  it("a fetch land's reveal riffles the searcher's library, and the found land's arrival is its own", () => {
    // "P1 fetched Forest from their library onto the battlefield".
    const fetched: LogEvent = {
      seq: ++seq,
      kind: "reveal",
      turn: 3,
      seat: 0,
      library_seat: 0,
      old_zone: "library",
      amount: 1,
      batch: 7,
      text: "P1 fetched Forest from their library onto the battlefield",
    };
    frame([fetched, zone("forest", "battlefield", { seat: 0, batch: 7 }, "library")]);
    const cues = effects("library_search");
    expect(cues).toHaveLength(1);
    expect(cues[0].getAttribute("data-effect-library")).toBe("me");
    expect(cues[0].style.left).toBe(`${centre(MY_LIB).x}px`);
    expect(sounds("library_search")).toEqual([["library_search", { volume: 1, batch: "g1:7" }]]);
  });

  it("another player's library riffles that pile; each library searched at once is one sound", () => {
    frame([
      search({ seat: 0, library_seat: 1, batch: 9 }),
      search({ seat: 0, target_seat: 0, library_seat: 0, batch: 9 }),
    ]);
    const cues = effects("library_search");
    expect(cues.map((c) => c.getAttribute("data-effect-library"))).toEqual(["opp", "me"]);
    expect(sounds("library_search")).toHaveLength(1);
  });

  it("a plain shuffle, or a reveal no search made, riffles nothing and plays nothing", () => {
    frame([
      search({ label: "shuffle", library_seat: undefined, batch: 11 }),
      {
        seq: ++seq,
        kind: "reveal",
        turn: 3,
        seat: 1,
        old_zone: "library",
        amount: 1,
        batch: 12,
        text: "",
      },
    ]);
    expect(effects()).toHaveLength(0);
    expect(sounds("library_search")).toHaveLength(0);
  });

  it("with no library pile on the board it riffles over the owner's seat at a card's size", () => {
    const board = makeBoard([]);
    const r = render(TableEffects as never, { view: viewWith([], []), boardEl: board } as never);
    advance(50);
    r.setProps({ view: viewWith([search()], []) } as never);
    flushSync();
    const cues = effects("library_search");
    expect(cues).toHaveLength(1);
    // makeBoard's opponent avatar: 60 x 60 at (560, 20).
    expect(cues[0].style.left).toBe("590px");
    expect(cues[0].style.top).toBe("50px");
  });

  it("Short plays the same, and still (reduced motion, Off) draws nothing and plays the sound once", () => {
    updateSettings("animations", "effects", "short");
    frame([search()]);
    expect(effects("library_search")).toHaveLength(1);
    for (const ms of runEnds())
      expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.library_search + 1e-6);
    cleanup();
    document.body.innerHTML = "";

    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(playSound).mockClear();
      vi.mocked(run).mockClear();
      still();
      frame([search({ batch: 20 })]);
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(sounds("library_search")).toHaveLength(1);
      cleanup();
      document.body.innerHTML = "";
    }
  });
});

describe("a double-faced card transforms, or day and night changes (#4448)", () => {
  const sounds = (name: string) => vi.mocked(playSound).mock.calls.filter(([n]) => n === name);
  const runsOn = (el: Element) => vi.mocked(run).mock.calls.filter((c) => c[0] === el);
  // The art a tile shows: cardImage.ts's URL, face 0 with no face.
  const art = (id: string, face: number) =>
    `/cards/sf-${id}/image?size=art_crop${face > 0 ? `&face=${face}` : ""}`;
  const dfc = (id: string, face: number, extra: Partial<CardView> = {}) =>
    permanent(id, "opp", {
      layout: "transform",
      active_face: face,
      faces: [
        { name: `${id} front`, image: `/cards/sf-${id}/image?face=0` },
        { name: `${id} back`, image: `/cards/sf-${id}/image?face=1` },
      ],
      ...extra,
    } as Partial<CardView>);
  const transform = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "transform",
    turn: 3,
    seat: 1,
    card_id: id,
    label: `${id} front`,
    batch: 4,
    text: `${id} front transformed into ${id} back`,
    ...extra,
  });
  const night = (extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "day_night",
    turn: 4,
    seat: -1,
    label: "night",
    batch: 9,
    text: "It becomes night",
    ...extra,
  });
  const tileOf = (board: HTMLElement, id: string) =>
    board.querySelector<HTMLElement>(`[data-instance-id="${id}"]`)!;
  const showFace = (board: HTMLElement, id: string, face: number) =>
    tileOf(board, id).querySelector("img")!.setAttribute("src", art(id, face));

  // `ids` are on the board front face up; then `log` lands with each card
  // in `flipped` back face up, its tile showing that face.
  function frame(ids: string[], flipped: string[], log: LogEvent[], tapped: string[] = []) {
    const board = makeBoard(ids);
    for (const id of ids) showFace(board, id, 0);
    for (const id of tapped) {
      const el = tileOf(board, id);
      const r = el.getBoundingClientRect();
      el.dataset.tapped = "true";
      el.getBoundingClientRect = () => new DOMRect(r.left, r.top, r.height, r.width);
    }
    const before = ids.map((id) => dfc(id, 0));
    const r = render(
      TableEffects as never,
      { view: viewWith([], before), boardEl: board } as never,
    );
    advance(50);
    for (const id of flipped) showFace(board, id, 1);
    const after = ids.map((id) => dfc(id, flipped.includes(id) ? 1 : 0));
    r.setProps({ view: viewWith(log, after) } as never);
    flushSync();
    return { board, r };
  }

  it("a transform flips the card over its tile from the face it was to the face it is, with one sound", () => {
    const { board } = frame(["a", "b"], ["a"], [transform("a")]);
    const cues = effects("transform");
    expect(cues).toHaveLength(1);
    expect(cues[0].getAttribute("data-effect-card")).toBe("a");
    // Centred on its tile, at its size.
    expect(cues[0].style.left).toBe(`${100 + TILE.width / 2}px`);
    expect(cues[0].style.top).toBe(`${300 + TILE.height / 2}px`);
    const from = cues[0].querySelector<HTMLElement>(".turn-from")!;
    const to = cues[0].querySelector<HTMLElement>(".turn-to")!;
    expect(from.querySelector("img")?.getAttribute("src")).toBe(art("a", 0));
    expect(to.querySelector("img")?.getAttribute("src")).toBe(art("a", 1));
    expect(from.style.width).toBe(`${TILE.width}px`);
    expect(from.style.height).toBe(`${TILE.height}px`);
    // It turns about its vertical axis, half a turn, edge-on midway.
    const turn = runsOn(cues[0].querySelector(".turn-card")!);
    expect(turn).toHaveLength(1);
    const frames = turn[0][1] as Keyframe[];
    expect(frames[0].transform).toMatch(/rotateY\(0deg\)/);
    expect(frames.find((k) => /rotateY\(90deg\)/.test(String(k.transform)))).toBeTruthy();
    expect(frames.at(-1)?.transform).toMatch(/rotateY\(180deg\)/);
    expect(runsOn(cues[0].querySelector(".turn-rim")!)).toHaveLength(1);
    // Its tile waits hidden under the copy, and only its tile.
    const held = runsOn(tileOf(board, "a"));
    expect(held).toHaveLength(1);
    expect((held[0][1] as Keyframe[])[0].opacity).toBe(0);
    expect((held[0][1] as Keyframe[]).at(-1)?.opacity).toBe(1);
    expect(held[0][2].fill).toBe("backwards");
    expect(runsOn(tileOf(board, "b"))).toHaveLength(0);
    // About 0.6 s, and nothing waits on it.
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.transform + 1e-6);
    expect(sounds("transform")).toEqual([["transform", { volume: 1, batch: "g1:4" }]]);
    advance(EFFECT_CAP_MS + 200);
    expect(sounds("transform")).toHaveLength(1);
    expect(effects()).toHaveLength(0);
  });

  it("a tapped card flips about its own long axis, at its unturned size", () => {
    frame(["a"], ["a"], [transform("a")], ["a"]);
    const cue = effects("transform")[0];
    expect(cue.querySelector<HTMLElement>(".turn")!.style.rotate).toBe("90deg");
    const from = cue.querySelector<HTMLElement>(".turn-from")!;
    expect(from.style.width).toBe(`${TILE.width}px`);
    expect(from.style.height).toBe(`${TILE.height}px`);
  });

  it("night falling washes the table and turns every werewolf a beat later, as one sound, the bell", () => {
    frame(
      ["a", "b", "c"],
      ["a", "b"],
      [night(), transform("a", { batch: 9 }), transform("b", { batch: 9 })],
    );
    const sky = effects("day_night");
    expect(sky).toHaveLength(1);
    expect(sky[0].getAttribute("data-sky")).toBe("night");
    for (const part of [".sky-tint", ".sky-band", ".sky-orb"]) {
      expect(runsOn(sky[0].querySelector(part)!)).toHaveLength(1);
    }
    const flips = effects("transform");
    expect(flips.map((c) => c.getAttribute("data-effect-card"))).toEqual(["a", "b"]);
    for (const c of flips) {
      const turn = runsOn(c.querySelector(".turn-card")!)[0];
      expect(turn[2].delay).toBe(DAY_NIGHT_FLIP_DELAY_MS);
    }
    // One moment: the bell once, and the flips play nothing of their own.
    expect(sounds("day_night")).toEqual([["day_night", { volume: 1, batch: "g1:9" }]]);
    expect(sounds("transform")).toHaveLength(0);
    // Inside the cap.
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_CAP_MS);
    advance(EFFECT_CAP_MS + 200);
    expect(effects()).toHaveLength(0);
  });

  it("day breaking with no werewolf out is the wash alone, in the day's colours", () => {
    frame(["a"], [], [night({ label: "day", text: "It becomes day", batch: 10 })]);
    const sky = effects("day_night");
    expect(sky).toHaveLength(1);
    expect(sky[0].getAttribute("data-sky")).toBe("day");
    expect(effects("transform")).toHaveLength(0);
    expect(sounds("day_night")).toHaveLength(1);
  });

  it("a melded permanent flips from its first card's front to the melded face", () => {
    const board = makeBoard(["host"]);
    const r = render(TableEffects as never, { view: viewWith([], []), boardEl: board } as never);
    advance(50);
    showFace(board, "host", 0);
    const host = permanent("host", "opp", {
      layout: "meld",
      melded_from: [
        { name: "Midnight Scavengers", image: "/cards/sf-scav/image" },
        { name: "Graf Rats", image: "/cards/sf-rats/image" },
      ],
    } as Partial<CardView>);
    const meld: LogEvent = {
      seq: ++seq,
      kind: "meld",
      turn: 3,
      seat: 1,
      card_id: "host",
      label: "Midnight Scavengers and Graf Rats",
      batch: 12,
      text: "Midnight Scavengers and Graf Rats melded into Chittering Host",
    };
    r.setProps({
      view: viewWith([zone("host", "battlefield", { batch: 12 }, "exile"), meld], [host]),
    } as never);
    flushSync();
    const cues = effects("transform");
    expect(cues).toHaveLength(1);
    expect(cues[0].querySelector(".turn-from img")?.getAttribute("src")).toBe(
      "/cards/sf-scav/image?size=art_crop",
    );
    expect(cues[0].querySelector(".turn-to img")?.getAttribute("src")).toBe(art("host", 0));
    expect(sounds("transform")).toHaveLength(1);
  });

  it("a card no longer on the board, or one with no other face, draws nothing and still plays the sound", () => {
    frame(["a"], [], [transform("gone", { batch: 14 })]);
    expect(effects()).toHaveLength(0);
    expect(sounds("transform")).toHaveLength(1);
  });

  it("Short plays the same, and still (reduced motion, Off) draws nothing, holds no tile and plays the sound once", () => {
    updateSettings("animations", "effects", "short");
    frame(["a"], ["a"], [night(), transform("a", { batch: 9 })]);
    expect(effects("transform")).toHaveLength(1);
    expect(effects("day_night")).toHaveLength(1);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS + 1e-6);
    cleanup();
    document.body.innerHTML = "";

    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(playSound).mockClear();
      vi.mocked(run).mockClear();
      still();
      frame(["a"], ["a"], [night({ batch: 20 }), transform("a", { batch: 20 })]);
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(sounds("day_night")).toHaveLength(1);
      expect(sounds("transform")).toHaveLength(0);
      cleanup();
      document.body.innerHTML = "";
    }
  });
});

describe("a permanent is sacrificed for its own ability (#4427)", () => {
  const runsOn = (el: Element) => vi.mocked(run).mock.calls.filter((c) => c[0] === el);
  const self = (id: string, batch = 4): LogEvent => ({
    seq: ++seq,
    kind: "sacrifice",
    seat: 1,
    card_id: id,
    cause: "cost",
    source: id,
    batch,
    text: "",
  });
  // a is a fetch land, b and c Treasures, d a creature.
  const types: Partial<CardView>[] = [
    { type_line: "Land" },
    { type_line: "Token Artifact — Treasure" },
    { type_line: "Token Artifact — Treasure" },
    { type_line: "Creature — Snake Shaman" },
  ];
  const crackSounds = () =>
    vi.mocked(playSound).mock.calls.filter((c) => String(c[0]).endsWith("_crack"));

  it("a fetch land splits along a crack where it stood and falls away, with no altar and its own sound", () => {
    deliver([self("a")], ["a"], types);
    const [cue] = effects("crack");
    expect(cue).toBeTruthy();
    expect(effects("sacrifice")).toHaveLength(0);
    expect(cue.getAttribute("data-effect-card")).toBe("a");
    expect(cue.getAttribute("data-crack")).toBe("land");
    // a's tile centre: left 100 + 40, top 300 + 56.
    expect(cue.style.left).toBe("140px");
    expect(cue.style.top).toBe("356px");
    // Two halves of its art, either side of one crack.
    const halves = [...cue.querySelectorAll<HTMLElement>(".crack-half")];
    expect(halves).toHaveLength(2);
    for (const h of halves) {
      expect(h.querySelector("img")?.getAttribute("src")).toBe("/a.jpg");
      expect(h.querySelector<HTMLElement>(".face")!.style.clipPath).toMatch(/^polygon\(/);
      expect(runsOn(h)).toHaveLength(1);
    }
    expect(cue.querySelectorAll(".crack-line path")).toHaveLength(2);
    expect(cue.querySelector(".crack-puff")).toBeTruthy();
    expect(cue.querySelector(".altar")).toBeNull();
    expect(cue.querySelector(".knife")).toBeNull();
    expect(cue.querySelector(".shatter")).toBeNull();
    // Its sound, a little quieter, capped on its payment; never the
    // sacrifice's or its knife's.
    expect(playSound).toHaveBeenCalledTimes(1);
    const [name, opts] = vi.mocked(playSound).mock.calls[0];
    expect(name).toBe("fetch_crack");
    expect(opts?.volume).toBeLessThan(1);
    expect(String(opts?.batch)).toMatch(/^g1:4:crack:/);
    advance(EFFECT_CAP_MS + 200);
    expect(playSound).not.toHaveBeenCalledWith("sacrifice", expect.anything());
    expect(playSound).not.toHaveBeenCalledWith("sacrifice_knife", expect.anything());
    // About 0.6 s, and gone by itself.
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.crack + 1e-6);
    expect(effects()).toHaveLength(0);
  });

  it("two Treasures spent on one spell each crack, with one sound", () => {
    const cast: LogEvent = { seq: 0, kind: "cast", seat: 1, card_id: "spell", batch: 6, text: "" };
    const pay = [self("b", 6), self("c", 6)];
    cast.seq = ++seq;
    deliver([...pay, cast], ["b", "c"], types);
    const cues = effects("crack");
    expect(cues.map((c) => c.getAttribute("data-effect-card"))).toEqual(["b", "c"]);
    expect(cues.map((c) => c.getAttribute("data-crack"))).toEqual(["treasure", "treasure"]);
    expect(crackSounds().map((c) => c[0])).toEqual(["treasure_crack"]);
  });

  it("a creature sacrificed for its own ability keeps the altar and the knife", () => {
    deliver([self("d")], ["d"], types);
    expect(effects("crack")).toHaveLength(0);
    const [cue] = effects("sacrifice");
    expect(cue.querySelector(".altar")).toBeTruthy();
    expect(playSound).toHaveBeenCalledWith("sacrifice", expect.anything());
    expect(crackSounds()).toHaveLength(0);
  });

  it("Short plays the same, and still (reduced motion, Off) draws nothing and plays the sound once", () => {
    updateSettings("animations", "effects", "short");
    deliver([self("a")], ["a"], types);
    expect(effects("crack")).toHaveLength(1);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.crack + 1e-6);
    cleanup();
    document.body.innerHTML = "";
    for (const still of [
      () => updateSettings("accessibility", "reduceMotion", true),
      () => updateSettings("animations", "effects", "off"),
    ]) {
      resetSettings();
      updateSettings("animations", "enabled", true);
      vi.mocked(run).mockClear();
      vi.mocked(playSound).mockClear();
      still();
      deliver([self("b", 7), self("c", 7)], ["b", "c"], types);
      expect(run).not.toHaveBeenCalled();
      expect(effects()).toHaveLength(0);
      expect(badges()).toHaveLength(0);
      expect(crackSounds().map((c) => c[0])).toEqual(["treasure_crack"]);
      cleanup();
      document.body.innerHTML = "";
    }
  });
});

describe("reduced motion and the Off setting", () => {
  // Each way, the same thing: a badge says what happened, no motion
  // starts, and the sound still plays.
  function expectStill(label: string, kind = "destroy") {
    expect(run).not.toHaveBeenCalled();
    expect(crumble).not.toHaveBeenCalled();
    expect(layer().querySelector(".copy")).toBeNull();
    expect(layer().querySelector("img")).toBeNull();
    const [badge] = badges();
    expect(badge?.textContent).toBe(label);
    expect(badge?.getAttribute("data-effect")).toBe(kind);
    expect(badge?.getAttribute("data-effect-motion")).toBe("still");
    expect(playSound).toHaveBeenCalledWith(kind, expect.anything());
  }
  const destroy = () => deliver([zone("b", "graveyard", { cause: "destroy", batch: 4 })], ["b"]);

  it("accessibility.reduceMotion: the badge shows and the motion does not", () => {
    updateSettings("accessibility", "reduceMotion", true);
    destroy();
    expectStill(L.effectBadgeDestroyed);
    // The badge sits where the card was.
    expect(badges()[0].style.left).toBe("240px");
  });

  it("animations.enabled off: the badge shows and the motion does not", () => {
    updateSettings("animations", "enabled", false);
    destroy();
    expectStill(L.effectBadgeDestroyed);
  });

  it("the OS signal: the badge shows and the motion does not", () => {
    vi.stubGlobal(
      "matchMedia",
      vi.fn(() => ({ matches: true, addEventListener() {}, removeEventListener() {} })),
    );
    window.matchMedia = globalThis.matchMedia;
    destroy();
    expectStill(L.effectBadgeDestroyed);
  });

  it("animations.effects Off: the badge shows and the motion does not", () => {
    updateSettings("animations", "effects", "off");
    destroy();
    expectStill(L.effectBadgeDestroyed);
  });

  it("an exile is labelled as one, never as a death", () => {
    updateSettings("accessibility", "reduceMotion", true);
    deliver([zone("a", "exile", { cause: "destroy", batch: 5 })], ["a"]);
    expectStill(L.effectBadgeExiled, "exile");
  });

  it("a board wipe is one badge and a static tint, and its cards add none", () => {
    updateSettings("accessibility", "reduceMotion", true);
    const wipe = IDS.map((id) =>
      zone(id, "graveyard", { cause: "destroy", batch: 9, source: "damnation" }),
    );
    deliver(wipe, IDS);
    expect(run).not.toHaveBeenCalled();
    expect(badges().map((b) => b.textContent)).toEqual([L.effectBadgeWipe(4)]);
    expect(layer().querySelector(".still-tint")).toBeTruthy();
    expect(layer().querySelector(".sweep")).toBeNull();
    expect(playSound).toHaveBeenCalledWith("board_wipe", expect.anything());
  });

  it("a counterspell adds no badge of its own: the stack linger's Countered badge says it", () => {
    updateSettings("accessibility", "reduceMotion", true);
    const c: LogEvent = {
      seq: ++seq,
      kind: "counter",
      seat: 1,
      target: "bolt",
      stack_item_id: "item-1",
      text: "",
    };
    deliver([c]);
    expect(effects()).toHaveLength(0);
    expect(playSound).toHaveBeenCalledWith("counter", expect.anything());
  });

  it("the badge holds for its reading time, not scaled by speed, and then goes", () => {
    updateSettings("accessibility", "reduceMotion", true);
    updateSettings("animations", "speed", 0.5);
    destroy();
    advance(EFFECT_STILL_HOLD_MS - 100);
    expect(badges()).toHaveLength(1);
    advance(300);
    expect(badges()).toHaveLength(0);
  });
});

describe("skip", () => {
  it("Escape clears every cue playing, and the next frame's cue still plays", () => {
    const { r } = deliver([zone("a", "graveyard", { cause: "destroy" })], ["a"]);
    expect(effects()).toHaveLength(1);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(effects()).toHaveLength(0);
    // Nothing was lost but the motion: a later frame cues as usual.
    r.setProps({
      view: viewWith(
        [
          zone("a", "graveyard", { cause: "destroy" }),
          zone("b", "exile", { cause: "exile_effect" }),
        ],
        [],
      ),
    } as never);
    flushSync();
    expect(effects("exile")).toHaveLength(1);
  });

  it("a click clears them too, and does not stop the click", () => {
    deliver([zone("a", "graveyard", { cause: "destroy" })], ["a"]);
    expect(effects()).toHaveLength(1);
    const seen = vi.fn();
    document.body.addEventListener("pointerdown", seen);
    document.body.dispatchEvent(new Event("pointerdown", { bubbles: true }));
    flushSync();
    expect(effects()).toHaveLength(0);
    expect(seen).toHaveBeenCalled();
  });

  it("a key that is not Escape does nothing", () => {
    deliver([zone("a", "graveyard", { cause: "destroy" })], ["a"]);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "a", bubbles: true }));
    flushSync();
    expect(effects()).toHaveLength(1);
  });

  it("clears a pending staggered cue's timers, so nothing draws after the skip", () => {
    const wipe = IDS.map((id) =>
      zone(id, "graveyard", { cause: "destroy", batch: 9, source: "damnation" }),
    );
    deliver(wipe, IDS);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    advance(EFFECT_CAP_MS);
    expect(effects()).toHaveLength(0);
  });
});

describe("priming", () => {
  it("a prime-key change cues nothing already on the wire", () => {
    const board = makeBoard(IDS);
    const r = render(
      TableEffects as never,
      { view: viewWith([], cardsAll), boardEl: board, beatsPrimeKey: "live" } as never,
    );
    advance(50);
    r.setProps({
      view: viewWith([zone("a", "graveyard", { cause: "destroy" })], cardsAll.slice(1)),
      beatsPrimeKey: "offline",
    } as never);
    flushSync();
    expect(effects()).toHaveLength(0);
    expect(playSound).not.toHaveBeenCalled();
  });
});

describe("a drain, the monarch and the initiative (#4450)", () => {
  const ME = { left: 300, top: 700, width: 200, height: 80 };
  const OPP = { left: 560, top: 20, width: 60, height: 60 };
  const sounds = (name: string) => vi.mocked(playSound).mock.calls.filter(([n]) => n === name);

  // The board (with the opponent's seat), plus the viewer's own.
  function frame(log: LogEvent[]) {
    const board = makeBoard(IDS);
    board.appendChild(rectEl("data-seat-id", "me", ME));
    const before = cardsAll.map((c) => ({ ...c }) as CardView);
    const r = render(
      TableEffects as never,
      { view: viewWith([], before), boardEl: board } as never,
    );
    advance(50);
    r.setProps({ view: viewWith(log, before) } as never);
    flushSync();
    return r;
  }
  const loss = (seat: number): LogEvent => ({
    seq: ++seq,
    kind: "life",
    seat,
    card_id: "a",
    amount: -1,
    batch: 9,
    text: "",
  });
  const crown = (kind: "monarch" | "initiative", seat: number): LogEvent => ({
    seq: ++seq,
    kind,
    seat,
    batch: 10,
    text: "",
  });

  it("a drain draws one thin line from the source to each player and plays one sound", () => {
    frame([loss(0), loss(1)]);
    const [cue] = effects("drain_table");
    expect(cue).toBeTruthy();
    const lines = [...cue.querySelectorAll<SVGLineElement>(".drain-line")];
    expect(lines).toHaveLength(2);
    // The source tile "a" is centred at (140, 356).
    for (const l of lines) {
      expect(l.getAttribute("x1")).toBe("140");
      expect(l.getAttribute("y1")).toBe("356");
    }
    expect(lines.map((l) => l.getAttribute("x2"))).toEqual(["400", "590"]);
    expect(sounds("drain_table")).toEqual([["drain_table", { volume: 1, batch: "g1:9" }]]);
    for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS.drain_table + 1e-6);
    advance(EFFECT_BASE_MS.drain_table + 200);
    expect(effects()).toHaveLength(0);
  });

  it.each(["monarch", "initiative"] as const)(
    "%s: a glint on the new holder's panel and one sound",
    (kind) => {
      frame([crown(kind, 1)]);
      const [cue] = effects(kind);
      expect(cue).toBeTruthy();
      expect(cue.style.left).toBe(`${OPP.left + OPP.width / 2}px`);
      expect(cue.style.top).toBe(`${OPP.top + OPP.height / 2}px`);
      expect(cue.querySelector(".glint-ring")).toBeTruthy();
      expect(cue.querySelectorAll(".glint-spark").length).toBeGreaterThan(0);
      expect(sounds(kind)).toEqual([[kind, { volume: 1, batch: "g1:10" }]]);
      for (const ms of runEnds()) expect(ms).toBeLessThanOrEqual(EFFECT_BASE_MS[kind] + 1e-6);
    },
  );

  it("Short plays the same cue, speed scales it", () => {
    updateSettings("animations", "effects", "short");
    updateSettings("animations", "speed", 2);
    frame([crown("monarch", 0)]);
    expect(effects("monarch")).toHaveLength(1);
    expect(Math.max(...runEnds())).toBeLessThanOrEqual(EFFECT_BASE_MS.monarch * 2 + 1e-6);
  });

  it.each([
    ["reduced motion", () => updateSettings("accessibility", "reduceMotion", true)],
    ["Off", () => updateSettings("animations", "effects", "off")],
  ])("still (%s): nothing is drawn, and each sound plays once", (_, still) => {
    still();
    frame([loss(0), loss(1), crown("monarch", 1), crown("initiative", 0)]);
    expect(effects()).toHaveLength(0);
    expect(run).not.toHaveBeenCalled();
    expect(sounds("drain_table")).toHaveLength(1);
    expect(sounds("monarch")).toHaveLength(1);
    expect(sounds("initiative")).toHaveLength(1);
  });
});
