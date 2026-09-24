import { describe, it, expect } from "vitest";

import { endCueFor, gameOverText, winnerOf } from "./gameOutcome";
import { endGateBadges, playerKeywordBadges } from "./playerKeywordBadges";
import type { GameView, OutcomeView, PlayerView } from "./protocol";

// gameOutcome.test.ts — ADR 0057 (#749), test plan item 20: the winner
// and the sounds come from GameView.outcome, an ended view with no
// outcome falls back to the survivors, and the banner's cause text and
// the seat badge's labels are pure helpers.

function seat(i: number, name: string, extra: Partial<PlayerView> = {}): PlayerView {
  return {
    id: `p${i}`,
    name,
    seat: i,
    life: 40,
    library: { kind: "library", count: 60, cards: [] },
    hand: { kind: "hand", count: 7, cards: [] },
    graveyard: { kind: "graveyard", count: 0, cards: [] },
    command: { kind: "command", count: 0, cards: [] },
    commander_damage: {},
    life_history: [],
    ...extra,
  } as PlayerView;
}

function view(state: string, seats: PlayerView[], outcome?: OutcomeView): GameView {
  return { state, seats, outcome } as unknown as GameView;
}

const alice = seat(0, "Alice");
const bob = seat(1, "Bob");
const carol = seat(2, "Carol");

describe("winnerOf", () => {
  it("reads an effect win from the outcome while every seat is still standing", () => {
    const v = view("ended", [alice, bob, carol], {
      kind: "win",
      winner: "p1",
      winner_seat: 1,
      cause: "effect",
      source_name: "Felidar Sovereign",
    });
    expect(winnerOf(v)?.name).toBe("Bob");
  });

  it("has no winner for a draw, even with nobody standing", () => {
    const v = view(
      "ended",
      [seat(0, "A", { eliminated: true }), seat(1, "B", { eliminated: true })],
      {
        kind: "draw",
        cause: "all_lost",
      },
    );
    expect(winnerOf(v)).toBeNull();
  });

  it("falls back to the one seat left standing when an ended view has no outcome", () => {
    const v = view("ended", [seat(0, "A", { eliminated: true }), bob]);
    expect(winnerOf(v)?.name).toBe("Bob");
  });

  it("names nobody while the game is active", () => {
    expect(winnerOf(view("active", [alice]))).toBeNull();
  });
});

describe("gameOverText", () => {
  it("names the source of an effect win", () => {
    const v = view("ended", [alice, bob], {
      kind: "win",
      winner: "p0",
      cause: "effect",
      source_name: "Felidar Sovereign",
    });
    expect(gameOverText(v)).toEqual({ winner: alice, text: "wins the game — Felidar Sovereign" });
  });

  it("is plain for the last player standing", () => {
    const v = view("ended", [alice, seat(1, "Bob", { eliminated: true })], {
      kind: "win",
      winner: "p0",
      cause: "last_standing",
    });
    expect(gameOverText(v).text).toBe("wins the game.");
  });

  it("says a draw is a draw", () => {
    const v = view("ended", [alice, bob], { kind: "draw", cause: "all_lost" });
    expect(gameOverText(v)).toEqual({ winner: null, text: "The game is a draw." });
  });

  it("keeps the old sentence for an ended view with nobody to name", () => {
    const v = view("ended", [alice, bob]);
    expect(gameOverText(v)).toEqual({ winner: null, text: "Game ended — no survivors." });
  });
});

describe("endCueFor", () => {
  const won = view("ended", [alice, bob, carol], { kind: "win", winner: "p2", cause: "effect" });
  it("plays the win for the winner and the loss for everyone else", () => {
    expect(endCueFor(won, "p2")).toBe("win");
    expect(endCueFor(won, "p0")).toBe("loss");
  });
  it("is silent for a spectator", () => {
    expect(endCueFor(won, null)).toBeNull();
  });
  it("is a loss for everyone in a draw", () => {
    expect(endCueFor(view("ended", [alice, bob], { kind: "draw", cause: "all_lost" }), "p0")).toBe(
      "loss",
    );
  });
});

describe("endGateBadges", () => {
  it("labels a fully gated seat and names the source", () => {
    const badges = endGateBadges({
      cant_lose: ["life", "empty_draw", "poison", "commander_damage", "effect"],
      end_gates: [
        {
          source_name: "Platinum Angel",
          cant_lose: ["life", "empty_draw", "poison", "commander_damage", "effect"],
        },
      ],
    });
    expect(badges.map((b) => b.short)).toEqual(["CAN'T LOSE"]);
    expect(badges[0].title).toBe("Can't lose the game — Platinum Angel. Conceding still loses.");
  });

  it("says which causes a partial gate stops", () => {
    const [b] = endGateBadges({
      cant_lose: ["life"],
      end_gates: [{ source_name: "Phyrexian Unlife", cant_lose: ["life"] }],
    });
    expect(b.title).toBe(
      "Can't lose the game to 0 or less life — Phyrexian Unlife. Conceding still loses.",
    );
  });

  it("labels a seat that can't win, marking a this-turn grant", () => {
    const badges = endGateBadges({
      cant_win: true,
      end_gates: [{ source_name: "Angel's Grace", cant_win: true, this_turn: true }],
    });
    expect(badges.map((b) => b.short)).toEqual(["CAN'T WIN"]);
    expect(badges[0].title).toBe("Can't win the game — Angel's Grace (this turn)");
  });

  it("is empty for an ungated seat, and rides after the keyword badges", () => {
    expect(endGateBadges({})).toEqual([]);
    const all = playerKeywordBadges(["hexproof"], true, { cant_win: true });
    expect(all.map((b) => b.key)).toEqual(["hexproof", "life-total-locked", "cant-win"]);
  });
});
