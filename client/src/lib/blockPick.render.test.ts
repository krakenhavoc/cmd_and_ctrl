// @vitest-environment jsdom
//
// #3898: while a blocker is selected on the board, an attacker it can
// block wears the target ring, one it can't is dimmed, marked disabled
// and says why on hover, and a creature in a menace group wears the
// pick ring. A click on a refused attacker still reaches the board,
// whose state machine declares nothing and has the dock say why (the
// touch screen's way to read the reason).

import { describe, it, expect, afterEach } from "vitest";

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import Card from "./components/board/Card.svelte";
import { blockBoard, fannedPiles, type BlockBoard } from "./blockPick";
import type { CardView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  blockBoard.set(null);
});

const creature = (id: string, name = "Grizzly Bears"): CardView =>
  ({
    instance_id: id,
    name,
    owner: "opp",
    controller: "opp",
    type_line: "Creature — Bear",
    attacking_target: "me",
  }) as unknown as CardView;

const board = (over: Partial<BlockBoard> = {}): BlockBoard => ({
  viewerID: "me",
  selected: "wall",
  targets: new Map([
    ["bear", { kind: "block" }],
    [
      "drake",
      { kind: "refused", reason: "Drake has flying, and Wall has neither flying nor reach." },
    ],
  ]),
  group: null,
  ...over,
});

function cardEl(card: CardView, onClick: () => void = () => {}): HTMLElement {
  const { container } = render(Card as never, { card, onClick } as Record<string, unknown>);
  return container.querySelector<HTMLElement>(".card")!;
}

describe("an attacker while a blocker is selected", () => {
  it("is ringed when the blocker can block it", () => {
    blockBoard.set(board());
    const el = cardEl(creature("bear"));
    expect(el.classList.contains("combat-target")).toBe(true);
    expect(el.classList.contains("block-refused")).toBe(false);
    expect(el.classList.contains("clickable")).toBe(true);
  });

  it("is dimmed, disabled and says why when it can't", () => {
    blockBoard.set(board());
    const el = cardEl(creature("drake", "Drake"));
    expect(el.classList.contains("block-refused")).toBe(true);
    expect(el.classList.contains("combat-target")).toBe(false);
    expect(el.classList.contains("clickable")).toBe(false);
    expect(el.getAttribute("aria-disabled")).toBe("true");
    expect(el.getAttribute("title")).toBe(
      "Drake — can't block: Drake has flying, and Wall has neither flying nor reach.",
    );
  });

  it("goes back to normal when the selection ends", () => {
    blockBoard.set(board());
    const el = cardEl(creature("drake", "Drake"));
    blockBoard.set(null);
    flushSync();
    expect(el.classList.contains("block-refused")).toBe(false);
    expect(el.getAttribute("title")).toBe("Drake");
  });
});

describe("a pile of attackers while blockers are declared", () => {
  const goblin = (id: string): CardView =>
    ({
      instance_id: id,
      name: "Goblin",
      owner: "opp",
      controller: "opp",
      type_line: "Token Creature — Goblin",
      is_token: true,
      power: 1,
      toughness: 1,
      tapped: true,
      attacking_target: "me",
      attacking_target_kind: "player",
      defending_player: "me",
      battle_x: Number(id.replace(/\D/g, "")),
    }) as unknown as CardView;

  function mountRow(cards: CardView[]) {
    const opened: string[] = [];
    const clicked: string[] = [];
    const r = render(
      BattlefieldRow as never,
      {
        label: "creatures",
        cards,
        viewerID: "opp",
        onCardClick: (c: CardView) => clicked.push(c.instance_id),
        onGroupClick: (k: string) => opened.push(k),
      } as never,
    );
    const drawn = () =>
      [...r.container.querySelectorAll<HTMLElement>(".card[data-instance-id]")].map(
        (e) => e.dataset.instanceId,
      );
    return { container: r.container, opened, clicked, drawn };
  }

  afterEach(() => fannedPiles.set(new Set()));

  it("fans out into its cards on click, and the next click picks the copy", () => {
    blockBoard.set(board({ selected: null, targets: new Map() }));
    const row = mountRow([goblin("g1"), goblin("g2"), goblin("g3")]);
    expect(row.drawn()).toEqual(["g1"]);
    row.container.querySelector<HTMLElement>('.card[data-instance-id="g1"]')!.click();
    flushSync();
    expect(row.opened).toEqual([]);
    expect(row.drawn()).toEqual(["g1", "g2", "g3"]);
    row.container.querySelector<HTMLElement>('.card[data-instance-id="g3"]')!.click();
    expect(row.clicked).toEqual(["g3"]);
  });

  it("opens the member list as before when nobody is declaring blockers", () => {
    const row = mountRow([goblin("g1"), goblin("g2")]);
    row.container.querySelector<HTMLElement>('.card[data-instance-id="g1"]')!.click();
    flushSync();
    expect(row.opened.length).toBe(1);
    expect(row.drawn()).toEqual(["g1"]);
  });
});

describe("a creature in a menace group", () => {
  it("wears the pick ring", () => {
    blockBoard.set(
      board({
        selected: null,
        targets: new Map(),
        group: { attacker: "brute", blockers: ["wall"], need: 2 },
      }),
    );
    const el = cardEl({
      ...creature("wall", "Wall"),
      controller: "me",
      attacking_target: undefined,
    });
    expect(el.classList.contains("picked")).toBe(true);
  });
});
