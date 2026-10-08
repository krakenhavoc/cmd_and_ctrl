// @vitest-environment jsdom
//
// revealCostPicker.render.test.ts — #2598: "Reveal X black cards from
// your hand" (Martyr of Bones) reuses DiscardCostModal with the verb
// changed. Any number of the offered cards, none included, confirms; the
// confirm says Reveal, not Discard; the picks go back as the ids and their
// number is the X.

import { describe, it, expect, afterEach } from "vitest";

import DiscardCostModal from "./components/board/DiscardCostModal.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { barPrimary } from "./test/dockView";

afterEach(cleanup);

function card(id: string): CardView {
  return { instance_id: id, name: id, owner: "me", controller: "me" } as CardView;
}

const source = {
  instance_id: "martyr",
  name: "Martyr of Bones",
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
        label: "X black cards",
        variable: true,
        verb: "Reveal",
        note: "cost · CR 602.2b",
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

describe("the reveal-X picker", () => {
  it("says Reveal, names the clause, and confirms zero", () => {
    const { container, confirmed } = mount([card("a"), card("b")]);
    expect(container.textContent).toContain("X black cards");
    expect(confirmButton(container).textContent).toContain("Reveal 0");
    expect(confirmButton(container).textContent).not.toContain("Discard");
    click(confirmButton(container));
    expect(confirmed).toEqual([[]]);
  });

  it("sends the picked cards, their number being the X", () => {
    const { container, confirmed } = mount([card("a"), card("b"), card("c")]);
    click(rows(container)[0]);
    click(rows(container)[2]);
    expect(confirmButton(container).textContent).toContain("Reveal 2");
    click(confirmButton(container));
    expect(confirmed).toEqual([["a", "c"]]);
  });
});
