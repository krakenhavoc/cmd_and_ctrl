// @vitest-environment jsdom
//
// #2880: while a pending choice's permanents are picked on the board, a
// permanent it offers wears the target ring (and the pick ring once
// picked), and every other permanent is dimmed and not clickable. The
// sheet's own copies of the cards (a Card with no click) are left alone.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import { boardChoicePick } from "./boardChoicePick";
import type { CardView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  boardChoicePick.set(null);
});

const perm = (id: string): CardView =>
  ({
    instance_id: id,
    name: "Nazgûl",
    owner: "me",
    controller: "me",
    type_line: "Creature — Wraith Knight",
  }) as unknown as CardView;

const onClick = () => {};

function cardEl(id: string, withClick = true): HTMLElement {
  const { container } = render(
    Card as never,
    { card: perm(id), ...(withClick ? { onClick } : {}) } as Record<string, unknown>,
  );
  return container.querySelector<HTMLElement>(".card")!;
}

describe("a permanent during a board choice pick", () => {
  it("is ringed when offered, and ringed as picked once picked", () => {
    boardChoicePick.set({
      choiceID: "c",
      eligible: new Set(["n1", "n2"]),
      selected: new Set(["n2"]),
      min: 1,
      max: 1,
    });
    const offered = cardEl("n1");
    expect(offered.classList.contains("targetable")).toBe(true);
    expect(offered.classList.contains("picked")).toBe(false);
    expect(offered.classList.contains("clickable")).toBe(true);
    const picked = cardEl("n2");
    expect(picked.classList.contains("picked")).toBe(true);
  });

  it("is dimmed and not clickable when the choice does not offer it", () => {
    boardChoicePick.set({
      choiceID: "c",
      eligible: new Set(["n1"]),
      selected: new Set(),
      min: 1,
      max: 1,
    });
    const other = cardEl("forest");
    expect(other.classList.contains("choice-dimmed")).toBe(true);
    expect(other.classList.contains("clickable")).toBe(false);
    expect(other.classList.contains("targetable")).toBe(false);
  });

  it("leaves a card with no click (the sheet's copy) alone, and nothing once it closes", () => {
    boardChoicePick.set({
      choiceID: "c",
      eligible: new Set(["n1"]),
      selected: new Set(["n1"]),
      min: 1,
      max: 1,
    });
    const copy = cardEl("n1", false);
    expect(copy.classList.contains("targetable")).toBe(false);
    expect(copy.classList.contains("picked")).toBe(false);
    const live = cardEl("n1");
    boardChoicePick.set(null);
    flushSync();
    expect(live.classList.contains("targetable")).toBe(false);
    expect(live.classList.contains("choice-dimmed")).toBe(false);
  });
});
