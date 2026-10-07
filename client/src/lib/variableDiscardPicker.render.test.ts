// @vitest-environment jsdom
//
// variableDiscardPicker.render.test.ts — #2527: "Discard X cards" (Gix,
// Yawgmoth Praetor) is DiscardCostModal with an open count. Any number
// of the offered cards, none included, confirms; the confirm names how
// many; the picks go back as the ids and their number is the X.

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
  instance_id: "gix",
  name: "Gix, Yawgmoth Praetor",
  owner: "me",
  controller: "me",
} as CardView;

function mount(options: CardView[], variable = true, need?: number) {
  const confirmed: string[][] = [];
  const view = render(
    DockHarness as never,
    {
      component: DiscardCostModal,
      props: {
        card: source,
        options,
        need,
        label: "X cards",
        variable,
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

describe("the discard-X picker", () => {
  it("confirms with nothing picked, as a discard of zero", () => {
    const { container, confirmed } = mount([card("a"), card("b")]);
    expect(confirmButton(container).disabled).toBe(false);
    expect(confirmButton(container).textContent).toContain("Discard 0");
    click(confirmButton(container));
    expect(confirmed).toEqual([[]]);
  });

  it("names how many are picked and sends them all", () => {
    const { container, confirmed } = mount([card("a"), card("b"), card("c")]);
    click(rows(container)[0]);
    click(rows(container)[2]);
    expect(confirmButton(container).textContent).toContain("Discard 2");
    expect(container.textContent).toContain("2 picked");
    click(confirmButton(container));
    expect(confirmed).toEqual([["a", "c"]]);
  });

  it("takes every card on offer and lets a pick be cleared", () => {
    const { container, confirmed } = mount([card("a"), card("b")]);
    for (const r of rows(container)) click(r);
    click(rows(container)[1]);
    click(confirmButton(container));
    expect(confirmed).toEqual([["a"]]);
  });

  it("still wants exactly N for a fixed-count discard", () => {
    const { container } = mount([card("a"), card("b"), card("c")], false, 2);
    expect(confirmButton(container).disabled).toBe(true);
    click(rows(container)[0]);
    click(rows(container)[1]);
    expect(confirmButton(container).disabled).toBe(false);
    click(rows(container)[2]);
    expect(confirmButton(container).disabled).toBe(false);
    expect(rows(container).filter((r) => r.getAttribute("aria-pressed") === "true")).toHaveLength(
      2,
    );
  });
});
