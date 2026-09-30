// @vitest-environment jsdom
//
// eitherCostPicker.render.test.ts — ADR 0100 §5: an either/or
// additional cost is a radio group in the cost picker. It opens on the
// first branch the server says is payable, an unpayable branch is
// disabled, and the chosen index rides the confirm as cost_branch.

import { describe, it, expect, afterEach } from "vitest";

import AlternativeCostModal from "./components/board/AlternativeCostModal.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function lightningAxe(discardPayable: boolean): CardView {
  return {
    instance_id: "axe",
    name: "Lightning Axe",
    owner: "me",
    controller: "me",
    mana_cost: "{R}",
    additional_cost: {
      label: "Discard a card or pay {5}",
      branches: [
        { key: "discard", label: "Discard a card", discard_cards: 1, payable: discardPayable },
        { key: "mana", label: "Pay {5}", mana_cost: "{5}", payable: true },
      ],
    },
  } as CardView;
}

function mount(c: CardView) {
  const confirmed: Array<number | undefined> = [];
  const view = render(
    AlternativeCostModal as never,
    {
      card: c,
      onConfirm: (_k: string | undefined, _o: number[], _g?: string, branch?: number) =>
        confirmed.push(branch),
      onCancel: () => {},
    } as never,
  );
  return { container: view.container, confirmed };
}

const radios = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>('[role="radiogroup"] .prompt-opt'),
];
const confirmButton = (c: HTMLElement): HTMLElement =>
  c.querySelector(".prompt-foot .primary") as HTMLElement;

describe("the either/or branch radio", () => {
  it("lists every branch and confirms the first payable one by default", () => {
    const { container, confirmed } = mount(lightningAxe(true));
    expect(radios(container).map((b) => b.querySelector(".name")?.textContent)).toEqual([
      "Discard a card",
      "Pay {5}",
    ]);
    click(confirmButton(container));
    expect(confirmed).toEqual([0]);
  });

  it("disables an unpayable branch and opens on the payable one", () => {
    const { container, confirmed } = mount(lightningAxe(false));
    const [discard, mana] = radios(container);
    expect(discard.disabled).toBe(true);
    expect(mana.disabled).toBe(false);
    expect(mana.getAttribute("aria-checked")).toBe("true");
    click(confirmButton(container));
    expect(confirmed).toEqual([1]);
  });

  it("confirms the branch the player picks", () => {
    const { container, confirmed } = mount(lightningAxe(true));
    click(radios(container)[1]);
    click(confirmButton(container));
    expect(confirmed).toEqual([1]);
  });
});
