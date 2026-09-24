// @vitest-environment jsdom
//
// #1201, the client half of #1197's player protection and hexproof.
// PlayerView.keywords (docs/protocol.md) has been on the wire since
// #1197 with no badge to show it: a player whose Lightning Bolt finds
// no legal target in a Leyline of Sanctity seat sees the right
// refusal with no visible reason. This is the badge.
//
// Two things have to be in the markup and neither can be checked
// without a DOM — that a hexproof seat gets a badge and a plain seat
// gets none, and that the badge names the source (hexproof, or the
// protected quality) rather than just showing a raw wire token.

import { describe, it, expect, afterEach } from "vitest";

import PlayerIdentity from "./components/board/PlayerIdentity.svelte";
import type { PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seat = (keywords?: PlayerView["keywords"], lifeTotalLocked?: boolean): PlayerView =>
  ({
    id: "me",
    name: "Me",
    seat: 0,
    life: 40,
    library: { kind: "library", count: 0, cards: [] },
    hand: { kind: "hand", count: 0, cards: [] },
    graveyard: { kind: "graveyard", count: 0, cards: [] },
    command: { kind: "command", count: 0, cards: [] },
    commander_damage: {},
    life_history: [],
    keywords,
    life_total_locked: lifeTotalLocked,
  }) as unknown as PlayerView;

const props = (s: PlayerView, over: Record<string, unknown> = {}) => ({
  seat: s,
  isSelf: false,
  isActive: false,
  hasPriority: false,
  attackTargetable: false,
  isMonarch: false,
  isInitiative: false,
  sendAction: () => {},
  ...over,
});

function badges(container: HTMLElement): { title: string; text: string }[] {
  return [...container.querySelectorAll(".seat-keywords .kw-badge")].map((el) => ({
    title: el.getAttribute("title") ?? "",
    text: (el.textContent ?? "").trim(),
  }));
}

describe("the seat keyword badge", () => {
  it("is absent on a plain seat", () => {
    const { container } = render(PlayerIdentity as never, props(seat()) as never);
    expect(container.querySelectorAll(".seat-keywords")).toHaveLength(0);
    expect(badges(container)).toHaveLength(0);
  });

  it("is absent on a seat with an empty keywords list", () => {
    const { container } = render(PlayerIdentity as never, props(seat([])) as never);
    expect(badges(container)).toHaveLength(0);
  });

  it("shows a hexproof badge naming Hexproof", () => {
    const { container } = render(PlayerIdentity as never, props(seat(["hexproof"])) as never);
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].title).toBe("Hexproof");
  });

  it("shows a protection badge that titles the quality rather than the raw token", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat(["protection from everything"])) as never,
    );
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].title).toBe("Protection from Everything");
    expect(got[0].text).toBe("ALL");
    // Never the raw lowercase wire token as the visible face.
    expect(container.querySelector(".seat-keywords")?.textContent).not.toContain(
      "protection from everything",
    );
  });

  it("abbreviates a quality other than the two catalogued today, since the token is general", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat(["protection from red"])) as never,
    );
    const got = badges(container);
    expect(got).toEqual([{ title: "Protection from Red", text: "RED" }]);
  });

  it("renders both a hexproof and a protection badge when a seat somehow has both", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat(["hexproof", "protection from everything"])) as never,
    );
    expect(badges(container)).toHaveLength(2);
  });

  it("dedupes two sources granting the same token (two Leylines)", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat(["hexproof", "hexproof"])) as never,
    );
    expect(badges(container)).toHaveLength(1);
  });

  it("renders on an opponent's seat too — protection and hexproof are public", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat(["hexproof"]), { isSelf: false }) as never,
    );
    expect(badges(container)).toHaveLength(1);
  });

  // #1200 (CR 119.7, CR 119.8): the life-total lock arrives as its own
  // bool rather than as a keyword token, because it is not an ability
  // the player has — see ADR 0085 Decision 7.
  it("shows a life-total-lock badge on a locked seat", () => {
    const { container } = render(PlayerIdentity as never, props(seat(undefined, true)) as never);
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].text).toBe("LIFE");
    expect(got[0].title).toContain("can't change");
  });

  it("shows no lock badge on a seat whose life total can change", () => {
    const { container } = render(PlayerIdentity as never, props(seat(undefined, false)) as never);
    expect(badges(container)).toHaveLength(0);
  });

  it("renders the lock alongside the protection Teferi's Protection grants with it", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(seat(["protection from everything"], true)) as never,
    );
    expect(badges(container).map((b) => b.text)).toEqual(["ALL", "LIFE"]);
  });
});
