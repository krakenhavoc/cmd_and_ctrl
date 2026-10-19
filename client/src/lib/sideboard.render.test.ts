// @vitest-environment jsdom
//
// sideboard.render.test.ts — ADR 0191, amended 2026-10-19 (#3833, #4054).
// A deck's sideboard is the cards you own outside the game. It is no pile
// on the table, but the owner can read it, read only, in a Sideboard tab of
// the zone browser, to check what a Wish can fetch. The server sends the
// cards to their owner alone (a count to everyone else). These tests pin
// what a player would notice: nothing is drawn in the player rail, the tab
// is yours alone and lists the whole sideboard, and nothing in it can move.

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
const WISH = sideCard("s1", "Burning Wish", {
  castable_here: true,
  exile_play: { player: ME, cost_override: "{0}", cast_only: true } as CardView["exile_play"],
});
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

  it("reads the owner's whole sideboard, and nothing for a seat that sent none", () => {
    expect(cardsForZone(view, "sideboard", ME).map((c) => c.instance_id)).toEqual(["s1", "s2"]);
    // Another seat carries only a count: there is nothing to show.
    expect(cardsForZone(view, "sideboard", THEM)).toEqual([]);
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

describe("the zone browser's sideboard tab", () => {
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
        onCastCard: () => {},
      } as never,
    );
    return { c: r.container, sent };
  }

  const tabsOf = (c: HTMLElement) =>
    [...c.querySelectorAll<HTMLElement>('[role="tab"]')].map((t) => t.textContent?.trim());

  it("lists the whole sideboard, says it is read only and what the zone is, and moves nothing", () => {
    const m = mountModal(gameWith({ sideboard: [WISH, BOLT], sideboard_count: 2 }), ME);
    const text = m.c.textContent ?? "";
    expect(text).toContain("sideboard");
    expect(text).toContain("(2)");
    expect(text).toContain("read only");
    expect(text).toContain("outside the game");
    expect(text).toContain("Burning Wish");
    expect(text).toContain("Lightning Bolt");
    expect(m.c.querySelectorAll(".cell").length).toBe(2);
    // No move buttons: nothing in it can be moved.
    expect(m.c.querySelector("button.move, [data-move]")).toBeNull();
    expect(m.sent).toEqual([]);
  });

  it("puts a cast button only on a card a live right opens", () => {
    const m = mountModal(gameWith({ sideboard: [WISH, BOLT], sideboard_count: 2 }), ME);
    const cells = [...m.c.querySelectorAll<HTMLElement>(".cell")];
    const castBtn = (cell: HTMLElement) => cell.querySelector("button.act");
    expect(castBtn(cells.find((c) => c.textContent?.includes("Burning Wish"))!)).not.toBeNull();
    expect(castBtn(cells.find((c) => c.textContent?.includes("Lightning Bolt"))!)).toBeNull();
  });

  it("is a fourth tab on your own seat, and every zone offers it", () => {
    const view = gameWith({ sideboard: [WISH], sideboard_count: 1 });
    expect(tabsOf(mountModal(view, ME).c)).toEqual(["graveyard", "exile", "command", "sideboard"]);
    cleanup();
    expect(tabsOf(mountModal(view, ME, "graveyard").c)).toEqual([
      "graveyard",
      "exile",
      "command",
      "sideboard",
    ]);
  });

  it("is offered to a Wish that names any card, though the sideboard is empty", () => {
    const view = gameWith({ outside_play: { source_name: "Wish", can_name: true } });
    expect(tabsOf(mountModal(view, ME, "graveyard").c)).toContain("sideboard");
  });

  it("is no tab with an empty sideboard and no Wish", () => {
    expect(tabsOf(mountModal(gameWith({}), ME, "graveyard").c)).toEqual([
      "graveyard",
      "exile",
      "command",
    ]);
  });

  it("carries no Name-a-card button: that is in the dock", () => {
    const m = mountModal(
      gameWith({ outside_play: { source_name: "Wish", can_name: true }, sideboard: [WISH] }),
      ME,
    );
    expect(m.c.querySelector('[data-testid="name-outside-card"]')).toBeNull();
  });

  it("is never offered for another seat, and shows an opponent nothing of it", () => {
    // The same seat seen by an opponent carries a count, no cards: no tab.
    const view = gameWith({ sideboard_count: 1 });
    expect(tabsOf(mountModal(view, THEM, "graveyard", ME).c)).toEqual([
      "graveyard",
      "exile",
      "command",
    ]);
    cleanup();
    expect(mountModal(view, THEM, "sideboard", ME).c.querySelectorAll(".cell").length).toBe(0);
  });
});
