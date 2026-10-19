// @vitest-environment jsdom
//
// sideboard.render.test.ts — ADR 0191, amended 2026-10-19 (#3833, #4054).
// A deck's sideboard is the cards you own outside the game. It is not on
// the table: no pile in the player rail, no tab in the zone browser. The
// server still sends the cards to their owner alone (a count to everyone
// else), and in a game only an effect reaches them, so the one view that
// remains lists the cards a live right lets the owner cast or play. These
// tests pin what a player would notice: nothing is drawn on the table,
// the right's view shows only what it opens, and nothing in it can move.

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import PileBar from "./components/board/PileBar.svelte";
import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { closeZoneBrowser } from "./zoneBrowser";
import { buildMovePayload, canManageZone, cardsForZone } from "./zoneBrowser.logic";
import { render, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  closeZoneBrowser();
});
beforeEach(() => closeZoneBrowser());

const ME = "me";
const THEM = "them";

const zone = (kind: string, count = 0): ZoneView => ({ kind, count, cards: [] });

function sideCard(id: string, name: string, extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    type_line: "Sorcery",
    known_by_you: true,
    ...extras,
  } as unknown as CardView;
}

// A card a live right opens, and one it does not.
const WISH = sideCard("s1", "Burning Wish", { castable_here: true });
const BOLT = sideCard("s2", "Lightning Bolt");

function seatOf(id: string, extras: Partial<PlayerView> = {}): PlayerView {
  return {
    id,
    name: id === ME ? "Me" : "Them",
    seat: id === ME ? 0 : 1,
    life: 40,
    library: zone("library", 60),
    hand: zone("hand"),
    graveyard: zone("graveyard"),
    command: zone("command"),
    commander_damage: {},
    life_history: [],
    ...extras,
  } as unknown as PlayerView;
}

function gameWith(me: Partial<PlayerView>, them: Partial<PlayerView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seatOf(ME, me), seatOf(THEM, them)],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    turn: { seq: 1, number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  } as unknown as GameView;
}

describe("sideboard zone logic", () => {
  const view = gameWith({ sideboard: [WISH, BOLT], sideboard_count: 2 }, { sideboard_count: 3 });

  it("lists the owner's cards a right opens, not the rest of the sideboard", () => {
    expect(cardsForZone(view, "sideboard", ME).map((c) => c.instance_id)).toEqual(["s1"]);
    // Another seat carries only a count: there is nothing to show.
    expect(cardsForZone(view, "sideboard", THEM)).toEqual([]);
  });

  it("lists nothing while no right is live", () => {
    const quiet = gameWith({ sideboard: [BOLT], sideboard_count: 1 });
    expect(cardsForZone(quiet, "sideboard", ME)).toEqual([]);
  });

  it("is read only for everybody, its owner included", () => {
    expect(canManageZone("sideboard", ME, ME)).toBe(false);
    expect(canManageZone("sideboard", THEM, ME)).toBe(false);
    expect(canManageZone("sideboard", null, ME)).toBe(false);
  });

  it("builds no move for a sideboard card (CR 400.11c)", () => {
    for (const dest of ["hand", "battlefield", "library"] as const) {
      expect(buildMovePayload("sideboard", ME, ME, WISH, dest)).toBeNull();
    }
  });
});

describe("the player rail", () => {
  function mountBar(me: Partial<PlayerView>, isSelf = true) {
    const view = render(
      PileBar as never,
      {
        seat: seatOf(ME, me),
        exile: zone("exile"),
        isSelf,
        sendAction: () => {},
      } as never,
    );
    return view.container;
  }

  it("has no sideboard pile, whatever the seat holds", () => {
    for (const me of [
      { sideboard: [WISH, BOLT], sideboard_count: 2 },
      { outside_play: { source_name: "Wish", can_name: true } },
      {},
    ]) {
      const c = mountBar(me);
      expect(c.querySelector('[data-pile="sideboard"]')).toBeNull();
      expect(c.textContent).not.toContain("sideboard");
      cleanup();
    }
  });

  it("keeps the library, graveyard and exile piles", () => {
    const c = mountBar({ sideboard: [WISH], sideboard_count: 1 });
    for (const pile of ["library", "graveyard", "exile"]) {
      expect(c.querySelector(`[data-pile="${pile}"]`), pile).not.toBeNull();
    }
  });

  it("draws nothing of another seat's sideboard either", () => {
    const c = mountBar({ sideboard_count: 5 }, false);
    expect(c.querySelector('[data-pile="sideboard"]')).toBeNull();
  });
});

describe("the zone browser", () => {
  function mountModal(
    view: GameView,
    viewerID: string | null,
    zoneKind: "sideboard" | "graveyard" = "sideboard",
    ownerID = ME,
  ) {
    const sent: ActionType[] = [];
    const r = render(
      ZoneBrowserModal as never,
      {
        view,
        viewerID,
        zoneKind,
        ownerSeat: { id: ownerID, name: ownerID === ME ? "Me" : "Them" },
        sendAction: (t: ActionType) => sent.push(t),
        onClose: () => {},
      } as never,
    );
    return { c: r.container, sent };
  }

  const tabsOf = (c: HTMLElement) =>
    [...c.querySelectorAll<HTMLElement>('[role="tab"]')].map((t) => t.textContent?.trim());

  it("lists the cards a right opens, says what they are, and moves nothing", () => {
    const m = mountModal(gameWith({ sideboard: [WISH, BOLT], sideboard_count: 2 }), ME);
    const text = m.c.textContent ?? "";
    expect(text).toContain("outside the game");
    expect(text).toContain("(1)");
    expect(m.c.querySelectorAll(".cell").length).toBe(1);
    expect(text).not.toContain("Lightning Bolt");
    // No move buttons: nothing in it can be moved.
    expect(m.c.querySelector("button.move, [data-move]")).toBeNull();
    expect(m.sent).toEqual([]);
  });

  it("is no tab: it stands alone, and no other zone offers it", () => {
    const view = gameWith({ sideboard: [WISH], sideboard_count: 1 });
    expect(tabsOf(mountModal(view, ME).c)).toEqual([]);
    cleanup();
    expect(tabsOf(mountModal(view, ME, "graveyard").c)).toEqual(["graveyard", "exile", "command"]);
  });

  it("carries no Name-a-card button: that is in the dock", () => {
    const m = mountModal(
      gameWith({ outside_play: { source_name: "Wish", can_name: true }, sideboard: [WISH] }),
      ME,
    );
    expect(m.c.querySelector('[data-testid="name-outside-card"]')).toBeNull();
  });

  it("shows an opponent nothing of a seat's sideboard", () => {
    const m = mountModal(gameWith({ sideboard_count: 1 }), THEM, "sideboard", ME);
    expect(m.c.querySelectorAll(".cell").length).toBe(0);
  });
});
