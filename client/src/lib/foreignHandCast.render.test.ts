// @vitest-environment jsdom
//
// foreignHandCast.render.test.ts — #5922. Sen Triplets: "You may play
// lands and cast spells from that player's hand this turn." The server
// stamps `castable_here` on the holder's copy of each card in that
// revealed hand it may play or cast now; a click on one runs the cast
// chain (onPlayCard), and a card without the stamp stays inert, as every
// card in an opponent's hand always was.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Hand from "./components/board/Hand.svelte";
import type { CardView, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";
const THEM = "them";

const theirs = (id: string, name: string, typeLine: string, castable: boolean): CardView =>
  ({
    instance_id: id,
    name,
    owner: THEM,
    controller: THEM,
    type_line: typeLine,
    known_by_you: true,
    castable_here: castable,
  }) as unknown as CardView;

const view = {
  id: "g1",
  state: "active",
  seats: [
    { id: ME, name: "Me", hand: { kind: "hand", count: 0, cards: [] } },
    { id: THEM, name: "Them", hand: { kind: "hand", count: 2, cards: [] } },
  ],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { seq: 4, number: 4, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  legal_moves: [],
} as unknown as GameView;

function mount(cards: CardView[]) {
  const played: string[] = [];
  const r = render(
    Hand as never,
    {
      hand: { kind: "hand", owner: THEM, count: cards.length, cards },
      isSelf: false,
      onPlayCard: (c: CardView) => played.push(c.instance_id),
      snap: view,
      viewerID: ME,
    } as never,
  );
  return { ...r, played };
}

const cardEl = (c: HTMLElement, name: string): HTMLElement =>
  c.querySelector<HTMLElement>(`.card[aria-label^='${name}']`)!;

describe("casting from another player's hand (#5922)", () => {
  it("casts a card the server marks castable_here, and leaves the rest inert", () => {
    const { container, played } = mount([
      theirs("bear", "Grizzly Bears", "Creature — Bear", true),
      theirs("giant", "Hill Giant", "Creature — Giant", false),
    ]);
    expect(cardEl(container, "Grizzly Bears").classList.contains("ready")).toBe(true);
    expect(cardEl(container, "Hill Giant").classList.contains("ready")).toBe(false);
    click(cardEl(container, "Grizzly Bears"));
    click(cardEl(container, "Hill Giant"));
    expect(played).toEqual(["bear"]);
  });

  it("plays a land from their hand the same way", () => {
    const { container, played } = mount([theirs("isl", "Island", "Basic Land — Island", true)]);
    click(cardEl(container, "Island"));
    expect(played).toEqual(["isl"]);
  });
});
