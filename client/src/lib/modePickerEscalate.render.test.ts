// @vitest-environment jsdom
//
// Escalate (CR 702.120a, #2126): ModePickerModal shows the per-extra-mode
// cost beside Confirm once a second mode is chosen, and the server has
// already clamped `max` to the count the viewer can pay for, so a third
// mode is simply not selectable when only one extra can be paid.

import { describe, it, expect, afterEach } from "vitest";

import ModePickerModal from "./components/board/ModePickerModal.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";

afterEach(cleanup);

function collectiveBrutality(max: number, maxExtra: number): CardView {
  return {
    instance_id: "cb-1",
    name: "Collective Brutality",
    owner: "me",
    controller: "me",
    modes: {
      prompt: "Choose one or more",
      min: 1,
      max,
      options: [
        { label: "Target opponent reveals their hand." },
        { label: "Target creature gets -2/-2 until end of turn." },
        { label: "Target opponent loses 2 life and you gain 2 life." },
      ],
      escalate: { label: "Escalate—Discard a card", discard_cards: 1, max_extra: maxExtra },
    },
  } as unknown as CardView;
}

const optionButtons = (c: HTMLElement): HTMLElement[] =>
  Array.from(c.querySelectorAll(".dock-sheet .prompt-options .prompt-opt")) as HTMLElement[];

function mountPicker(card: CardView) {
  return render(
    DockHarness as never,
    {
      component: ModePickerModal,
      props: { card, onConfirm: () => {}, onCancel: () => {} },
    } as never,
  );
}

describe("ModePickerModal — escalate (CR 702.120a)", () => {
  it("summarises the escalate cost once per mode beyond the first", () => {
    const { container } = mountPicker(collectiveBrutality(3, 2));
    const opts = optionButtons(container);

    click(opts[0]);
    expect(container.querySelector(".extra-cost")).toBeNull();

    click(opts[1]);
    expect(container.querySelector(".extra-cost")?.textContent).toBe("Escalate—Discard a card ×1");

    click(opts[2]);
    expect(container.querySelector(".extra-cost")?.textContent).toBe("Escalate—Discard a card ×2");
  });

  it("does not let a third mode be chosen when the server clamped max to two", () => {
    const { container } = mountPicker(collectiveBrutality(2, 1));
    const opts = optionButtons(container);

    click(opts[0]);
    click(opts[1]);
    click(opts[2]);
    expect(container.querySelector(".extra-cost")?.textContent).toBe("Escalate—Discard a card ×1");
  });
});
