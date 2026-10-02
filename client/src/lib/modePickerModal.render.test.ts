// @vitest-environment jsdom
//
// Spree (CR 702.172a, ADR 0065's 2026-09-23 amendment): ModePickerModal
// is the CAST-time picker (Spec.Modes on a hand card), the sibling of
// modePrompt.render.test.ts's mode_pick choice prompt. This file pins
// the one thing that amendment adds to the markup: each bullet's own
// additional cost, shown beside its label, plus a running summary of
// the chosen bullets' costs beside Confirm.

import { describe, it, expect, afterEach } from "vitest";

import ModePickerModal from "./components/board/ModePickerModal.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";

afterEach(cleanup);

function threeStepsAhead(): CardView {
  return {
    instance_id: "tsa-1",
    name: "Three Steps Ahead",
    owner: "me",
    controller: "me",
    modes: {
      prompt: "Spree (choose one or more)",
      min: 1,
      max: 3,
      options: [
        { label: "Counter target spell.", cost: "{1}{U}" },
        {
          label: "Create a token that's a copy of target artifact or creature you control.",
          cost: "{3}",
        },
        { label: "Draw two cards, then discard a card.", cost: "{2}" },
      ],
    },
  } as unknown as CardView;
}

const optionButtons = (c: HTMLElement): HTMLElement[] =>
  Array.from(c.querySelectorAll(".dock-sheet .prompt-options .prompt-opt")) as HTMLElement[];

// ADR 0111 PR 6: the picker is a sheet in the action dock.
function mountPicker() {
  return render(
    DockHarness as never,
    {
      component: ModePickerModal,
      props: {
        card: threeStepsAhead(),
        onConfirm: () => {},
        onCancel: () => {},
      },
    } as never,
  );
}

describe("ModePickerModal — Spree's per-mode cost (CR 702.172a)", () => {
  it("shows each bullet's own additional cost beside its label", () => {
    const { container } = mountPicker();

    const opts = optionButtons(container);
    expect(opts.length).toBe(3);
    expect(opts[0].querySelector(".cost")?.textContent).toBe("+{1}{U}");
    expect(opts[1].querySelector(".cost")?.textContent).toBe("+{3}");
    expect(opts[2].querySelector(".cost")?.textContent).toBe("+{2}");
  });

  it("summarises the chosen bullets' costs in the sheet, in the order chosen", () => {
    const { container } = mountPicker();

    // Nothing chosen yet: no summary.
    expect(container.querySelector(".extra-cost")).toBeNull();

    const opts = optionButtons(container);
    click(opts[2]); // {2}
    click(opts[0]); // {1}{U}
    const summary = container.querySelector(".extra-cost");
    expect(summary).not.toBeNull();
    expect(summary?.textContent).toBe("+{2} {1}{U}");
  });
});
