// @vitest-environment jsdom
//
// #4065: a seat the reflow draws as an avatar card has no card tiles on
// the table, so an arrow to or from one of its cards ends on the avatar
// card (`data-seat-stand-in`, boardAnchor.findStandIn) instead of being
// dropped: a blocker there still shows whom it blocks.

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { flushSync } from "svelte";

import CombatArrows from "./components/board/CombatArrows.svelte";
import { CombatCues } from "./combatCues.svelte";
import type { CardView, GameView } from "./protocol";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";

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
  hadAnimate = proto.animate;
  hadLength = svg.getTotalLength;
  proto.animate = () => ({ onfinish: null, oncancel: null, cancel() {} });
  svg.getTotalLength = () => 100;
  (globalThis as Record<string, unknown>).ResizeObserver ??= FakeObserver;
  // Every element has a box: where it is does not matter here, only
  // whether the arrow found both ends.
  Element.prototype.getBoundingClientRect = function (this: Element) {
    const left = this.hasAttribute("data-seat-stand-in") ? 900 : 100;
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

// Bob's ogre attacks Cat; Cat's bear blocks it. Cat is an avatar card.
const ogre = {
  instance_id: "ogre",
  name: "Ogre",
  owner: "bob",
  controller: "bob",
  type_line: "Creature — Ogre",
  attacking_target: "cat",
} as unknown as CardView;
const bear = {
  instance_id: "bear",
  name: "Grizzly Bears",
  owner: "cat",
  controller: "cat",
  type_line: "Creature — Bear",
  blocking_target: "ogre",
} as unknown as CardView;

function mountArrows(boardHTML: string) {
  const board = document.createElement("div");
  board.innerHTML = boardHTML;
  document.body.appendChild(board);
  const v = dockTestView({});
  v.battlefield = { kind: "battlefield", count: 2, cards: [ogre, bear] } as GameView["battlefield"];
  return render(
    CombatArrows as never,
    { view: v, boardEl: board, cues: new CombatCues() } as never,
  );
}

async function settle(): Promise<void> {
  await new Promise((res) => setTimeout(res, 40));
  flushSync();
}

describe("arrows to a seat drawn as an avatar card", () => {
  it("draw the block from the avatar card, and the attack to its avatar", async () => {
    const { container } = mountArrows(
      '<div data-instance-id="ogre"></div>' +
        '<div data-seat-stand-in="cat"><div data-seat-id="cat"></div></div>',
    );
    await settle();
    expect(container.querySelectorAll("path.attack")).toHaveLength(1);
    expect(container.querySelectorAll("path.block")).toHaveLength(1);
  });

  it("draw no block when the seat has neither its tiles nor an avatar card", async () => {
    const { container } = mountArrows(
      '<div data-instance-id="ogre"></div><div data-seat-id="cat"></div>',
    );
    await settle();
    expect(container.querySelectorAll("path.attack")).toHaveLength(1);
    expect(container.querySelectorAll("path.block")).toHaveLength(0);
  });
});
