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
  Array.from(c.querySelectorAll(".prompt-options .prompt-opt")) as HTMLElement[];

describe("ModePickerModal — Spree's per-mode cost (CR 702.172a)", () => {
  it("shows each bullet's own additional cost beside its label", () => {
    const { container } = render(
      ModePickerModal as never,
      {
        card: threeStepsAhead(),
        onConfirm: () => {},
        onCancel: () => {},
      } as never,
    );

    const opts = optionButtons(container);
    expect(opts.length).toBe(3);
    expect(opts[0].querySelector(".cost")?.textContent).toBe("+{1}{U}");
    expect(opts[1].querySelector(".cost")?.textContent).toBe("+{3}");
    expect(opts[2].querySelector(".cost")?.textContent).toBe("+{2}");
  });

  it("summarises the chosen bullets' costs beside Confirm, in the order chosen", () => {
    const { container } = render(
      ModePickerModal as never,
      {
        card: threeStepsAhead(),
        onConfirm: () => {},
        onCancel: () => {},
      } as never,
    );

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
