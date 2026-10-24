// @vitest-environment jsdom
//
// #2947 (CR 707.12, CR 608.2g): casting a copy of a card while the
// effect that made it is still resolving. The chooser is shown the
// copies and the server's ways of casting them (the legal moves naming
// the prompt), each sent exactly as the server wrote it, and "Don't
// cast" is always there.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, CardView, GameView, LegalMoveView } from "./protocol";
import { barPrimary, nameOf } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ritualCopy = {
  instance_id: "copy-1",
  name: "Dark Ritual",
  owner: "me",
  controller: "me",
  type_line: "Instant",
  known_by_you: true,
} as CardView;

const castMove = (label: string): LegalMoveView =>
  ({
    type: "cast_spell",
    player: "me",
    kind: "cast",
    label: `Kaervek, the Punisher: ${label}`,
    source: "copy-1",
    params: { instance_id: "copy-1", from_zone: "exile", resolution_cast: "rc-1" },
    resolution_cast: "rc-1",
  }) as LegalMoveView;

const declineMove = {
  type: "resolve_choice",
  player: "me",
  kind: "choice",
  label: "Kaervek, the Punisher: don't cast",
  params: { choice_id: "rc-1", card_ids: [] },
  always_legal: true,
  resolution_cast: "rc-1",
} as LegalMoveView;

interface CascadeOffer {
  keyword: string;
  reason: string;
  decline: string;
}

const snapWith = (moves: LegalMoveView[], chooser = "me", offer?: CascadeOffer): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "them", name: "Bob" },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 1, cards: [ritualCopy] },
    turn: { number: 3, active_seat: 0, priority_holder: -1, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "rc-1",
        kind: "resolution_cast",
        chooser,
        count: 1,
        reason: offer?.reason ?? "Kaervek, the Punisher",
        options: [ritualCopy],
        eligible: ["copy-1"],
        ...(offer
          ? {
              may_cast_keyword: offer.keyword,
              accept_label: "Cast it free",
              decline_label: offer.decline,
            }
          : {}),
      },
    ],
    legal_moves: moves,
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
  player?: string;
}

function mount(moves: LegalMoveView[], chooser = "me", offer?: CascadeOffer): { sent: Sent[] } {
  const sent: Sent[] = [];
  render(
    ChoiceDockHarness as never,
    {
      snap: snapWith(moves, chooser, offer),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown, player?: string) =>
        sent.push({ type, params, player }),
      lastError: null,
    } as never,
  );
  return { sent };
}

const dialog = (): Element | null =>
  document.querySelector('[role="dialog"][aria-label^="Kaervek, the Punisher"]');

describe("ResolutionCastSheet — cast a copy during the resolution (#2947)", () => {
  it("names the source and the copy", () => {
    mount([castMove("Cast Dark Ritual"), declineMove]);
    expect(dialog()).not.toBeNull();
    expect(nameOf(dialog()!)).toBe("Kaervek, the Punisher: you may cast the copy of Dark Ritual");
  });

  it("makes a lone cast the primary, sent exactly as the server wrote it", () => {
    const { sent } = mount([castMove("Cast Dark Ritual"), declineMove]);
    const primary = barPrimary();
    expect(primary && nameOf(primary)).toBe("Cast Dark Ritual");
    click(primary!);
    expect(sent).toEqual([
      {
        type: "cast_spell",
        player: "me",
        params: { instance_id: "copy-1", from_zone: "exile", resolution_cast: "rc-1" },
      },
    ]);
  });

  it('offers "Don\'t cast" beside the casts, and alone when none is possible', () => {
    const { sent } = mount([declineMove]);
    const primary = barPrimary();
    expect(primary && nameOf(primary)).toBe("Don't cast");
    click(primary!);
    expect(sent).toEqual([
      { type: "resolve_choice", player: "me", params: { choice_id: "rc-1", card_ids: [] } },
    ]);
  });

  it("lists several ways to cast, and sends the one clicked", () => {
    const { sent } = mount([
      castMove("Cast Dark Ritual"),
      castMove("Cast Dark Ritual (again)"),
      declineMove,
    ]);
    const rows = [...document.querySelectorAll<HTMLButtonElement>("ul.casts button")];
    expect(rows.map((b) => nameOf(b))).toEqual(["Cast Dark Ritual", "Cast Dark Ritual (again)"]);
    click(rows[1]);
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("cast_spell");
  });

  it("is not shown to another seat", () => {
    mount([], "them");
    expect(dialog()).toBeNull();
  });
});

describe("ResolutionCastSheet — cascade and discover (#4243)", () => {
  const cascadeOffer = { keyword: "cascade", reason: "Cascade", decline: "Put it on the bottom" };
  const discoverOffer = {
    keyword: "discover",
    reason: "Discover 4",
    decline: "Put it into your hand",
  };
  const hitDialog = (reason: string): Element | null =>
    document.querySelector(`[role="dialog"][aria-label^="${reason}"]`);

  it("names the exiled card and what the decline does", () => {
    mount([castMove("Cast Dark Ritual"), declineMove], "me", cascadeOffer);
    expect(nameOf(hitDialog("Cascade")!)).toBe(
      "Cascade: you may cast Dark Ritual without paying its mana cost",
    );
    expect(document.body.textContent).toContain("it goes to the bottom of your library");
    const buttons = [...document.querySelectorAll("button")].map((b) => nameOf(b));
    expect(buttons).toContain("Put it on the bottom");
    expect(buttons).not.toContain("Don't cast");
  });

  it("words discover's decline as the hand, and sends the decline as written", () => {
    const { sent } = mount([declineMove], "me", discoverOffer);
    expect(nameOf(hitDialog("Discover 4")!)).toBe(
      "Discover 4: you may cast Dark Ritual without paying its mana cost",
    );
    expect(document.body.textContent).toContain("it goes into your hand");
    const primary = barPrimary();
    expect(primary && nameOf(primary)).toBe("Put it into your hand");
    click(primary!);
    expect(sent).toEqual([
      { type: "resolve_choice", player: "me", params: { choice_id: "rc-1", card_ids: [] } },
    ]);
  });
});
