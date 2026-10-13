import { describe, expect, it } from "vitest";

import type { GameMeta, SeatInfo } from "./api";
import {
  canRemoveSeat,
  REMOVE_SEAT_CONSEQUENCE,
  removeSeatConsequence,
  removeSeatQuestion,
} from "./seatRemoval";

// #2785: who is shown the Remove control. The server is the gate; this
// only has to agree with it often enough that nobody is offered a
// button that always fails.

function seat(id: string, extra: Partial<SeatInfo> = {}): SeatInfo {
  return { player_id: id, name: id, seat: 0, deck_uploaded: false, ...extra };
}

function game(extra: Partial<GameMeta> = {}): GameMeta {
  return { state: "lobby", players: [], ...extra } as unknown as GameMeta;
}

const host = seat("host", { is_host: true });
const bob = seat("bob");
const carol = seat("carol");
const bot = seat("bot", { is_bot: true });
const agent = seat("agent", { is_agent: true, agent_client: "claude-code" });

describe("canRemoveSeat", () => {
  it("is offered to the host on another person's seat", () => {
    expect(canRemoveSeat(game(), bob, { isAdmin: false, mySeat: host })).toBe(true);
  });

  it("is offered on an agent guest seat too", () => {
    expect(canRemoveSeat(game(), agent, { isAdmin: false, mySeat: host })).toBe(true);
  });

  it("is offered to an admin with no seat", () => {
    expect(canRemoveSeat(game(), bob, { isAdmin: true, mySeat: null })).toBe(true);
  });

  it("is offered to the creator whether or not they sit at the table", () => {
    const g = game({ is_creator: true });
    expect(canRemoveSeat(g, bob, { isAdmin: false, mySeat: null })).toBe(true);
    expect(canRemoveSeat(g, bob, { isAdmin: false, mySeat: carol })).toBe(true);
  });

  it("is hidden from an ordinary seated player", () => {
    expect(canRemoveSeat(game(), bob, { isAdmin: false, mySeat: carol })).toBe(false);
  });

  it("is hidden from a spectator or a signed-in visitor with no seat", () => {
    expect(canRemoveSeat(game(), bob, { isAdmin: false, mySeat: null })).toBe(false);
  });

  it("is never on the host's seat, even for an admin", () => {
    expect(canRemoveSeat(game(), host, { isAdmin: true, mySeat: null })).toBe(false);
  });

  it("is never on a bot, which has its own control", () => {
    expect(canRemoveSeat(game(), bot, { isAdmin: true, mySeat: host })).toBe(false);
  });

  it("is never on your own seat", () => {
    const g = game({ is_creator: true });
    expect(canRemoveSeat(g, carol, { isAdmin: false, mySeat: carol })).toBe(false);
  });

  it("is hidden once the game has ended", () => {
    const g = game({ state: "ended" });
    expect(canRemoveSeat(g, bob, { isAdmin: true, mySeat: host })).toBe(false);
  });

  // #3093: a started game takes the admin alone.
  it("is offered on a started game to an admin only", () => {
    const g = game({ state: "active", is_creator: true });
    expect(canRemoveSeat(g, bob, { isAdmin: true, mySeat: null })).toBe(true);
    expect(canRemoveSeat(g, bob, { isAdmin: true, mySeat: carol })).toBe(true);
    expect(canRemoveSeat(g, bob, { isAdmin: false, mySeat: host })).toBe(false);
    expect(canRemoveSeat(g, bob, { isAdmin: false, mySeat: null })).toBe(false);
  });

  it("keeps the host, bot and own-seat exclusions on a started game", () => {
    const g = game({ state: "active" });
    expect(canRemoveSeat(g, host, { isAdmin: true, mySeat: null })).toBe(false);
    expect(canRemoveSeat(g, bot, { isAdmin: true, mySeat: null })).toBe(false);
    expect(canRemoveSeat(g, carol, { isAdmin: true, mySeat: carol })).toBe(false);
  });
});

describe("the confirm copy", () => {
  it("names the person and says the invite still works", () => {
    expect(removeSeatQuestion("Bob")).toBe("Remove Bob from this table?");
    expect(REMOVE_SEAT_CONSEQUENCE).toMatch(/disconnected/);
    expect(REMOVE_SEAT_CONSEQUENCE).toMatch(/invite link still works/);
  });

  it("says a started game concedes the player out", () => {
    expect(removeSeatConsequence(game({ state: "lobby" }))).toBe(REMOVE_SEAT_CONSEQUENCE);
    expect(removeSeatConsequence(game({ state: "active" }))).toMatch(
      /^This concedes the player out of the game\./,
    );
  });
});
