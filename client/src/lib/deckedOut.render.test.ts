// @vitest-environment jsdom
//
// #4879, the rendered half: a player decked out (cause empty_draw) plays
// player_decked and no rock-shatter (EliminationFx), the library's ghost is
// drawn over their own mat with their hand flung from it (TableEffects),
// and their stone reads "Decked" over a drained seat (Tombstone). Still
// and reduced motion draw the stone alone. The rules are deckedOut.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: vi.fn() }));
vi.mock("./motion", async (importOriginal) => {
  const real = await importOriginal<typeof import("./motion")>();
  return {
    ...real,
    run: vi.fn(() => ({ cancel() {}, done: Promise.resolve(true) })),
  };
});

import EliminationFx from "./components/board/EliminationFx.svelte";
import TableEffects from "./components/board/TableEffects.svelte";
import Tombstone from "./components/board/Tombstone.svelte";
import { play as playSound } from "./sounds";
import { run } from "./motion";
import { L } from "./labels";
import type { CardView, GameView, LogEvent, PlayerView } from "./protocol";
import { resetSettings, updateSettings } from "./settings";
import { DECKED_BEATS, busyScale } from "./tableEffects";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

const handCard = (id: string, owner: string): CardView =>
  ({
    instance_id: id,
    name: id,
    owner,
    controller: owner,
    known_by_you: true,
    type_line: "Instant",
    scryfall_id: `sf-${id}`,
  }) as unknown as CardView;

let seq = 100;
const outE = (seat: number, cause: string): LogEvent =>
  ({ seq: ++seq, kind: "eliminated", seat, cause, turn: 3, text: "" }) as LogEvent;

// Two seats, "me" (0) and "opp" (1). `out` maps a seat to its loss cause.
function view(log: LogEvent[], out: Record<number, string> = {}, hands = true): GameView {
  const v = dockTestView({ log } as Partial<GameView>);
  v.seats = v.seats.map((s, n) => {
    const p = { ...s } as PlayerView;
    if (out[n]) {
      p.eliminated = true;
      p.loss_cause = out[n];
      p.hand = { kind: "hand", owner: p.id, count: 0, cards: [] } as PlayerView["hand"];
    } else if (hands) {
      // The viewer sees their own hand's faces; the opponent's is a count.
      p.hand =
        n === 0
          ? ({
              kind: "hand",
              owner: p.id,
              count: 3,
              cards: [handCard("h1", p.id), handCard("h2", p.id), handCard("h3", p.id)],
            } as PlayerView["hand"])
          : ({ kind: "hand", owner: p.id, count: 4, cards: [] } as PlayerView["hand"]);
    }
    return p;
  });
  return v;
}

function rect(el: HTMLElement, r: { left: number; top: number; width: number; height: number }) {
  el.getBoundingClientRect = () => new DOMRect(r.left, r.top, r.width, r.height);
  return el;
}

// A 1200 x 800 board: the opponent's mat on top, the viewer's below, each
// with its library pile and its hand area.
function makeBoard(): HTMLElement {
  const board = rect(document.createElement("div"), { left: 0, top: 0, width: 1200, height: 800 });
  const mats: [string, number][] = [
    ["opp", 0],
    ["me", 400],
  ];
  for (const [id, top] of mats) {
    const panel = rect(document.createElement("div"), { left: 0, top, width: 1200, height: 400 });
    panel.className = "panel seat-panel";
    const seatEl = rect(document.createElement("div"), {
      left: 20,
      top: top + 20,
      width: 60,
      height: 60,
    });
    seatEl.setAttribute("data-seat-id", id);
    const pile = rect(document.createElement("button"), {
      left: 1100,
      top: top + 250,
      width: 60,
      height: 84,
    });
    pile.setAttribute("data-pile", "library");
    pile.setAttribute("data-pile-owner", id);
    const hand = rect(document.createElement("div"), {
      left: 300,
      top: top + 300,
      width: 600,
      height: 100,
    });
    hand.setAttribute("data-hand-owner", id);
    panel.append(seatEl, pile, hand);
    board.appendChild(panel);
  }
  document.body.appendChild(board);
  return board;
}

function advance(ms: number) {
  flushSync();
  vi.advanceTimersByTime(ms);
  flushSync();
}

function stubOSReduceMotion(on: boolean) {
  vi.stubGlobal("matchMedia", (query: string) => ({
    matches: on && query.includes("prefers-reduced-motion"),
    media: query,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    onchange: null,
    dispatchEvent: () => false,
  }));
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
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
  vi.mocked(playSound).mockClear();
  vi.mocked(run).mockClear();
  stubOSReduceMotion(false);
});

afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

const sounds = () => vi.mocked(playSound).mock.calls.map(([n]) => n);
const ghosts = () => [...document.querySelectorAll<HTMLElement>("[data-decked-out]")];

describe("the elimination layer, for a decked player", () => {
  function deliver(cause: string) {
    const board = makeBoard();
    const r = render(EliminationFx as never, { view: view([]), boardEl: board } as never);
    flushSync();
    r.setProps({ view: view([outE(1, cause)], { 1: cause }) } as never);
    flushSync();
  }

  it("plays player_decked, once, and no shatter", () => {
    deliver("empty_draw");
    expect(sounds()).toEqual(["player_decked"]);
    expect(document.querySelectorAll("[data-burst]")).toHaveLength(0);
  });

  it("keeps player_eliminated and the shatter for every other cause", () => {
    for (const cause of ["life", "commander_damage", "poison", "concede", "effect"]) {
      cleanup();
      document.body.innerHTML = "";
      vi.mocked(playSound).mockClear();
      deliver(cause);
      expect(sounds()).toEqual(["player_eliminated"]);
      expect(document.querySelectorAll("[data-burst]")).toHaveLength(1);
    }
  });
});

describe("the library's ghost", () => {
  function deliver(cause: string, seat = 1) {
    const board = makeBoard();
    const r = render(TableEffects as never, { view: view([]), boardEl: board } as never);
    advance(50);
    r.setProps({ view: view([outE(seat, cause)], { [seat]: cause }) } as never);
    flushSync();
    return { r, board };
  }

  it("is drawn over the decked player's own mat, clipped to it", () => {
    deliver("empty_draw");
    const [g] = ghosts();
    expect(g).toBeTruthy();
    expect(g.dataset.deckedOut).toBe("opp");
    expect(g.getAttribute("class")).toContain("decked");
    const mat = g.querySelector<HTMLElement>(".dk-mat")!;
    // The opponent's panel: the top half of the board.
    expect([mat.style.left, mat.style.top, mat.style.width, mat.style.height]).toEqual([
      "0px",
      "0px",
      "1200px",
      "400px",
    ]);
    // The skull, the pulse and the wisps are inside the mat's clip.
    expect(mat.querySelector(".dk-skull svg")).toBeTruthy();
    const pulse = mat.querySelector<HTMLElement>(".dk-pulse")!;
    expect([pulse.style.left, pulse.style.top]).toEqual(["1100px", "250px"]);
    expect(mat.querySelectorAll(".dk-wisp").length).toBeGreaterThan(4);
    // The skull is the mat's height at full size.
    const skull = mat.querySelector<HTMLElement>(".dk-skull")!;
    expect(parseFloat(skull.style.height)).toBeCloseTo(400 * 0.96);
    // It takes no pointer events, and the layer stays hidden from AT.
    expect(document.querySelector("[data-table-effects]")?.getAttribute("aria-hidden")).toBe(
      "true",
    );
  });

  it("flings an opponent's hand as card backs, from their hand area", () => {
    deliver("empty_draw");
    const cards = [...ghosts()[0].querySelectorAll<HTMLImageElement>(".dk-card")];
    expect(cards).toHaveLength(4);
    for (const c of cards) expect(c.getAttribute("src")).toBe("/card-back-small.jpg");
    // Outside the mat's clip, so they can leave the board.
    expect(cards[0].closest(".dk-mat")).toBeNull();
    // Their hand area's middle row.
    expect(cards[0].style.top).toBe("350px");
  });

  it("flings the player's own hand face up", () => {
    deliver("empty_draw", 0);
    const cards = [...ghosts()[0].querySelectorAll<HTMLImageElement>(".dk-card")];
    expect(cards).toHaveLength(3);
    for (const c of cards) expect(c.getAttribute("src")).not.toBe("/card-back-small.jpg");
  });

  it("runs about 2.2 s at Full and flings the hand as the skull fills the mat", () => {
    deliver("empty_draw");
    const calls = vi.mocked(run).mock.calls;
    const ends = calls.map((c) => ((c[2].delay as number) ?? 0) + (c[2].duration as number));
    expect(Math.max(...ends)).toBeCloseTo(2200, 0);
    const flung = calls.filter(([el]) => (el as HTMLElement).classList.contains("dk-card"));
    expect(flung).toHaveLength(4);
    for (const [, , o] of flung) expect(o.delay).toBeCloseTo(2200 * DECKED_BEATS.fill);
    advance(2400);
    expect(ghosts()).toHaveLength(0);
  });

  it("is half as long in Short", () => {
    updateSettings("animations", "effects", "short");
    deliver("empty_draw");
    const ends = vi
      .mocked(run)
      .mock.calls.map((c) => ((c[2].delay as number) ?? 0) + (c[2].duration as number));
    expect(Math.max(...ends)).toBeLessThanOrEqual(1100);
    expect(Math.max(...ends)).toBeGreaterThan(900);
  });

  it("is compressed on a busy table", () => {
    const board = makeBoard();
    const r = render(TableEffects as never, { view: view([]), boardEl: board } as never);
    advance(50);
    // Both players decked in one frame: two headlines at once.
    r.setProps({
      view: view([outE(0, "empty_draw"), outE(1, "empty_draw")], {
        0: "empty_draw",
        1: "empty_draw",
      }),
    } as never);
    flushSync();
    expect(ghosts()).toHaveLength(2);
    const ends = vi
      .mocked(run)
      .mock.calls.map((c) => ((c[2].delay as number) ?? 0) + (c[2].duration as number));
    expect(Math.max(...ends)).toBeCloseTo(2200 * busyScale(2), 0);
  });

  it("is skipped by a click or Escape", () => {
    deliver("empty_draw");
    expect(ghosts()).toHaveLength(1);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(ghosts()).toHaveLength(0);
  });

  it("is not drawn for any other cause", () => {
    for (const cause of ["life", "commander_damage", "poison", "concede", "effect"]) {
      cleanup();
      document.body.innerHTML = "";
      deliver(cause);
      expect(ghosts()).toHaveLength(0);
    }
  });

  it("draws nothing still: reduced motion, the OS signal, or effects Off", () => {
    updateSettings("accessibility", "reduceMotion", true);
    deliver("empty_draw");
    expect(ghosts()).toHaveLength(0);
    cleanup();
    document.body.innerHTML = "";
    updateSettings("accessibility", "reduceMotion", false);
    stubOSReduceMotion(true);
    deliver("empty_draw");
    expect(ghosts()).toHaveLength(0);
    cleanup();
    document.body.innerHTML = "";
    stubOSReduceMotion(false);
    updateSettings("animations", "effects", "off");
    deliver("empty_draw");
    expect(ghosts()).toHaveLength(0);
  });
});

describe("the decked stone", () => {
  const mount = (v: GameView, n: number, compact = false) =>
    render(Tombstone as never, { seat: v.seats[n], view: v, compact } as never);
  const stone = () => document.querySelector<HTMLElement>("[data-tombstone]");

  it("reads Decked, from the label, over a drained seat", () => {
    mount(view([], { 1: "empty_draw" }), 1);
    const el = stone()!;
    expect(el.dataset.decked).toBe("true");
    expect(el.querySelector(".rip")?.textContent).toBe(L.tombstoneDecked);
    expect(L.tombstoneDecked).toBe("Decked");
    expect(el.querySelector(".drain")).toBeTruthy();
  });

  it("keeps R. I. P. and no drain for every other cause", () => {
    mount(view([], { 1: "life" }), 1);
    const el = stone()!;
    expect(el.dataset.decked).toBeUndefined();
    expect(el.querySelector(".rip")?.textContent).toBe(L.tombstoneRip);
    expect(el.querySelector(".drain")).toBeNull();
  });

  it("is the grey seat and the badge alone under reduced motion", () => {
    updateSettings("accessibility", "reduceMotion", true);
    const before = view([]);
    const r = mount(before, 1);
    flushSync();
    const after = view([outE(1, "empty_draw")], { 1: "empty_draw" });
    r.setProps({ seat: after.seats[1], view: after } as never);
    flushSync();
    const el = stone()!;
    expect(el.dataset.motion).toBe("still");
    expect(el.style.getPropertyValue("--tomb-delay")).toBe("0ms");
    expect(el.querySelector(".drain")).toBeTruthy();
    expect(el.querySelector(".rip")?.textContent).toBe(L.tombstoneDecked);
  });

  it("settles at the ghost's badge beat when it arrives with motion", () => {
    const before = view([]);
    const r = mount(before, 1);
    flushSync();
    const after = view([outE(1, "empty_draw")], { 1: "empty_draw" });
    r.setProps({ seat: after.seats[1], view: after } as never);
    flushSync();
    expect(stone()!.style.getPropertyValue("--tomb-delay")).toBe(
      `${Math.round(2200 * DECKED_BEATS.badge)}ms`,
    );
  });
});
