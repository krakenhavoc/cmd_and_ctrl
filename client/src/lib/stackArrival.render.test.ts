// @vitest-environment jsdom
//
// ADR 0119 amendment 2026-10-16 §D — the arrival trace on a board: a new
// item draws one from its source; the first frame, a reconnect and each
// of its gates draw none; a start that is not on screen falls back to
// the caster's avatar, and with neither there is no trace. The rules are
// pinned in stackArrival.test.ts; this pins the wiring.

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { get } from "svelte/store";

import StackArrival from "./components/board/StackArrival.svelte";
import { settings, updateSettings } from "./settings";
import type { GameView, StackItemView } from "./protocol";
import { cleanup, render } from "./test/render.svelte";

const realRect = Element.prototype.getBoundingClientRect;

function rect(left: number, top: number, width: number, height: number): DOMRect {
  return {
    left,
    top,
    width,
    height,
    right: left + width,
    bottom: top + height,
    x: left,
    y: top,
    toJSON: () => ({}),
  } as DOMRect;
}

let board: HTMLElement;
let laneRect: DOMRect;
const saved = get(settings);

beforeEach(() => {
  Element.prototype.getBoundingClientRect = function (this: Element) {
    if (this.classList.contains("board")) return rect(0, 0, 1400, 900);
    if (this.hasAttribute("data-seat-id")) return rect(1200, 40, 60, 60);
    if (this.hasAttribute("data-instance-id")) {
      // "far" is off the board: not on screen.
      return this.getAttribute("data-instance-id") === "far"
        ? rect(2000, 2000, 80, 110)
        : rect(900, 600, 80, 110);
    }
    if (this.hasAttribute("data-stack-item-id")) return rect(20, 300, 200, 280);
    if (this.classList.contains("stack-lane")) return laneRect;
    return rect(0, 0, 0, 0);
  };
  laneRect = rect(10, 250, 230, 360);
  board = document.createElement("div");
  board.className = "board";
  board.innerHTML = `
    <div data-seat-id="opp"></div>
    <div data-instance-id="purphoros"></div>
    <div data-instance-id="far"></div>
    <div class="stack-lane">
      <div data-stack-item-id="bolt"></div>
      <div data-stack-item-id="t1"></div>
      <div data-stack-item-id="t2"></div>
    </div>`;
  document.body.appendChild(board);
  updateSettings("animations", "enabled", true);
  updateSettings("animations", "stackArrival", true);
  updateSettings("accessibility", "reduceMotion", false);
});

afterEach(() => {
  cleanup();
  Element.prototype.getBoundingClientRect = realRect;
  updateSettings("animations", "enabled", saved.animations.enabled);
  updateSettings("animations", "stackArrival", saved.animations.stackArrival);
  updateSettings("accessibility", "reduceMotion", saved.accessibility.reduceMotion);
});

const spell = (id: string): StackItemView =>
  ({ id, kind: "spell", controller: "opp", owner: "opp", source_card_id: id }) as StackItemView;
const trig = (id: string, source: string): StackItemView =>
  ({
    id,
    kind: "triggered",
    controller: "opp",
    owner: "opp",
    source_card_id: source,
  }) as StackItemView;

function view(items: StackItemView[], seq: number): GameView {
  return {
    seats: [
      { id: "me", seat: 0 },
      { id: "opp", seat: 1 },
    ],
    stack_items: items,
    log: [{ seq, kind: "step", seat: 0, text: "" }],
  } as unknown as GameView;
}

function mount(first: GameView) {
  const r = render(
    StackArrival as never,
    { view: first, boardEl: board, beatsPrimeKey: "a" } as never,
  );
  const traces = () => Array.from(r.container.querySelectorAll<SVGGElement>("g.trace"));
  return { ...r, traces };
}

describe("StackArrival", () => {
  it("traces nothing on the first frame, then a new spell from its caster's avatar", () => {
    const r = mount(view([spell("bolt")], 1));
    expect(r.traces()).toHaveLength(0);
    r.setProps({ view: view([spell("bolt"), trig("t1", "purphoros")], 2) } as never);
    const t = r.traces();
    expect(t).toHaveLength(1);
    // The caster's seat colour, and a head plus its tail.
    expect(t[0].style.getPropertyValue("--seat")).not.toBe("");
    expect(t[0].querySelectorAll("path.seg.head")).toHaveLength(1);
    expect(t[0].querySelectorAll("path.seg").length).toBeGreaterThan(2);
    // The ability leaves from its source permanent, at (900,600)+: the
    // path starts there, not at the avatar.
    const d = t[0].querySelector("path")!.getAttribute("d")!;
    const [x, y] = d.split(" ").slice(1, 3).map(Number);
    expect(x).toBeGreaterThanOrEqual(900);
    expect(x).toBeLessThanOrEqual(980);
    expect(y).toBeGreaterThanOrEqual(600);
    // The pile's arrows wait for it.
    expect(board.style.getPropertyValue("--stack-arrival-hold")).toBe("320ms");
  });

  it("falls back to the avatar when the source is off screen", () => {
    const r = mount(view([], 1));
    r.setProps({ view: view([trig("t1", "far")], 2) } as never);
    const d = r.traces()[0].querySelector("path")!.getAttribute("d")!;
    const [x] = d.split(" ").slice(1, 3).map(Number);
    expect(x).toBeGreaterThanOrEqual(1200);
  });

  it("traces nothing from an avatar the stack panel covers", () => {
    // The pile docked over the caster's avatar, as on a two-seat table.
    laneRect = rect(1000, 0, 400, 400);
    const r = mount(view([], 1));
    r.setProps({ view: view([spell("bolt")], 2) } as never);
    expect(r.traces()).toHaveLength(0);
  });

  it("traces nothing for an item the stack does not draw", () => {
    const r = mount(view([], 1));
    r.setProps({ view: view([spell("unseen")], 2) } as never);
    expect(r.traces()).toHaveLength(0);
  });

  it("traces nothing on a reconnect", () => {
    const r = mount(view([spell("bolt")], 1));
    r.setProps({
      beatsPrimeKey: "b",
      view: view([spell("bolt"), trig("t1", "purphoros")], 5),
    } as never);
    expect(r.traces()).toHaveLength(0);
  });

  it("traces nothing with animations off, its toggle off, or reduced motion", () => {
    for (const [group, key, value] of [
      ["animations", "enabled", false],
      ["animations", "stackArrival", false],
      ["accessibility", "reduceMotion", true],
    ] as const) {
      updateSettings(group, key as never, value as never);
      const r = mount(view([spell("bolt")], 1));
      r.setProps({ view: view([spell("bolt"), trig("t1", "purphoros")], 2) } as never);
      expect(r.traces()).toHaveLength(0);
      r.destroy();
      updateSettings(group, key as never, !value as never);
    }
  });
});
