// @vitest-environment jsdom
//
// #3044: the land strip on a big board. What only the markup can say: a
// pile of 28 basics is one pile with its count, its stack only
// STRIP_DEPTH steps deep, every copy still drawn (so each can be
// hovered, clicked, targeted and found by instance ID); tapped lands
// pile by name after the untapped ones; and a click on the pile acts on
// its top copy. The pure rules are in landStrip.test.ts.

import { afterEach, describe, expect, it } from "vitest";

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import type { CardView } from "./protocol";
import { STRIP_DEPTH } from "./landStrip";
import { cleanup, click, render } from "./test/render.svelte";

afterEach(() => cleanup());

const land = (name: string, n: number, extra: Partial<CardView> = {}): CardView => ({
  instance_id: `${name[0].toLowerCase()}${n}`,
  name,
  owner: "me",
  controller: "me",
  type_line: `Basic Land — ${name}`,
  battle_x: n,
  ...extra,
});

function mountStrip(cards: CardView[]) {
  const clicked: string[] = [];
  const r = render(
    BattlefieldRow as never,
    {
      label: "lands",
      cards,
      viewerID: "me",
      strip: true,
      compact: true,
      onCardClick: (c: CardView) => clicked.push(c.instance_id),
      onGroupClick: () => {},
    } as never,
  );
  return { container: r.container, clicked };
}

const piles = (c: HTMLElement) => [...c.querySelectorAll<HTMLElement>(".pile")];
const count = (p: HTMLElement) => p.querySelector(".pile-count")?.textContent?.trim() ?? "";

describe("the land strip on a big board (#3044)", () => {
  it("draws 28 Forests as one pile with its count and a capped stack", () => {
    const forests = Array.from({ length: 28 }, (_, i) => land("Forest", i + 1));
    const { container } = mountStrip(forests);
    const all = piles(container);
    expect(all).toHaveLength(1);
    const pile = all[0];
    expect(count(pile)).toBe("28");
    expect(pile.classList.contains("deep")).toBe(true);
    const items = [...pile.querySelectorAll<HTMLElement>(':scope > [role="listitem"]')];
    // Every copy is still on the board, findable by its instance ID.
    expect(items).toHaveLength(28);
    expect(pile.querySelectorAll(".card[data-instance-id]")).toHaveLength(28);
    // Only STRIP_DEPTH - 1 copies step out of the stack; the rest are buried.
    expect(items.filter((i) => i.classList.contains("buried"))).toHaveLength(28 - STRIP_DEPTH);
    expect(items[27].style.getPropertyValue("--i")).toBe(String(STRIP_DEPTH - 1));
    expect(items[0].style.getPropertyValue("--i")).toBe("0");
  });

  it("piles tapped lands by name after the untapped ones", () => {
    const cards = [
      ...Array.from({ length: 10 }, (_, i) => land("Forest", i + 1)),
      ...Array.from({ length: 12 }, (_, i) => land("Forest", i + 11, { tapped: true })),
      ...Array.from({ length: 6 }, (_, i) => land("Island", i + 30, { tapped: i < 2 })),
    ];
    const { container } = mountStrip(cards);
    const all = piles(container);
    expect(all.map((p) => [p.classList.contains("tapped"), count(p)])).toEqual([
      [false, "10"],
      [false, "4"],
      [true, "12"],
      [true, "2"],
    ]);
  });

  it("acts on the pile's top copy when it is clicked", () => {
    const forests = Array.from({ length: 28 }, (_, i) => land("Forest", i + 1));
    const { container, clicked } = mountStrip(forests);
    const cards = [...piles(container)[0].querySelectorAll<HTMLElement>(".card")];
    click(cards[cards.length - 1]);
    expect(clicked).toEqual(["f28"]);
  });

  it("keeps a small pile's every copy stepped, as before", () => {
    const { container } = mountStrip([land("Forest", 1), land("Forest", 2), land("Forest", 3)]);
    const pile = piles(container)[0];
    expect(pile.classList.contains("deep")).toBe(false);
    expect(pile.querySelectorAll(".buried")).toHaveLength(0);
  });
});
