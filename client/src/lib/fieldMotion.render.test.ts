// @vitest-environment jsdom
//
// #3601: the arrival ring is measured while the card's ETB pop still has
// it scaled down. The rect a ring takes must frame the whole card once the
// pop has settled, not the 0.55 box it had mid-pop.

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { flushSync } from "svelte";

import FieldMotion from "./components/board/FieldMotion.svelte";
import { render, cleanup } from "./test/render.svelte";
import { settled } from "./motion";
import type { CardView, GameView } from "./protocol";

const ME = "me";
const OPP = "opp";

const realRect = Element.prototype.getBoundingClientRect;
const doc = document as unknown as { getAnimations?: () => unknown[] };
const realGetAnimations = doc.getAnimations;

const rect = (left: number, top: number, width: number, height: number) =>
  ({
    left,
    top,
    width,
    height,
    right: left + width,
    bottom: top + height,
    x: left,
    y: top,
    toJSON: () => ({}),
  }) as DOMRect;

// The tile is drawn at 200x280 at (100, 100). Inside a popping listitem
// (data-popping) it is drawn scaled down to 0.55 about its centre.
const FULL = rect(100, 100, 200, 280);
const POPPED = rect(100 + 100 * 0.45, 100 + 140 * 0.45, 110, 154);

interface FakeAnimation {
  effect: { target: Element; getComputedTiming: () => { endTime: number } };
  finished: Promise<unknown>;
}

let boardEl: HTMLDivElement;
let pop: HTMLDivElement;
let animation: FakeAnimation;
let finish: () => void;

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}
const g = globalThis as Record<string, unknown>;

function card(id: string, owner: string): CardView {
  return {
    instance_id: id,
    name: "Falkenrath Noble",
    owner,
    controller: owner,
    known_by_you: true,
    scryfall_id: `sf-${id}`,
  } as CardView;
}

function gameView(cards: CardView[], logLength: number): GameView {
  const zone = (kind: string, cs: CardView[] = []) => ({ kind, count: cs.length, cards: cs });
  return {
    id: "g1",
    state: "active",
    seats: [
      { id: ME, name: "Me", seat: 0, life: 40 },
      { id: OPP, name: "Bot 2", seat: 1, life: 40 },
    ],
    battlefield: zone("battlefield", cards),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    turn: { number: 3, active_seat: 1, priority_holder: 0, phase: "main1", step: "main" },
    log: Array.from({ length: logLength }, (_, i) => ({
      seq: i + 1,
      kind: "note",
      seat: 1,
      text: `event ${i + 1}`,
    })),
  } as unknown as GameView;
}

async function flushMicrotasks(): Promise<void> {
  for (let i = 0; i < 10; i++) await Promise.resolve();
  flushSync();
}

beforeEach(() => {
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  Element.prototype.getBoundingClientRect = function (this: Element) {
    if (this.classList.contains("board")) return rect(0, 0, 1400, 900);
    if (this.closest("[data-popping]")) return POPPED;
    if (this.hasAttribute("data-instance-id")) return FULL;
    return rect(0, 0, 0, 0);
  };

  boardEl = document.createElement("div");
  boardEl.className = "board";
  pop = document.createElement("div");
  pop.setAttribute("role", "listitem");
  pop.setAttribute("data-popping", "");
  const tile = document.createElement("div");
  tile.dataset.instanceId = "falk";
  pop.appendChild(tile);
  boardEl.appendChild(pop);
  document.body.appendChild(boardEl);

  // One finite entrance animation on the listitem, as etbPulse starts.
  let resolve!: () => void;
  const finished = new Promise<void>((r) => (resolve = r));
  finish = () => {
    pop.removeAttribute("data-popping");
    resolve();
  };
  animation = {
    effect: { target: pop, getComputedTiming: () => ({ endTime: 320 }) },
    finished,
  };
  doc.getAnimations = () => [animation];
});

afterEach(() => {
  cleanup();
  boardEl.remove();
  Element.prototype.getBoundingClientRect = realRect;
  doc.getAnimations = realGetAnimations;
});

describe("the arrival ring after an entrance pop (#3601)", () => {
  it("frames the whole card once the pop has settled, not the mid-pop box", async () => {
    const r = render(FieldMotion, {
      view: gameView([], 0),
      viewerID: ME,
      boardEl,
      beatsPrimeKey: "live:false",
    });
    await flushMicrotasks();

    // The permanent enters the opponent's battlefield while it is popping.
    r.setProps({ view: gameView([card("falk", OPP)], 1) });
    await flushMicrotasks();

    const ring = r.container.querySelector<HTMLElement>("[data-field-arrival='falk']");
    expect(ring).not.toBeNull();
    expect(ring!.style.width).toBe("110px");
    expect(ring!.style.height).toBe("154px");

    finish();
    await flushMicrotasks();

    const settledRing = r.container.querySelector<HTMLElement>("[data-field-arrival='falk']");
    expect(settledRing!.style.left).toBe("100px");
    expect(settledRing!.style.top).toBe("100px");
    expect(settledRing!.style.width).toBe("200px");
    expect(settledRing!.style.height).toBe("280px");
  });
});

describe("settled", () => {
  it("resolves at once when nothing animates the element or an ancestor", async () => {
    animation.effect.target = document.createElement("div");
    const tile = pop.querySelector("[data-instance-id]")!;
    let done = false;
    void settled(tile).then(() => (done = true));
    await flushMicrotasks();
    expect(done).toBe(true);
  });

  it("waits for a finite animation on an ancestor and ignores endless ones", async () => {
    const tile = pop.querySelector("[data-instance-id]")!;
    const endless: FakeAnimation = {
      effect: { target: pop, getComputedTiming: () => ({ endTime: Infinity }) },
      finished: new Promise(() => {}),
    };
    doc.getAnimations = () => [endless, animation];
    let done = false;
    void settled(tile).then(() => (done = true));
    await flushMicrotasks();
    expect(done).toBe(false);
    finish();
    await flushMicrotasks();
    expect(done).toBe(true);
  });
});
