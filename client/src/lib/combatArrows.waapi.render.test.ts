// @vitest-environment jsdom
//
// #3076: CombatArrows draws an arrow on with a Web Animation (the
// dash offset walking from the path's length to 0) and cancels it when
// the arrow goes. jsdom has no Web Animations, so Element.prototype.animate
// is stubbed and the call is pinned.

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { flushSync } from "svelte";

import CombatArrows from "./components/board/CombatArrows.svelte";
import { CombatCues } from "./combatCues.svelte";
import { EASE } from "./motion";
import type { CardView, GameView } from "./protocol";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";

interface Call {
  el: Element;
  keyframes: Keyframe[];
  options: KeyframeAnimationOptions;
  cancelled: boolean;
}
let calls: Call[] = [];
const proto = Element.prototype as unknown as { animate?: unknown };
const svg = SVGElement.prototype as unknown as { getTotalLength?: () => number };
const realRect = Element.prototype.getBoundingClientRect;
let hadAnimate: unknown;
let hadLength: unknown;

class FakeObserver {
  observe(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  calls = [];
  hadAnimate = proto.animate;
  hadLength = svg.getTotalLength;
  proto.animate = function (
    this: Element,
    keyframes: Keyframe[],
    options: KeyframeAnimationOptions,
  ) {
    const call: Call = { el: this, keyframes, options, cancelled: false };
    calls.push(call);
    return {
      onfinish: null,
      oncancel: null,
      cancel() {
        call.cancelled = true;
      },
    };
  };
  svg.getTotalLength = () => 123;
  (globalThis as Record<string, unknown>).ResizeObserver ??= FakeObserver;
  Element.prototype.getBoundingClientRect = function (this: Element) {
    const left = this.getAttribute("data-instance-id") ? 100 : 400;
    return { left, top: 50, width: 40, height: 60, right: left + 40, bottom: 110 } as DOMRect;
  };
});

afterEach(() => {
  cleanup();
  proto.animate = hadAnimate;
  svg.getTotalLength = hadLength as never;
  Element.prototype.getBoundingClientRect = realRect;
  document.body.innerHTML = "";
});

const attacker = {
  instance_id: "ogre",
  name: "Ogre",
  owner: "me",
  controller: "me",
  known_by_you: true,
  type_line: "Creature — Ogre",
  attacking_target: "opp",
} as unknown as CardView;

function mountArrows() {
  const board = document.createElement("div");
  board.innerHTML = '<div data-instance-id="ogre"></div><div data-seat-id="opp"></div>';
  document.body.appendChild(board);
  const v = dockTestView({});
  v.battlefield = { kind: "battlefield", count: 1, cards: [attacker] } as GameView["battlefield"];
  const r = render(
    CombatArrows as never,
    {
      view: v,
      boardEl: board,
      cues: new CombatCues(),
    } as never,
  );
  return { ...r, board, view: v };
}

describe("CombatArrows draw-on", () => {
  it("walks the dash offset from the path length to 0 in 340 ms", async () => {
    const { container } = mountArrows();
    await new Promise((res) => setTimeout(res, 40));
    flushSync();
    const path = container.querySelector<SVGPathElement>("path.attack");
    expect(path).not.toBeNull();
    const draw = calls.find((c) => c.el === path);
    expect(draw).toBeDefined();
    expect(draw!.keyframes).toEqual([{ strokeDashoffset: 123 }, { strokeDashoffset: 0 }]);
    expect(draw!.options).toMatchObject({ duration: 340, easing: EASE.power2Out });
    // Left drawn: the animation is the only thing that ever hides it.
    expect(path!.style.strokeDasharray).toBe("123");
    expect(path!.style.strokeDashoffset).toBe("0");
  });

  it("cancels the draw when the arrow goes", async () => {
    const m = mountArrows();
    await new Promise((res) => setTimeout(res, 40));
    flushSync();
    expect(calls.length).toBeGreaterThan(0);
    m.destroy();
    for (const c of calls) expect(c.cancelled).toBe(true);
  });
});
