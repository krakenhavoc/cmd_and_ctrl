// @vitest-environment jsdom
//
// Pawprint modes (CR 700.2i, #4786): "Choose up to five {P} worth of
// modes. You may choose the same mode more than once." ModePickerModal
// counts the {P} the chosen modes spend rather than how many there are,
// shows each bullet's {P}, and won't add a bullet that would overspend.

import { describe, it, expect, afterEach } from "vitest";

import ModePickerModal from "./components/board/ModePickerModal.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";

afterEach(cleanup);

function seasonOfLoss(): CardView {
  return {
    instance_id: "sol-1",
    name: "Season of Loss",
    owner: "me",
    controller: "me",
    modes: {
      prompt: "Choose up to five {P} worth of modes",
      min: 0,
      max: 5,
      repeatable: true,
      pawprints: 5,
      options: [
        { label: "Each player sacrifices a creature of their choice.", pawprints: 1 },
        {
          label: "Draw a card for each creature that died under your control this turn.",
          pawprints: 2,
        },
        { label: "Each opponent loses X life.", pawprints: 3 },
      ],
    },
  } as unknown as CardView;
}

const optionButtons = (c: HTMLElement): HTMLElement[] =>
  Array.from(c.querySelectorAll(".dock-sheet .prompt-options .prompt-opt")) as HTMLElement[];
const text = (el: Element | null | undefined): string =>
  el?.textContent?.replace(/\s+/g, " ").trim() ?? "";
const castButton = (c: HTMLElement): HTMLButtonElement =>
  [...c.querySelectorAll<HTMLButtonElement>(".dock-bar button.request-primary")][0];

function mountPicker(confirmed: number[][]) {
  return render(
    DockHarness as never,
    {
      component: ModePickerModal,
      props: {
        card: seasonOfLoss(),
        onConfirm: (modes: number[]) => confirmed.push(modes),
        onCancel: () => {},
      },
    } as never,
  );
}

describe("ModePickerModal — pawprint modes (CR 700.2i)", () => {
  it("shows each bullet's {P}", () => {
    const { container } = mountPicker([]);
    const opts = optionButtons(container);
    expect(opts.map((o) => text(o.querySelector(".paws")))).toEqual(["{P}", "{P}{P}", "{P}{P}{P}"]);
  });

  it("spends the budget in {P} and refuses a bullet that would overspend it", () => {
    const confirmed: number[][] = [];
    const { container } = mountPicker(confirmed);
    const opts = optionButtons(container);

    click(opts[2]); // 3
    click(opts[1]); // 5
    expect(text(container.querySelector(".prompt-count"))).toBe("5 / 5 {P}");
    // Even the cheapest bullet no longer fits.
    expect(opts[0].getAttribute("aria-disabled")).toBe("true");
    click(opts[0]);
    expect(text(container.querySelector(".prompt-count"))).toBe("5 / 5 {P}");

    click(castButton(container));
    expect(confirmed).toEqual([[2, 1]]);
  });

  it("lets a one-{P} bullet be chosen five times", () => {
    const confirmed: number[][] = [];
    const { container } = mountPicker(confirmed);
    const opts = optionButtons(container);
    for (let n = 0; n < 6; n++) click(opts[0]);
    expect(text(container.querySelector(".prompt-count"))).toBe("5 / 5 {P}");
    click(castButton(container));
    expect(confirmed).toEqual([[0, 0, 0, 0, 0]]);
  });
});
