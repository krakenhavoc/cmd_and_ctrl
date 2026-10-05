// @vitest-environment jsdom
//
// legalHighlights.render.test.ts — ADR 0105 sub-PR 2 (#1789). What the
// cast surfaces DRAW from legalActions.ts, which only markup can show:
//
//   - a hand card with a legal move wears the ready ring and one
//     without does not, and `.timing-disabled` is exactly what it was
//     — the ring is decoration, the dim is a gate, and turning
//     highlights off touches only the first;
//   - the commander in the strip beside the hand is greyed, and
//     withholds the click, when the server offers no cast for it;
//   - a pile says how many of its cards are ready.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Hand from "./components/board/Hand.svelte";
import PileButton from "./components/board/PileButton.svelte";
import ExileStrip from "./components/board/ExileStrip.svelte";
import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { CardView, GameView, LegalMoveView, ZoneView } from "./protocol";
import { NO_LEGAL_ACTIONS, legalActionsOf, type LegalActions } from "./legalActions";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  localStorage.clear();
});

const ME = "me";

const bolt: CardView = {
  instance_id: "bolt",
  name: "Lightning Bolt",
  owner: ME,
  controller: ME,
  type_line: "Instant",
  mana_cost: "{R}",
};
const concentrate: CardView = {
  instance_id: "conc",
  name: "Concentrate",
  owner: ME,
  controller: ME,
  type_line: "Sorcery",
  mana_cost: "{2}{U}{U}",
};
const commander: CardView = {
  instance_id: "cmdr",
  name: "Kenrith, the Returned King",
  owner: ME,
  controller: ME,
  type_line: "Legendary Creature — Human Noble",
  mana_cost: "{4}{W}",
};

const castMove = (source: string, fromZone: string): LegalMoveView => ({
  type: "cast_spell",
  player: ME,
  kind: "cast",
  label: "Cast",
  source,
  params: { instance_id: source, from_zone: fromZone },
});
const passMove: LegalMoveView = {
  type: "pass_priority",
  player: ME,
  kind: "pass",
  label: "Pass",
  source: "00000000-0000-0000-0000-000000000000",
};

// frame is the viewer's own frame on their priority, with `moves` as
// the list and the digest the server would have folded from it.
function frame(moves: LegalMoveView[] | undefined, extra: Partial<GameView> = {}): GameView {
  const base = {
    id: "g",
    state: "active",
    seats: [
      {
        id: ME,
        name: "Me",
        hand: { kind: "hand", count: 2, cards: [bolt, concentrate] },
        graveyard: { kind: "graveyard", count: 0, cards: [] },
        library: { kind: "library", count: 0, cards: [] },
        command: { kind: "command", count: 1, cards: [commander] },
      },
      { id: "them", name: "Them", hand: { kind: "hand", count: 0, cards: [] } },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { seq: 4, number: 4, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    legal_moves: moves,
    ...extra,
  } as unknown as GameView;
  return base;
}

function mountHand(view: GameView, legal: LegalActions) {
  return render(
    Hand as never,
    {
      hand: view.seats[0].hand,
      isSelf: true,
      onPlayCard: () => {},
      snap: view,
      viewerID: ME,
      legal,
    } as never,
  );
}

// A ready card's accessible name is its name plus what it is ready for
// ("Lightning Bolt, castable", ADR 0105 §7), so match the name as the
// label's first part.
const cardOf = (c: HTMLElement, name: string) =>
  [...c.querySelectorAll<HTMLElement>(".card")].find((el) => {
    const label = el.getAttribute("aria-label") ?? "";
    return label === name || label.startsWith(`${name}, `);
  }) ?? null;
const slotOf = (c: HTMLElement, name: string) =>
  cardOf(c, name)?.closest<HTMLElement>(".hand-slot");

describe("hand: the ready ring", () => {
  const view = frame([passMove, castMove("bolt", "hand")]);

  it("rings a card with a move and leaves the one without it alone", () => {
    const { container } = mountHand(view, legalActionsOf(view));
    expect(cardOf(container, "Lightning Bolt")?.classList.contains("ready")).toBe(true);
    expect(cardOf(container, "Concentrate")?.classList.contains("ready")).toBe(false);
  });

  it("leaves .timing-disabled exactly as it was, ring or no ring", () => {
    const on = mountHand(view, legalActionsOf(view));
    expect(slotOf(on.container, "Lightning Bolt")?.classList.contains("timing-disabled")).toBe(
      false,
    );
    expect(slotOf(on.container, "Concentrate")?.classList.contains("timing-disabled")).toBe(true);
    on.destroy();

    // Highlights off (or autopass passing): no ring anywhere, and the
    // gate is unmoved.
    const off = mountHand(view, NO_LEGAL_ACTIONS);
    expect(off.container.querySelector(".card.ready")).toBeNull();
    expect(slotOf(off.container, "Lightning Bolt")?.classList.contains("timing-disabled")).toBe(
      false,
    );
    expect(slotOf(off.container, "Concentrate")?.classList.contains("timing-disabled")).toBe(true);
  });

  it("reads the digest when the server sends one", () => {
    const withDigest = frame([passMove, castMove("bolt", "hand")], {
      legal_actions: {
        pass: true,
        sources: { bolt: { kinds: ["cast"], moves: 1, zones: ["hand"], faces: [0] } },
      },
    });
    const { container } = mountHand(withDigest, legalActionsOf(withDigest));
    expect(cardOf(container, "Lightning Bolt")?.classList.contains("ready")).toBe(true);
    expect(cardOf(container, "Concentrate")?.classList.contains("ready")).toBe(false);
  });

  it("rings nothing on a frame with no list, and dims nothing new", () => {
    const quiet = frame(undefined);
    const { container } = mountHand(quiet, legalActionsOf(quiet));
    expect(container.querySelector(".card.ready")).toBeNull();
    // canCastFromHand's permissive reading of an absent list, untouched.
    expect(container.querySelector(".hand-slot.timing-disabled")).toBeNull();
  });
});

// #2349: the command zone tile and its "cast" hint are gone; the
// commander is cast from the strip beside the hand, which reads the
// same gate (canCastFromHand) and the same ring.
describe("command zone: the commander in the strip", () => {
  function mountZone(view: GameView, legal: LegalActions = NO_LEGAL_ACTIONS) {
    const cast: string[] = [];
    const r = render(
      ExileStrip as never,
      {
        view,
        viewerID: ME,
        legal,
        onCastCard: (c: CardView, zone: string) => cast.push(`${c.instance_id}@${zone}`),
      } as never,
    );
    const card = cardOf(r.container, "Kenrith, the Returned King")!;
    const slot = card.closest<HTMLElement>(".strip-slot")!;
    return { ...r, cast, card, slot };
  }

  it("is greyed, with a reason, when the server offers no cast for the commander", () => {
    const { slot, card, cast } = mountZone(frame([passMove, castMove("bolt", "hand")]));
    expect(slot.classList.contains("blocked")).toBe(true);
    expect(slot.title).not.toBe("");
    click(card);
    expect(cast).toEqual([]);
  });

  it("is live, and ready, when the server offers one", () => {
    const view = frame([passMove, castMove("cmdr", "command")]);
    const { slot, card, cast } = mountZone(view, legalActionsOf(view));
    expect(slot.classList.contains("blocked")).toBe(false);
    expect(card.classList.contains("ready")).toBe(true);
    click(card);
    expect(cast).toEqual(["cmdr@command"]);
  });

  it("highlights off: still live, no accent", () => {
    const view = frame([passMove, castMove("cmdr", "command")]);
    const { slot, container } = mountZone(view, NO_LEGAL_ACTIONS);
    expect(slot.classList.contains("blocked")).toBe(false);
    expect(container.querySelector(".card.ready")).toBeNull();
  });
});

describe("piles: N ready", () => {
  it("shows the count only when there is one", () => {
    const zone = { kind: "graveyard", count: 3, cards: [] } as unknown as ZoneView;
    const some = render(PileButton as never, { label: "grave", zone, readyCount: 2 } as never);
    expect(some.container.querySelector(".ready-count")?.textContent).toBe("2 ready");
    some.destroy();
    const none = render(PileButton as never, { label: "grave", zone } as never);
    expect(none.container.querySelector(".ready-count")).toBeNull();
  });
});

describe("exile strip: the ring and the chip's ready count come from the move list", () => {
  const plotted: CardView = {
    instance_id: "plotted",
    name: "Plotted Bolt",
    owner: ME,
    controller: ME,
    type_line: "Instant",
    mana_cost: "{R}",
    exile_play: { player: ME, cast_only: true, cost_override: "{0}" },
    castable_here: true,
    cast_prices: [{ cost: "{0}" }],
  };
  const unaffordable: CardView = {
    ...plotted,
    instance_id: "broke",
    name: "Expensive Thing",
    cast_prices: [{ cost: "{9}" }],
  };
  const view = frame([passMove, castMove("plotted", "exile")], {
    exile: { kind: "exile", count: 2, cards: [plotted, unaffordable] },
  });

  it("rings the card the server would accept and not the one it would refuse", () => {
    const { container } = render(
      ExileStrip as never,
      { view, viewerID: ME, onCastCard: () => {}, legal: legalActionsOf(view) } as never,
    );
    expect(cardOf(container, "Plotted Bolt")?.classList.contains("ready")).toBe(true);
    expect(cardOf(container, "Expensive Thing")?.classList.contains("ready")).toBe(false);
    expect(container.querySelector(".toggle-count.ready")?.textContent).toBe("1 ready");
  });

  it("highlights off: no ring and no ready count", () => {
    const { container } = render(
      ExileStrip as never,
      { view, viewerID: ME, onCastCard: () => {}, legal: NO_LEGAL_ACTIONS } as never,
    );
    expect(container.querySelector(".card.ready")).toBeNull();
    expect(container.querySelector(".toggle-count.ready")).toBeNull();
    // The click gate is unchanged: the refused card is still greyed.
    expect(
      cardOf(container, "Expensive Thing")?.closest(".strip-slot")?.classList.contains("blocked"),
    ).toBe(true);
  });
});

describe("zone browser: the graveyard button is shown-but-disabled", () => {
  const looting: CardView = {
    instance_id: "looting",
    name: "Faithless Looting",
    owner: ME,
    controller: ME,
    type_line: "Sorcery",
    mana_cost: "{R}",
    alternative_costs: [{ key: "flashback", label: "Flashback {2}{R}", mana_cost: "{2}{R}" }],
    alternative_cost_required: true,
    castable_here: true,
  } as unknown as CardView;

  function mountBrowser(moves: LegalMoveView[]) {
    const view = frame(moves);
    view.seats[0].graveyard = { kind: "graveyard", count: 1, cards: [looting] } as ZoneView;
    const cast: string[] = [];
    const r = render(
      ZoneBrowserModal as never,
      {
        view,
        viewerID: ME,
        zoneKind: "graveyard",
        ownerSeat: { id: ME, name: "Me" },
        sendAction: () => {},
        onClose: () => {},
        onCastCard: (c: CardView) => cast.push(c.instance_id),
        legal: legalActionsOf(view),
      } as never,
    );
    const button = r.container.querySelector<HTMLButtonElement>(
      "[aria-label='cast from graveyard'] button",
    );
    return { ...r, button, cast };
  }

  it("castable_here says yes but the move list says no: shown, disabled, not ringed", () => {
    const { button, cast, container } = mountBrowser([passMove]);
    expect(button).not.toBeNull();
    expect(button!.disabled).toBe(true);
    expect(button!.classList.contains("ready")).toBe(false);
    expect(container.querySelector(".card.ready")).toBeNull();
    click(button!);
    expect(cast).toEqual([]);
  });

  it("with the move: live, with the ready accent and the ring", () => {
    const { button, cast, container } = mountBrowser([passMove, castMove("looting", "graveyard")]);
    expect(button!.disabled).toBe(false);
    expect(button!.classList.contains("ready")).toBe(true);
    expect(cardOf(container, "Faithless Looting")?.classList.contains("ready")).toBe(true);
    click(button!);
    expect(cast).toEqual(["looting"]);
  });
});
