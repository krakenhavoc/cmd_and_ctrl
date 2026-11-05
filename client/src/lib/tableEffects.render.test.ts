// @vitest-environment jsdom
//
// ADR 0209 §1 and §3 (#4396, #4416), the rendered half: a battlefield
// exit, a counterspell, a planeswalker, a created token (#4441), counters
// put on a permanent (#4442), a commander cast from the command zone or
// going home (#4443), an Equipment attached or an Aura enchanting a
// permanent (#4444) and a board wipe each draw their cue
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
  FIZZLE_MOTES,
  TOKEN_STAGGER_MS,
  WIPE_FIRST_CARD_MS,
  busyScale,
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
