// @vitest-environment jsdom
//
// #4114 — the badge a divided effect puts on each target. The plan is
// pure; measureShareBadges runs against a hand-built DOM with stubbed
// rects, since jsdom does no layout.

import { describe, expect, it } from "vitest";

import type { CardView, PlayerView, StackItemView, ZoneView } from "./protocol";
import { buildStackLane } from "./stackLane";
import { measureShareBadges, planShareBadges } from "./stackShares";
import { TOP_ARROW_COLOR } from "./stackArrows";

const zone = (kind: string, cards: CardView[] = []): ZoneView =>
  ({ kind, count: cards.length, cards }) as ZoneView;
const seat = (id: string, name: string, n: number): PlayerView =>
  ({ id, name, seat: n, life: 40, graveyard: zone("graveyard") }) as PlayerView;
const card = (id: string, name: string, controller: string) =>
  ({ instance_id: id, name, owner: controller, controller, known_by_you: true }) as CardView;

const BOT = "bot";
const dragon = card("dragon", "Armament Dragon", BOT);
const elves = card("elves", "Llanowar Elves", BOT);

function model(distribution: Record<string, number> | undefined) {
  const trig = {
    id: "t1",
    kind: "triggered",
    controller: BOT,
    owner: BOT,
    source_card_id: "dragon",
    label: "Armament Dragon — distribute three +1/+1 counters",
    ability_text: "distribute three +1/+1 counters among one, two, or three target creatures",
    targets: [
      { kind: "card", id: "elves" },
      { kind: "card", id: "dragon" },
    ],
    distribution,
  } as StackItemView;
  return buildStackLane({
    stack: zone("stack"),
    stackItems: [trig],
    pendingTriggers: [],
    seats: [seat("me", "Me", 0), seat(BOT, "Bot", 1)],
    battlefield: zone("battlefield", [dragon, elves]),
    exile: zone("exile"),
    viewerID: "me",
    priorityHolder: "me",
    splitSecondActive: false,
  });
}

const LABEL = "Armament Dragon — distribute three +1/+1 counters";

describe("planShareBadges", () => {
  it("badges each target with its share, in the top item's colour", () => {
    expect(planShareBadges(model({ elves: 2, dragon: 1 }).stackItems)).toEqual([
      {
        kind: "permanent",
        id: "elves",
        text: "+2",
        title: `Llanowar Elves gets 2 counters from ${LABEL}`,
        color: TOP_ARROW_COLOR,
      },
      {
        kind: "permanent",
        id: "dragon",
        text: "+1",
        title: `Armament Dragon gets 1 counter from ${LABEL}`,
        color: TOP_ARROW_COLOR,
      },
    ]);
  });

  it("badges nothing for an item that divides nothing", () => {
    expect(planShareBadges(model(undefined).stackItems)).toEqual([]);
  });
});

describe("measureShareBadges", () => {
  it("pins a badge to the top-right corner of each target on the table", () => {
    const board = document.createElement("div");
    board.innerHTML = `
      <div data-instance-id="elves" id="elves-el"></div>
      <div data-instance-id="dragon" id="dragon-el"></div>`;
    document.body.appendChild(board);
    const realRect = Element.prototype.getBoundingClientRect;
    Element.prototype.getBoundingClientRect = function (this: Element) {
      return (
        this.id === "elves-el"
          ? { left: 100, top: 50, right: 160, bottom: 140, width: 60, height: 90 }
          : { left: 300, top: 50, right: 360, bottom: 140, width: 60, height: 90 }
      ) as DOMRect;
    };
    try {
      const badges = measureShareBadges(board, model({ elves: 2, dragon: 1 }).stackItems);
      expect(badges.map((b) => [b.key, b.text, b.right, b.top])).toEqual([
        ["permanent:elves", "+2", 160, 50],
        ["permanent:dragon", "+1", 360, 50],
      ]);
    } finally {
      Element.prototype.getBoundingClientRect = realRect;
      board.remove();
    }
  });

  it("badges nothing for a target that is not on screen", () => {
    const board = document.createElement("div");
    document.body.appendChild(board);
    try {
      expect(measureShareBadges(board, model({ elves: 2, dragon: 1 }).stackItems)).toEqual([]);
    } finally {
      board.remove();
    }
  });
});
