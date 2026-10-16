// @vitest-environment jsdom
//
// splice.render.test.ts — ADR 0202 (CR 702.47): an Arcane spell offers
// the splice cards in the caster's hand as toggles in the cost picker.
// The ticked cards ride the confirm in the order they were ticked
// (CR 702.47b), their splice costs ride the preview, and their clauses
// join the target walk after the spell's own, each pick stamped with the
// text it answers.

import { describe, it, expect, afterEach } from "vitest";
import { get } from "svelte/store";

import AlternativeCostModal from "./components/board/AlternativeCostModal.svelte";
import type { CardView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { barPrimary } from "./test/dockView";
import {
  advance,
  allPicks,
  applyCastChoices,
  begin,
  beginSpliceOnly,
  cancel,
  spliceSteps,
  targeting,
} from "./targeting";
import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";

afterEach(() => {
  cleanup();
  cancel();
});

// Consuming Vortex in hand with Glacial Ray and Desperate Ritual beside it.
function vortex(): CardView {
  return {
    instance_id: "vortex",
    name: "Consuming Vortex",
    owner: "me",
    controller: "me",
    mana_cost: "{1}{U}",
    target_mode: "creature",
    legal_targets: { min: 1, max: 1, cards: ["bear"] },
    splice_offers: [
      {
        card_id: "ray",
        name: "Glacial Ray",
        label: "Splice onto Arcane {1}{R}",
        mana_cost: "{1}{R}",
        target_mode: "any",
        legal_targets: { label: "any target", min: 1, max: 1, players: ["opp"], cards: ["bear"] },
      },
      {
        card_id: "ritual",
        name: "Desperate Ritual",
        label: "Splice onto Arcane {1}{R}",
        mana_cost: "{1}{R}",
      },
    ],
  } as unknown as CardView;
}

function mount(c: CardView) {
  const confirmed: Array<string[] | undefined> = [];
  const view = render(
    DockHarness as never,
    {
      component: AlternativeCostModal,
      props: {
        card: c,
        onConfirm: (
          _k: string | undefined,
          _o: number[],
          _g?: string,
          _b?: number,
          _a?: string,
          splices?: string[],
        ) => confirmed.push(splices),
        onCancel: () => {},
      },
    } as never,
  );
  return { container: view.container, confirmed };
}

const spliceRows = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>(
    '.dock-sheet [aria-label="Splice onto this spell"] .prompt-opt',
  ),
];

describe("the splice toggles", () => {
  it("lists every splice card, none ticked", () => {
    const { container, confirmed } = mount(vortex());
    const rows = spliceRows(container);
    expect(rows.map((b) => b.querySelector(".name")?.textContent?.trim())).toEqual([
      "Glacial Ray",
      "Desperate Ritual",
    ]);
    expect(rows.map((b) => b.getAttribute("aria-pressed"))).toEqual(["false", "false"]);
    click(barPrimary(container)!);
    expect(confirmed).toEqual([[]]);
  });

  it("confirms the ticked cards in the order they were ticked", () => {
    const { container, confirmed } = mount(vortex());
    const [ray, ritual] = spliceRows(container);
    click(ritual);
    click(ray);
    expect(ray.textContent).toContain("happens second");
    click(barPrimary(container)!);
    expect(confirmed).toEqual([["ritual", "ray"]]);
  });

  it("unticks a card", () => {
    const { container, confirmed } = mount(vortex());
    const [ray] = spliceRows(container);
    click(ray);
    click(ray);
    click(barPrimary(container)!);
    expect(confirmed).toEqual([[]]);
  });
});

describe("the spliced texts on the wire", () => {
  it("rides cast_spell and the preview as `splices`", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { splices: ["ray", "ritual"] });
    expect(params.splices).toEqual(["ray", "ritual"]);
    expect(castPreviewParams({ splices: ["ray"] }).splices).toEqual(["ray"]);
    expect(castPreviewParamsFromPayload({ splices: ["ray"] }).splices).toEqual(["ray"]);
    const none: Record<string, unknown> = {};
    applyCastChoices(none, { splices: [] });
    expect(none.splices).toBeUndefined();
  });

  it("adds a step per targeted spliced text, stamped with its position", () => {
    const steps = spliceSteps(vortex(), { splices: ["ritual", "ray"] });
    // Desperate Ritual takes no target; Glacial Ray is the second text.
    expect(steps.length).toBe(1);
    expect(steps[0].splice).toBe(2);
    expect(steps[0].label).toBe("any target (Glacial Ray)");
  });

  it("walks the spell's own target first, then the spliced text's", () => {
    begin(vortex(), "creature", { splices: ["ray"] });
    let t = get(targeting)!;
    expect(t.steps.length).toBe(2);
    t = advance({ ...t, picked: [{ kind: "card", id: "bear" }] })!;
    expect(t.step).toBe(1);
    const picks = allPicks({ ...t, picked: [{ kind: "player", id: "opp" }] });
    expect(picks).toEqual([
      { kind: "card", id: "bear", mode: 0, slot: 0 },
      { kind: "player", id: "opp", mode: 0, slot: 0, splice: 1 },
    ]);
  });

  it("opens a walk for a spliced target when the spell itself takes none", () => {
    const reach = { ...vortex(), target_mode: undefined, legal_targets: undefined };
    expect(beginSpliceOnly(reach as CardView, { splices: ["ray"] })).toBe(true);
    expect(get(targeting)!.steps.map((s) => s.splice)).toEqual([1]);
    cancel();
    expect(beginSpliceOnly(reach as CardView, { splices: ["ritual"] })).toBe(false);
  });
});
