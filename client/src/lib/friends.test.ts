// ADR 0213 slice 3: the words and decisions of lib/friends.ts.

import { describe, expect, it } from "vitest";

import {
  canBefriendSeat,
  friendControlText,
  friendErrorLine,
  friendStateOf,
  outcomeLine,
  playedWithRest,
  sinceLine,
  stateAfter,
} from "./friends";
import { L } from "./labels";
import { LobbyApiError } from "./session";
import type { Tablemate } from "./tablemates";

const NOW = Date.UTC(2026, 10, 5, 12);
const DAY = 86_400_000;

function mate(extra: Partial<Tablemate> = {}): Tablemate {
  return { user_id: "u", display_name: "Bob", last_played_at: NOW, ...extra };
}

describe("a row's friend control", () => {
  it("follows the row's friend and request marks", () => {
    expect(friendStateOf(mate())).toBe("none");
    expect(friendStateOf(mate({ request: "incoming" }))).toBe("incoming");
    expect(friendStateOf(mate({ request: "outgoing" }))).toBe("outgoing");
    expect(friendStateOf(mate({ friend: true, request: "incoming" }))).toBe("friend");
  });

  it("says Add friend, Accept, or where it stands", () => {
    expect(friendControlText("none")).toBe(L.addFriend);
    expect(friendControlText("incoming")).toBe(L.acceptFriend);
    expect(friendControlText("outgoing")).toBe("Request sent");
    expect(friendControlText("friend")).toBe("Friends");
  });

  it("moves on from an answer", () => {
    expect(stateAfter("sent")).toBe("outgoing");
    expect(stateAfter("pending")).toBe("outgoing");
    expect(stateAfter("friends")).toBe("friend");
    expect(stateAfter("already_friends")).toBe("friend");
  });
});

describe("the lines", () => {
  it("says what asking did", () => {
    expect(outcomeLine("sent", "Bob")).toBe("Friend request sent to Bob");
    expect(outcomeLine("pending", "Bob")).toBe("Your request to Bob is still waiting");
    expect(outcomeLine("friends", "Bob")).toBe("You and Bob are friends now");
    expect(outcomeLine("already_friends", "Bob")).toBe("You and Bob are already friends");
  });

  it("dates each list's rows", () => {
    expect(sinceLine("friend", NOW - 3 * DAY, NOW)).toBe("friends since 3 days ago");
    expect(sinceLine("incoming", NOW, NOW)).toBe("asked you today");
    expect(sinceLine("outgoing", NOW - DAY, NOW)).toBe("you asked yesterday");
    expect(sinceLine("blocked", NOW - 400 * DAY, NOW)).toBe("blocked a year ago");
    expect(sinceLine("friend", 0, NOW)).toBe("friends");
  });

  it("turns a refusal into a sentence", () => {
    expect(
      friendErrorLine(new LobbyApiError(403, "can't send a friend request to that person")),
    ).toBe("Can't send a friend request to that person.");
    expect(friendErrorLine(new LobbyApiError(503, "anything"))).toBe(
      "Friends aren’t available on this server.",
    );
    expect(friendErrorLine(new Error("network"))).toBe("That didn’t work. Try again.");
  });
});

describe("who is offered", () => {
  it("Played with leaves out your friends, who are listed above it", () => {
    const rest = playedWithRest([mate({ user_id: "a" }), mate({ user_id: "b", friend: true })]);
    expect(rest.map((m) => m.user_id)).toEqual(["a"]);
  });

  it("a seat offers Add friend only for another signed-in person at your table", () => {
    const person = { seat: 2, discord_id: "123" };
    expect(canBefriendSeat(person, 0)).toBe(true);
    expect(canBefriendSeat(person, null)).toBe(false); // you don't sit there
    expect(canBefriendSeat({ ...person, seat: 0 }, 0)).toBe(false); // you
    expect(canBefriendSeat({ seat: 3 }, 0)).toBe(false); // a guest
    expect(canBefriendSeat({ ...person, is_bot: true }, 0)).toBe(false);
    expect(canBefriendSeat({ ...person, is_agent: true }, 0)).toBe(false);
  });
});
