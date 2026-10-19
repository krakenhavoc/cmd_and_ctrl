// @vitest-environment jsdom
//
// spliceCost.render.test.ts — #4055 (ADR 0202's 2026-10-19 amendment, CR
// 702.47a): a splice cost that is not mana. The cost picker shows the cost as
// printed in place of a mana symbol, asks which opponent gains the life for
// Roar of Jukai, and the object picks (a sacrifice, a return, a tap, a
// graveyard exile) ride `splice_payments`, entry k paying `splices[k]`.

import { describe, it, expect, afterEach } from "vitest";

import AlternativeCostModal from "./components/board/AlternativeCostModal.svelte";
import type { CardView, PlayerView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { barPrimary } from "./test/dockView";
import {
  applyCastChoices,
  cancel,
  spliceCostStepsOf,
  splicePaymentNames,
  withSplicePick,
} from "./targeting";

afterEach(() => {
  cleanup();
  cancel();
});

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
        card_id: "torrent",
        name: "Torrent of Stone",
        label: "Splice onto Arcane—Sacrifice two Mountains.",
        mana_cost: "",
        cost_label: "Sacrifice two Mountains",
        sacrifice_label: "two Mountains",
        sacrifice_options: { min: 2, max: 2, cards: ["m1", "m2", "m3"] },
        target_mode: "creature",
        legal_targets: { min: 1, max: 1, cards: ["bear"] },
      },
      {
        card_id: "roar",
        name: "Roar of Jukai",
        label: "Splice onto Arcane—An opponent gains 5 life.",
        mana_cost: "",
        cost_label: "An opponent gains 5 life",
        chooses_opponent: true,
        opponent_options: ["o1", "o2"],
      },
      {
        card_id: "ray",
        name: "Glacial Ray",
        label: "Splice onto Arcane {1}{R}",
        mana_cost: "{1}{R}",
      },
    ],
  } as unknown as CardView;
}

const seats = [
  { id: "o1", name: "Alice" },
  { id: "o2", name: "Bob" },
] as unknown as PlayerView[];

function mount(c: CardView) {
  const confirmed: Array<{ splices?: string[]; opponents?: Record<string, string> }> = [];
  const view = render(
    DockHarness as never,
    {
      component: AlternativeCostModal,
      props: {
        card: c,
        seats,
        onConfirm: (
          _k: string | undefined,
          _o: number[],
          _g?: string,
          _b?: number,
          _a?: string,
          splices?: string[],
          opponents?: Record<string, string>,
        ) => confirmed.push({ splices, opponents }),
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

const opponentRadios = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>(
    '.dock-sheet [aria-label="Give the life to"] [role="radio"]',
  ),
];

describe("the cost as printed", () => {
  it("shows a cost that is not mana in place of the mana symbol", () => {
    const { container } = mount(vortex());
    const [torrent, roar, ray] = spliceRows(container);
    expect(torrent.querySelector(".note")?.textContent).toBe("Sacrifice two Mountains");
    expect(roar.querySelector(".note")?.textContent).toBe("An opponent gains 5 life");
    // A mana splice keeps its symbols.
    expect(ray.querySelector(".note.cost")).not.toBeNull();
  });

  it("confirms a card with a cost to pay like any other", () => {
    const { container, confirmed } = mount(vortex());
    click(spliceRows(container)[0]);
    click(barPrimary(container)!);
    expect(confirmed).toEqual([{ splices: ["torrent"], opponents: {} }]);
  });
});

describe('the opponent an "An opponent gains N life" cost gives its life to', () => {
  it("asks only while the card is ticked, and holds Cast until one is chosen", () => {
    const { container, confirmed } = mount(vortex());
    expect(opponentRadios(container)).toHaveLength(0);
    click(spliceRows(container)[1]);
    const radios = opponentRadios(container);
    expect(radios.map((r) => r.textContent?.trim())).toEqual(["Alice", "Bob"]);
    expect(barPrimary(container)!.hasAttribute("disabled")).toBe(true);
    click(barPrimary(container)!);
    expect(confirmed).toEqual([]);
    click(radios[1]);
    expect(barPrimary(container)!.hasAttribute("disabled")).toBe(false);
    click(barPrimary(container)!);
    expect(confirmed).toEqual([{ splices: ["roar"], opponents: { roar: "o2" } }]);
  });

  it("forgets the pick when the card is unticked", () => {
    const { container, confirmed } = mount(vortex());
    const [, roar] = spliceRows(container);
    click(roar);
    click(opponentRadios(container)[0]);
    click(roar);
    click(roar);
    expect(barPrimary(container)!.hasAttribute("disabled")).toBe(true);
    click(opponentRadios(container)[1]);
    click(barPrimary(container)!);
    expect(confirmed).toEqual([{ splices: ["roar"], opponents: { roar: "o2" } }]);
  });
});

describe("the object picks", () => {
  it("names the picker each component opens", () => {
    const [torrent] = vortex().splice_offers!;
    expect(spliceCostStepsOf(torrent)).toEqual([
      {
        kind: "sacrifice",
        verb: "Sacrifice",
        label: "two Mountains",
        count: 2,
        options: ["m1", "m2", "m3"],
      },
    ]);
    expect(
      spliceCostStepsOf({
        card_id: "h",
        name: "Horobi's Whisper",
        label: "",
        mana_cost: "",
        exile_cost_n: 4,
        exile_cost_label: "four cards",
        exile_cost_options: ["a", "b", "c", "d", "e"],
      }),
    ).toEqual([
      {
        kind: "exile",
        verb: "Exile",
        label: "four cards",
        count: 4,
        options: ["a", "b", "c", "d", "e"],
      },
    ]);
    expect(
      spliceCostStepsOf({
        card_id: "t",
        name: "Hundred-Talon Strike",
        label: "",
        mana_cost: "",
        tap_others_label: "an untapped white creature you control",
        tap_others_options: { min: 1, max: 1, cards: ["w"] },
      }),
    ).toEqual([
      {
        kind: "tap",
        verb: "Tap",
        label: "an untapped white creature you control",
        count: 1,
        options: ["w"],
      },
    ]);
    expect(
      spliceCostStepsOf({
        card_id: "v",
        name: "Veil of Secrecy",
        label: "",
        mana_cost: "",
        return_label: "a blue creature you control",
        return_options: { min: 1, max: 1, cards: ["b"] },
      }),
    ).toEqual([
      {
        kind: "return",
        verb: "Return",
        label: "a blue creature you control",
        count: 1,
        options: ["b"],
      },
    ]);
  });

  it("asks nothing for a mana splice or a cost that is only an opponent", () => {
    const [, roar, ray] = vortex().splice_offers!;
    expect(spliceCostStepsOf(roar)).toEqual([]);
    expect(spliceCostStepsOf(ray)).toEqual([]);
  });

  it("grows the payments one card at a time", () => {
    let pays = withSplicePick(undefined, 1, "sacrifice", ["m1", "m2"]);
    expect(pays).toEqual([{}, { sacrificeIDs: ["m1", "m2"] }]);
    pays = withSplicePick(pays, 0, "tap", ["w"]);
    expect(pays).toEqual([{ tapIDs: ["w"] }, { sacrificeIDs: ["m1", "m2"] }]);
    expect(pays.map(splicePaymentNames)).toEqual([true, true]);
    expect(splicePaymentNames({})).toBe(false);
  });
});

describe("splice_payments on the wire", () => {
  it("rides cast_spell beside `splices`, entry k paying splices[k]", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, {
      splices: ["ray", "torrent", "roar"],
      splicePayments: [{}, { sacrificeIDs: ["m1", "m2"] }, { opponent: "o2" }],
    });
    expect(params.splices).toEqual(["ray", "torrent", "roar"]);
    expect(params.splice_payments).toEqual([
      {},
      { sacrifice_ids: ["m1", "m2"] },
      { opponent: "o2" },
    ]);
  });

  it("is omitted when no card has a cost to pay", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { splices: ["ray"], splicePayments: [{}] });
    expect(params.splice_payments).toBeUndefined();
    const none: Record<string, unknown> = {};
    applyCastChoices(none, { splices: ["ray"] });
    expect(none.splice_payments).toBeUndefined();
  });
});
