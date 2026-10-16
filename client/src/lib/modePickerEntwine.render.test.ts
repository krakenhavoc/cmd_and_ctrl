// @vitest-environment jsdom
//
// Entwine (CR 702.42a, #3866): with "Entwine {2}" ticked the server's
// `if_optional_paid` bounds are every bullet, no fewer, and the picker
// opens with them all chosen in printed order (CR 702.42b). Unentwined it
// opens empty, as every modal picker does.

import { describe, it, expect, afterEach } from "vitest";

import ModePickerModal from "./components/board/ModePickerModal.svelte";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { modesUnderChoices } from "./targeting";

afterEach(cleanup);

function spectralShift(landTargets: string[]): CardView {
  return {
    instance_id: "ss-1",
    name: "Spectral Shift",
    owner: "me",
    controller: "me",
    optional_costs: [
      { index: 0, key: "entwine", label: "Entwine {2}", mana_cost: "{2}", max_times: 1 },
    ],
    modes: {
      prompt: "Choose one",
      min: 1,
      max: 1,
      if_optional_paid: { min: 2, max: 2 },
      options: [
        {
          label:
            "Change the text of target spell or permanent by replacing all instances of one basic land type with another.",
          legal_targets: { min: 1, max: 1, cards: landTargets },
        },
        {
          label:
            "Change the text of target spell or permanent by replacing all instances of one color word with another.",
          legal_targets: { min: 1, max: 1, cards: ["bear"] },
        },
      ],
    },
  } as unknown as CardView;
}

const checked = (c: HTMLElement): string[] =>
  Array.from(c.querySelectorAll(".dock-sheet .prompt-options .prompt-opt")).map(
    (b) => (b as HTMLElement).getAttribute("aria-checked") ?? "",
  );

function mountPicker(card: CardView) {
  return render(
    DockHarness as never,
    {
      component: ModePickerModal,
      props: { card, onConfirm: () => {}, onCancel: () => {} },
    } as never,
  );
}

describe("ModePickerModal — entwine (CR 702.42a)", () => {
  it("opens with every bullet chosen when entwined", () => {
    const card = modesUnderChoices(spectralShift(["forest"]), { optionalCosts: [0] });
    const { container } = mountPicker(card);
    expect(checked(container)).toEqual(["true", "true"]);
  });

  it("opens empty when not entwined", () => {
    const card = modesUnderChoices(spectralShift(["forest"]), {});
    const { container } = mountPicker(card);
    expect(checked(container)).toEqual(["false", "false"]);
  });

  it("chooses nothing for the caster when a bullet has no legal target", () => {
    const card = modesUnderChoices(spectralShift([]), { optionalCosts: [0] });
    const { container } = mountPicker(card);
    expect(checked(container)).toEqual(["false", "false"]);
  });
});
