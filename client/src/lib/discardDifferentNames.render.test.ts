// @vitest-environment jsdom
//
// discardDifferentNames.render.test.ts — #4896: "Discard three cards with
// different names" (Ormos, Archive Keeper) is DiscardCostModal with the
// names rule. A card whose name is already picked is greyed, the confirm
// waits for three names, and a hand without three names is short.

import { describe, it, expect, afterEach } from "vitest";

import DiscardCostModal from "./components/board/DiscardCostModal.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { barPrimary } from "./test/dockView";

afterEach(cleanup);

function card(id: string, name: string): CardView {
  return { instance_id: id, name, owner: "me", controller: "me" } as CardView;
}

const source = {
  instance_id: "ormos",
  name: "Ormos, Archive Keeper",
  owner: "me",
  controller: "me",
} as CardView;

function mount(options: CardView[]) {
  const confirmed: string[][] = [];
  const view = render(
    DockHarness as never,
    {
      component: DiscardCostModal,
      props: {
        card: source,
        options,
        need: 3,
        label: "three cards with different names",
        differentNames: true,
        onConfirm: (ids: string[]) => confirmed.push(ids),
        onCancel: () => {},
      },
    } as never,
  );
  return { container: view.container, confirmed };
}

const rows = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>(".dock-sheet .prompt-options .prompt-opt"),
];
const confirmButton = (c: HTMLElement): HTMLButtonElement => barPrimary(c)!;

describe("the different-names discard picker", () => {
  it("greys a second card with a picked name and confirms three names", () => {
    const { container, confirmed } = mount([
      card("a", "Opt"),
      card("b", "Opt"),
      card("c", "Shock"),
      card("d", "Ponder"),
    ]);
    expect(container.textContent).toContain("must have different names");
    click(rows(container)[0]);
    expect(rows(container)[1].disabled).toBe(true);
    click(rows(container)[1]);
    click(rows(container)[2]);
    expect(confirmButton(container).disabled).toBe(true);
    click(rows(container)[3]);
    expect(confirmButton(container).disabled).toBe(false);
    click(confirmButton(container));
    expect(confirmed).toEqual([["a", "c", "d"]]);
  });

  it("frees the name again when the pick is taken back", () => {
    const { container } = mount([
      card("a", "Opt"),
      card("b", "Opt"),
      card("c", "Shock"),
      card("d", "Ponder"),
    ]);
    click(rows(container)[0]);
    expect(rows(container)[1].disabled).toBe(true);
    click(rows(container)[0]);
    expect(rows(container)[1].disabled).toBe(false);
  });

  it("is short when the hand holds too few names", () => {
    const { container } = mount([card("a", "Opt"), card("b", "Opt"), card("c", "Shock")]);
    expect(container.textContent).toContain("You don't have 3 cards with different names");
    expect(confirmButton(container).disabled).toBe(true);
  });
});
