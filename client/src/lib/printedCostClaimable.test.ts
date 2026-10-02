// @vitest-environment jsdom
//
// printedCostClaimable.test.ts — #1012. `castable_here` is one bit and
// means "you may cast this from here", never "you may cast this from
// here for the cost in the corner". The client used to infer the
// second sentence from the first, and the cost picker opened with "Its
// mana cost" preselected on a flashback-only graveyard cast — an
// announcement the server refuses with ErrCastCostRequired.
//
// The server now says which prices a cast may claim
// (`alternative_cost_required`), and these are the three readers of
// it: the picker, the zone browser's button label, and the tooltip
// derivation in timing.ts.

import { describe, it, expect, afterEach } from "vitest";

import AlternativeCostModal from "./components/board/AlternativeCostModal.svelte";
import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { printedCostClaimable } from "./targeting";
import { canCastFromHand } from "./timing";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { barPrimary } from "./test/dockView";

afterEach(cleanup);

const ME = "me";

function card(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "looting",
    name: "Faithless Looting",
    owner: ME,
    controller: ME,
    type_line: "Sorcery",
    mana_cost: "{R}",
    ...extras,
  };
}

const flashbackOnly = card({
  alternative_costs: [{ key: "flashback", label: "Flashback {2}{R}", mana_cost: "{2}{R}" }],
  alternative_cost_required: true,
  castable_here: true,
});

const overloadInHand = card({
  instance_id: "rift",
  name: "Cyclonic Rift",
  mana_cost: "{1}{U}",
  alternative_costs: [{ key: "overload", label: "Overload {6}{U}", mana_cost: "{6}{U}" }],
});

describe("printedCostClaimable — #1012", () => {
  it("is true unless the server says the printed cost is out", () => {
    expect(printedCostClaimable(card())).toBe(true);
    expect(printedCostClaimable(overloadInHand)).toBe(true);
    expect(printedCostClaimable(flashbackOnly)).toBe(false);
  });

  it("reads the flag rather than the shape of the offer list", () => {
    // A Gravecrawler under an Underworld Breach: one offer AND the
    // printed cost. Inferring "an offer list on a graveyard card means
    // the printed cost is gone" gets this exact card wrong, which is
    // why the flag exists.
    const crawler = card({
      instance_id: "crawler",
      name: "Gravecrawler",
      castable_here: true,
      alternative_costs: [{ key: "escape", label: "Escape {1}{B}", mana_cost: "{1}{B}" }],
    });
    expect(printedCostClaimable(crawler)).toBe(true);
  });
});

// --- the cost picker -------------------------------------------------

function mountPicker(c: CardView) {
  const confirmed: Array<{ key: string | undefined; optional: number[] }> = [];
  // ADR 0111 PR 6: a sheet in the action dock; Cast is the bar's primary.
  const view = render(
    DockHarness as never,
    {
      component: AlternativeCostModal,
      props: {
        card: c,
        onConfirm: (key: string | undefined, optional: number[]) =>
          confirmed.push({ key, optional }),
        onCancel: () => {},
      },
    } as never,
  );
  return { container: view.container, confirmed };
}

const optionNames = (c: HTMLElement): string[] =>
  [...c.querySelectorAll(".dock-sheet .prompt-options .prompt-opt .name")].map(
    (e) => e.textContent ?? "",
  );

const confirmButton = (c: HTMLElement): HTMLElement => barPrimary(c)!;

describe("the cost picker drops a price the server would refuse — #1012", () => {
  it("offers the printed cost when it is claimable, and confirms it by default", () => {
    const { container, confirmed } = mountPicker(overloadInHand);
    expect(optionNames(container)).toEqual(["Its mana cost", "Overload {6}{U}"]);
    click(confirmButton(container));
    expect(confirmed).toEqual([{ key: undefined, optional: [] }]);
  });

  it("drops the printed cost when it is not, and confirms the first offer", () => {
    const { container, confirmed } = mountPicker(flashbackOnly);
    expect(optionNames(container)).toEqual(["Flashback {2}{R}"]);
    click(confirmButton(container));
    // The whole bug in one assertion: the default announcement is the
    // flashback cost, not the `undefined` the server rejects.
    expect(confirmed).toEqual([{ key: "flashback", optional: [] }]);
  });
});

// --- the miracle window (#1686) ---------------------------------------
//
// A live miracle grant opens the miracle claim at instant speed and
// has nothing to say about the printed cost, which stays a sorcery
// (CastPermission.ForClaim, server/internal/game/cast_zones.go). Drawn
// on another player's turn, the sorcery-speed window is shut for the
// printed claim only — it is still one of the zone's prices
// (`alternative_cost_required` stays false, unlike a zone-bound
// flashback-only offer), just not choosable this instant
// (`printed_cost_timing_closed: true`). The miracle offer itself
// carries no `timing_closed`, because its own claim IS open.
//
// Before #1686 the server had no timing-aware signal at all: the
// picker read `alternative_cost_required`/`alternative_costs` alone,
// showed "Its mana cost" next to "Miracle {W}" regardless of the
// moment, and confirming the printed one failed at announce with a
// sorcery-speed toast.
const miracleOnAnothersTurn = card({
  instance_id: "term",
  name: "Terminus",
  type_line: "Sorcery",
  mana_cost: "{4}{W}{W}",
  alternative_costs: [{ key: "miracle", label: "Miracle {W}", mana_cost: "{W}" }],
  alternative_cost_required: false,
  printed_cost_timing_closed: true,
  castable_here: true,
});

describe("the picker during a miracle window offers only what the engine would accept — #1686", () => {
  it("drops the printed cost when only the miracle claim is open", () => {
    const { container, confirmed } = mountPicker(miracleOnAnothersTurn);
    expect(optionNames(container)).toEqual(["Miracle {W}"]);
    click(confirmButton(container));
    expect(confirmed).toEqual([{ key: "miracle", optional: [] }]);
  });

  it("shows both prices once the printed cost's own timing is open too", () => {
    // The owner's own precombat main phase, say — sorcery speed is
    // open, so both claims are genuinely on the menu and nothing here
    // should hide either one.
    const bothOpen = card({
      instance_id: "term",
      name: "Terminus",
      type_line: "Sorcery",
      mana_cost: "{4}{W}{W}",
      alternative_costs: [{ key: "miracle", label: "Miracle {W}", mana_cost: "{W}" }],
      castable_here: true,
    });
    expect(optionNames(mountPicker(bothOpen).container)).toEqual(["Its mana cost", "Miracle {W}"]);
  });
});

// --- the zone browser's button label ----------------------------------

function snapWithGraveyard(cards: CardView[]): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      {
        id: ME,
        name: "Me",
        hand: { kind: "hand", count: 0, cards: [] },
        graveyard: { kind: "graveyard", count: cards.length, cards },
      },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  } as unknown as GameView;
}

function mountBrowser(cards: CardView[]) {
  const view = render(
    ZoneBrowserModal as never,
    {
      view: snapWithGraveyard(cards),
      viewerID: ME,
      zoneKind: "graveyard",
      ownerSeat: { id: ME, name: "Me" },
      sendAction: (_type: ActionType) => {},
      onClose: () => {},
      onCastCard: () => {},
    } as never,
  );
  return view.container;
}

const castLabel = (c: HTMLElement): string =>
  (c.querySelector("[aria-label='cast from graveyard'] button")?.textContent ?? "").trim();

describe("the graveyard cast button names one way in — #1012", () => {
  it("names the offer when it is the only price", () => {
    expect(castLabel(mountBrowser([flashbackOnly]))).toBe("Flashback {2}{R}");
  });

  it("falls back to the verb when the printed cost is a price too", () => {
    const crawler = card({
      instance_id: "crawler",
      name: "Gravecrawler",
      castable_here: true,
      alternative_costs: [{ key: "escape", label: "Escape {1}{B}", mana_cost: "{1}{B}" }],
    });
    expect(castLabel(mountBrowser([crawler]))).toBe("cast");
  });
});

// --- the tooltip derivation -------------------------------------------

describe("canCastFromHand weighs the printed clause only when it is payable — #1012", () => {
  function snapWithMoves(): GameView {
    return {
      id: "g",
      state: "active",
      seats: [{ id: ME, name: "Me", hand: { kind: "hand", count: 0, cards: [] } }],
      battlefield: { kind: "battlefield", count: 0, cards: [] },
      stack: { kind: "stack", count: 0, cards: [] },
      exile: { kind: "exile", count: 0, cards: [] },
      turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
      legal_moves: [],
    } as unknown as GameView;
  }

  it("still explains an unsatisfiable printed clause", () => {
    const bolt = card({
      instance_id: "bolt",
      name: "Lightning Bolt",
      type_line: "Instant",
      legal_targets: { cards: [], players: [], min: 1, max: 1 },
    });
    expect(canCastFromHand(bolt, snapWithMoves(), ME).legal).toBe(false);
  });

  it("does not let an unclaimable printed clause excuse the offers", () => {
    // The printed clause HAS a legal target, but the printed cost is
    // not a price this cast may claim — so the offer's clause is the
    // only one that counts, and it has none.
    const bound = card({
      instance_id: "bound",
      name: "Bound Spell",
      type_line: "Instant",
      legal_targets: { cards: ["a-creature"], players: [], min: 1, max: 1 },
      alternative_cost_required: true,
      alternative_costs: [
        {
          key: "flashback",
          label: "Flashback {2}{R}",
          mana_cost: "{2}{R}",
          legal_targets: { cards: [], players: [], min: 1, max: 1 },
        },
      ],
    });
    const verdict = canCastFromHand(bound, snapWithMoves(), ME);
    expect(verdict.legal).toBe(false);
    expect(verdict.reason).toBe("No legal target");
  });
});
