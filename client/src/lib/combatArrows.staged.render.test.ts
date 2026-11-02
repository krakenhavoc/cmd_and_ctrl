// @vitest-environment jsdom
//
// #4327: an attacker the viewer has staged (#3266) shows its line before
// the declaration is sent, dashed in the attack colour like a blocker
// still being collected for a group (#3898). Once the server has it as
// attacking, the solid arrow replaces the dashed one.

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
  Element.prototype.getBoundingClientRect = function () {
    return { left: 100, top: 50, width: 40, height: 60, right: 140, bottom: 110 } as DOMRect;
  };
});

afterEach(() => {
  cleanup();
  proto.animate = hadAnimate;
  svg.getTotalLength = hadLength as never;
  Element.prototype.getBoundingClientRect = realRect;
  document.body.innerHTML = "";
});

const creature = (id: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name: id,
    owner: "me",
    controller: "me",
    type_line: "Creature — Wolf",
    ...extra,
  }) as unknown as CardView;

function mountArrows(cards: CardView[], staged: { attacker: string; target: string }[]) {
  const board = document.createElement("div");
  board.innerHTML =
    cards.map((c) => `<div data-instance-id="${c.instance_id}"></div>`).join("") +
    '<div data-seat-id="bob"></div><div data-instance-id="walker"></div>';
  document.body.appendChild(board);
  const v = dockTestView({});
  v.battlefield = {
    kind: "battlefield",
    count: cards.length,
    cards: [...cards, creature("walker", { controller: "bob", type_line: "Planeswalker" })],
  } as GameView["battlefield"];
  return render(
    CombatArrows as never,
    { view: v, boardEl: board, cues: new CombatCues(), staged } as never,
  );
}

async function settle(): Promise<void> {
  await new Promise((res) => setTimeout(res, 40));
  flushSync();
}

describe("staged attackers' lines (#4327)", () => {
  it("draws a dashed attack line from each staged attacker to its player or planeswalker", async () => {
    const { container } = mountArrows(
      [creature("w1"), creature("w2")],
      [
        { attacker: "w1", target: "bob" },
        { attacker: "w2", target: "walker" },
      ],
    );
    await settle();
    const lines = container.querySelectorAll("path.attack-pending");
    expect(lines).toHaveLength(2);
    expect(lines[0]!.getAttribute("marker-end")).toBe("url(#arrowhead-attack)");
    expect(container.querySelectorAll("path.attack")).toHaveLength(0);
  });

  it("draws the solid arrow, not the dashed one, once the creature is attacking", async () => {
    const { container } = mountArrows(
      [creature("w1", { attacking_target: "bob" })],
      [{ attacker: "w1", target: "bob" }],
    );
    await settle();
    expect(container.querySelectorAll("path.attack-pending")).toHaveLength(0);
    expect(container.querySelectorAll("path.attack")).toHaveLength(1);
  });

  it("draws nothing for a staged creature that has left the battlefield", async () => {
    const { container } = mountArrows([], [{ attacker: "w1", target: "bob" }]);
    await settle();
    expect(container.querySelectorAll("path")).toHaveLength(0);
  });
});
