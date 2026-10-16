// @vitest-environment jsdom
//
// sideboard.render.test.ts — ADR 0191. A deck's sideboard is the cards
// you own outside the game. The server sends the cards to their owner
// alone (a count to everyone else), and the board offers them as a pile
// that opens the read-only zone browser. These tests pin the three
// things a player would notice: it is yours alone, it opens, and nothing
// in it can be moved.

import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { get } from "svelte/store";

import PileBar from "./components/board/PileBar.svelte";
import ZoneBrowserModal from "./components/board/ZoneBrowserModal.svelte";
import type { ActionType, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { closeZoneBrowser, zoneBrowser } from "./zoneBrowser";
import { buildMovePayload, canManageZone, cardsForZone } from "./zoneBrowser.logic";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  closeZoneBrowser();
});
beforeEach(() => closeZoneBrowser());

const ME = "me";
const THEM = "them";

const zone = (kind: string, count = 0): ZoneView => ({ kind, count, cards: [] });

function sideCard(id: string, name: string): CardView {
  return {
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    type_line: "Sorcery",
    known_by_you: true,
  } as unknown as CardView;
}

const WISH = sideCard("s1", "Burning Wish");
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

  it("reads the owner's own cards and nothing for a seat that sent none", () => {
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

describe("the sideboard pile", () => {
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

  it("is offered on your own seat when it holds cards, face down", () => {
    const c = mountBar({ sideboard: [WISH, BOLT], sideboard_count: 2 });
    const pile = c.querySelector<HTMLElement>('[data-pile="sideboard"]');
    expect(pile, "a sideboard pile on your own seat").not.toBeNull();
    expect(pile!.getAttribute("aria-label")).toContain("sideboard: 2 cards");
    // #3832: "outside the game" explains the pile in its tooltip and
    // accessible name; it is not text on the tile, where it overflowed.
    expect(pile!.getAttribute("aria-label")).toContain("outside the game");
    expect(pile!.getAttribute("title")).toContain("outside the game");
    expect(pile!.textContent).not.toContain("outside the game");
    // Face down: no card name is rendered on the tile.
    expect(pile!.textContent).not.toContain("Burning Wish");
  });

  it("opens the read-only sideboard browser", () => {
    const c = mountBar({ sideboard: [WISH, BOLT], sideboard_count: 2 });
    click(c.querySelector<HTMLElement>('[data-pile="sideboard"]')!);
    expect(get(zoneBrowser)).toEqual({ zoneKind: "sideboard", ownerID: ME, ownerName: "Me" });
  });

  it("is absent with no sideboard", () => {
    const c = mountBar({});
    expect(c.querySelector('[data-pile="sideboard"]')).toBeNull();
  });

  it("is never drawn for another seat, even with a count", () => {
    const c = mountBar({ sideboard_count: 5 }, false);
    expect(c.querySelector('[data-pile="sideboard"]')).toBeNull();
  });

  it("is not drawn for another seat even if cards somehow arrive on it", () => {
    // The server never sends them; the tile is still yours alone.
    const c = mountBar({ sideboard: [WISH], sideboard_count: 1 }, false);
    expect(c.querySelector('[data-pile="sideboard"]')).toBeNull();
  });
});

describe("the sideboard browser", () => {
  function mountModal(view: GameView, viewerID: string | null, ownerID = ME) {
    const sent: ActionType[] = [];
    const r = render(
      ZoneBrowserModal as never,
      {
        view,
        viewerID,
        zoneKind: "sideboard",
        ownerSeat: { id: ownerID, name: ownerID === ME ? "Me" : "Them" },
        sendAction: (t: ActionType) => sent.push(t),
        onClose: () => {},
      } as never,
    );
    return { c: r.container, sent };
  }

  it("lists the cards, says it is read only and explains what the zone is", () => {
    const m = mountModal(gameWith({ sideboard: [WISH, BOLT], sideboard_count: 2 }), ME);
    const text = m.c.textContent ?? "";
    expect(text).toContain("sideboard");
    expect(text).toContain("(2)");
    expect(text).toContain("read only");
    expect(text).toContain("outside the game");
    expect(m.c.querySelectorAll(".cell").length).toBe(2);
    // No move buttons: nothing in it can be moved.
    expect(m.c.querySelector("button.move, [data-move]")).toBeNull();
    expect(m.sent).toEqual([]);
  });

  it("offers the sideboard as a tab on your own seat only", () => {
    const own = mountModal(gameWith({ sideboard: [WISH], sideboard_count: 1 }), ME);
    const tabs = [...own.c.querySelectorAll<HTMLElement>('[role="tab"]')].map((t) =>
      t.textContent?.trim(),
    );
    expect(tabs).toEqual(["graveyard", "exile", "command", "sideboard"]);
    cleanup();
    // The same seat seen by an opponent carries no cards, so no tab.
    const other = mountModal(gameWith({ sideboard_count: 1 }), THEM, ME);
    const otherTabs = [...other.c.querySelectorAll<HTMLElement>('[role="tab"]')].map((t) =>
      t.textContent?.trim(),
    );
    expect(otherTabs).toEqual(["graveyard", "exile", "command"]);
  });
});
