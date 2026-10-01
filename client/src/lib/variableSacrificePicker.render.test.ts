// @vitest-environment jsdom
//
// variableSacrificePicker.render.test.ts — ADR 0100 §5: a variable
// sacrifice on a cast is SacrificeCostModal with an open count. "Any
// number" confirms at zero and at every count the board allows, and a
// lone pick can be cleared again; "sacrifice X" says the number picked
// is X.

import { describe, it, expect, afterEach } from "vitest";

import SacrificeCostModal from "./components/board/SacrificeCostModal.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function creature(id: string): CardView {
  return {
    instance_id: id,
    name: id,
    owner: "me",
    controller: "me",
    power: 2,
    toughness: 2,
  } as CardView;
}

const source = {
  instance_id: "vb",
  name: "Vicious Betrayal",
  owner: "me",
  controller: "me",
} as CardView;

function mount(options: CardView[], countIsX = false) {
  const confirmed: string[][] = [];
  const view = render(
    SacrificeCostModal as never,
    {
      source,
      label: "any number of creatures",
      options,
      count: 0,
      min: 0,
      countIsX,
      onConfirm: (ids: string[]) => confirmed.push(ids),
      onCancel: () => {},
    } as never,
  );
  return { container: view.container, confirmed };
}

const rows = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>(".prompt-options .prompt-opt"),
];
const confirmButton = (c: HTMLElement): HTMLButtonElement =>
  c.querySelector(".prompt-foot .primary") as HTMLButtonElement;

describe("the any-number sacrifice picker", () => {
  it("confirms with nothing picked", () => {
    const { container, confirmed } = mount([creature("a"), creature("b")]);
    expect(container.textContent).toContain("Pick as many as you like, or none.");
    expect(confirmButton(container).disabled).toBe(false);
    click(confirmButton(container));
    expect(confirmed).toEqual([[]]);
  });

  it("takes every permanent on offer", () => {
    const { container, confirmed } = mount([creature("a"), creature("b"), creature("c")]);
    for (const r of rows(container)) click(r);
    click(confirmButton(container));
    expect(confirmed).toEqual([["a", "b", "c"]]);
  });

  it("clears a lone pick on a second click", () => {
    const { container, confirmed } = mount([creature("a")]);
    click(rows(container)[0]);
    click(rows(container)[0]);
    click(confirmButton(container));
    expect(confirmed).toEqual([[]]);
  });

  it("says the number picked is X for a sacrifice-X cost", () => {
    const { container } = mount([creature("a")], true);
    expect(container.textContent).toContain("The number you pick is X.");
  });
});
