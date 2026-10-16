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

  // ADR 0106 §4 decision 6 (#1806): a live "can't be countered" grant
  // or an unspent promise is one badge, each line in the tooltip.
  it("shows a can't-be-countered badge listing each grant and its source", () => {
    const s = {
      ...seat(),
      counter_shields: [
        { source_name: "Veil of Summer", text: "Spells you control can't be countered this turn." },
        {
          source_name: "Insist",
          text: "The next creature spell you cast this turn can't be countered.",
          next_only: true,
        },
      ],
    } as PlayerView;
    const { container } = render(PlayerIdentity as never, props(s) as never);
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].text).toBe("NO COUNTER");
    expect(got[0].title).toContain("can't be countered this turn. — Veil of Summer");
    expect(got[0].title).toContain("— Insist");
  });
});

// ADR 0113 §3, owner decision 1 (#2074, CR 402.2): a seat whose
// effective maximum hand size is not seven shows it, from the existing
// max_hand_size field. Seven, or no field at all, shows nothing.
describe("the seat's maximum hand size badge", () => {
  const withMax = (max?: number): PlayerView => ({ ...seat(), max_hand_size: max }) as PlayerView;

  it("is absent at the default seven", () => {
    const { container } = render(PlayerIdentity as never, props(withMax(7)) as never);
    expect(badges(container)).toHaveLength(0);
  });

  it("is absent when the view carries no maximum", () => {
    const { container } = render(PlayerIdentity as never, props(withMax(undefined)) as never);
    expect(badges(container)).toHaveLength(0);
  });

  it("shows HAND MAX 0 under Jin-Gitaxias", () => {
    const { container } = render(PlayerIdentity as never, props(withMax(0)) as never);
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].text).toBe("HAND MAX 0");
    expect(got[0].title).toContain("Maximum hand size 0");
  });

  it("shows a raised maximum too", () => {
    const { container } = render(PlayerIdentity as never, props(withMax(11)) as never);
    expect(badges(container).map((b) => b.text)).toEqual(["HAND MAX 11"]);
  });

  it("shows NO HAND MAX for no maximum hand size", () => {
    const { container } = render(PlayerIdentity as never, props(withMax(-1)) as never);
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].text).toBe("NO HAND MAX");
    expect(got[0].title).toContain("No maximum hand size");
  });

  it("renders on your own seat as well as an opponent's", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(withMax(2), { isSelf: true }) as never,
    );
    expect(badges(container).map((b) => b.text)).toEqual(["HAND MAX 2"]);
  });
});

// ADR 0199 (#3953, CR 701.20a): a seat that plays with its hand revealed
// says so, and names why, so the table knows why its hand is face up.
describe("the seat's revealed-hand badge", () => {
  const revealed = (sources?: string[]): PlayerView =>
    ({ ...seat(), hand_revealed: sources }) as PlayerView;

  it("is absent on a hidden hand", () => {
    const { container } = render(PlayerIdentity as never, props(revealed(undefined)) as never);
    expect(badges(container)).toHaveLength(0);
  });

  it("shows HAND OPEN with every source in its tooltip", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(revealed(["Telepathy", "Stromgald Spy"])) as never,
    );
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].text).toBe("HAND OPEN");
    expect(got[0].title).toBe("Plays with their hand revealed — Telepathy, Stromgald Spy");
  });
});

// #4016 (CR 121.2b, CR 121.3): a seat under a draw limit says so, names
// the permanents imposing it, and says how many draws are left, so a
// draw that does nothing (or a "you may draw" never asked) has a reason.
describe("the seat's draw-limit badge", () => {
  const limited = (draw_limit?: PlayerView["draw_limit"]): PlayerView =>
    ({ ...seat(), draw_limit }) as PlayerView;

  it("is absent when no limit applies", () => {
    const { container } = render(PlayerIdentity as never, props(limited(undefined)) as never);
    expect(badges(container)).toHaveLength(0);
  });

  it("shows DRAW MAX 1 with the source and the draws left", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(limited({ max: 1, left: 0, sources: ["Narset, Parter of Veils"] })) as never,
    );
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].text).toBe("DRAW MAX 1");
    expect(got[0].title).toBe(
      "Can't draw more than one card each turn — Narset, Parter of Veils. No draws left this turn.",
    );
  });

  it("counts a draw still to come", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(limited({ max: 1, left: 1, sources: ["Spirit of the Labyrinth"] })) as never,
    );
    expect(badges(container)[0].title).toBe(
      "Can't draw more than one card each turn — Spirit of the Labyrinth. 1 draw left this turn.",
    );
  });

  it("shows NO DRAW for a limit of zero", () => {
    const { container } = render(
      PlayerIdentity as never,
      props(limited({ max: 0, left: 0, sources: ["Maralen of the Mornsong"] })) as never,
    );
    const got = badges(container);
    expect(got).toHaveLength(1);
    expect(got[0].text).toBe("NO DRAW");
    expect(got[0].title).toBe("Can't draw cards — Maralen of the Mornsong");
  });
});

// #3262: two text badges beside the life chip widened its centred row
// until the life total slid out of the panel. The badges now hang on
// their own line under the chip, and the avatar makes room for them;
// the chip itself holds only the number. Layout is the CSS's job, so
// this pins the structure the CSS relies on.
describe("badges and the life total (#3262)", () => {
  it("keeps the life number in its own chip and gives the badges their own line", () => {
    const s = { ...seat(), cant_gain_life: true, max_hand_size: -1 } as PlayerView;
    const { container } = render(PlayerIdentity as never, props(s) as never);
    expect(badges(container).map((b) => b.text)).toEqual(["NO GAIN", "NO HAND MAX"]);
    expect(container.querySelector(".life-chip .life")?.textContent).toBe("40");
    expect(container.querySelector(".life-chip .seat-keywords")).toBeNull();
    expect(container.querySelector(".avatar-wrap")?.classList.contains("has-keywords")).toBe(true);
  });

  it("makes no room under the avatar when there are no badges", () => {
    const { container } = render(PlayerIdentity as never, props(seat()) as never);
    expect(container.querySelector(".avatar-wrap")?.classList.contains("has-keywords")).toBe(false);
  });
});
