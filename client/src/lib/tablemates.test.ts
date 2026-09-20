import { afterEach, describe, expect, it, vi } from "vitest";
import {
  canInviteTablemates,
  inviteSentMessage,
  tablemateSubtitle,
  type Tablemate,
} from "./tablemates";
import { fetchTablemates, sendInviteDM } from "./api";
import { setSession, type Session } from "./session";
import { ZERO_USER_ID } from "./myDecks";

const REAL_USER = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const GAME = "11111111-2222-3333-4444-555555555555";

function mate(partial: Partial<Tablemate> = {}): Tablemate {
  return {
    user_id: REAL_USER,
    display_name: "Bob",
    last_played_at: Date.parse("2026-09-18T20:00:00Z"),
    ...partial,
  };
}

function sess(partial: Partial<Session["principal"]> = {}, rest: Partial<Session> = {}): Session {
  return {
    token: "t",
    expiresAt: "2026-12-01T00:00:00Z",
    principal: {
      role: "player",
      user_id: REAL_USER,
      issued_at: "2026-09-19T00:00:00Z",
      expires_at: "2026-12-01T00:00:00Z",
      ...partial,
    },
    ...rest,
  };
}

describe("canInviteTablemates", () => {
  it("offers the picker to a signed-in player seated at this table", () => {
    expect(canInviteTablemates(sess(), GAME, GAME)).toBe(true);
  });

  it("does not offer it for a table the player is not seated at", () => {
    expect(canInviteTablemates(sess(), "another-game", GAME)).toBe(false);
    expect(canInviteTablemates(sess(), undefined, GAME)).toBe(false);
  });

  // #1154: it used to. An admin session carries no user_id, so
  // GET /me/tablemates refuses it and the picker could only ever be
  // empty — and mounting it cost the admin their session, because the
  // refusal was a 401 and authFetch clears the session on any 401.
  it("does not offer it to the admin, who has no tablemates", () => {
    expect(canInviteTablemates(sess({ role: "admin", user_id: ZERO_USER_ID }), null, GAME)).toBe(
      false,
    );
    expect(canInviteTablemates(sess({ role: "admin", user_id: undefined }), GAME, GAME)).toBe(
      false,
    );
  });

  it("does not offer it to a guest, whose zero user_id is not a person", () => {
    expect(canInviteTablemates(sess({ user_id: ZERO_USER_ID }), GAME, GAME)).toBe(false);
    expect(canInviteTablemates(sess({ user_id: undefined }), GAME, GAME)).toBe(false);
  });

  it("does not offer it with no session at all", () => {
    expect(canInviteTablemates(null, GAME, GAME)).toBe(false);
    expect(canInviteTablemates(undefined, GAME, GAME)).toBe(false);
  });
});

describe("tablemateSubtitle", () => {
  const now = Date.parse("2026-09-19T12:00:00Z");

  it("says today for a table shared within the day", () => {
    expect(tablemateSubtitle(mate({ last_played_at: now - 3 * 3600_000 }), now)).toBe(
      "played today",
    );
  });

  it("says yesterday, then counts days", () => {
    expect(tablemateSubtitle(mate({ last_played_at: now - 86_400_000 }), now)).toBe(
      "played yesterday",
    );
    expect(tablemateSubtitle(mate({ last_played_at: now - 5 * 86_400_000 }), now)).toBe(
      "played 5 days ago",
    );
  });

  it("rolls up to months and years", () => {
    expect(tablemateSubtitle(mate({ last_played_at: now - 40 * 86_400_000 }), now)).toBe(
      "played a month ago",
    );
    expect(tablemateSubtitle(mate({ last_played_at: now - 200 * 86_400_000 }), now)).toBe(
      "played 6 months ago",
    );
    expect(tablemateSubtitle(mate({ last_played_at: now - 400 * 86_400_000 }), now)).toBe(
      "played a year ago",
    );
  });

  it("degrades gracefully when the server sent no time", () => {
    expect(tablemateSubtitle(mate({ last_played_at: 0 }), now)).toBe("played together");
  });
});

describe("inviteSentMessage", () => {
  it("prefers the name the server echoed", () => {
    expect(inviteSentMessage({ sent: true, display_name: "Bob" }, "stale")).toBe(
      "invite sent to Bob",
    );
  });

  it("falls back to the name the picker already had", () => {
    expect(inviteSentMessage({ sent: true }, "Bob")).toBe("invite sent to Bob");
  });

  it("says something even with no name at all", () => {
    expect(inviteSentMessage({ sent: true }, "")).toBe("invite sent");
  });
});

// The two api.ts wrappers: pin the URL, method and body, the way
// adminLobby.test.ts does for the rest of the surface.
describe("the tablemate API calls", () => {
  const calls: { url: string; method: string; body?: unknown }[] = [];

  function stubFetch(body: unknown): void {
    calls.length = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: string, init: RequestInit = {}) => {
        calls.push({
          url: input,
          method: init.method ?? "GET",
          body: typeof init.body === "string" ? JSON.parse(init.body) : undefined,
        });
        return {
          ok: true,
          status: 200,
          statusText: "stub",
          json: async () => body,
          clone() {
            return this;
          },
        } as unknown as Response;
      }),
    );
  }

  afterEach(() => {
    vi.unstubAllGlobals();
    setSession(null);
  });

  it("fetchTablemates reads /me/tablemates and tolerates an absent list", async () => {
    stubFetch({ tablemates: [mate()] });
    expect(await fetchTablemates()).toHaveLength(1);
    expect(calls[0]).toMatchObject({ url: "/me/tablemates", method: "GET" });

    stubFetch({});
    expect(await fetchTablemates()).toEqual([]);
  });

  it("sendInviteDM posts our user id, never a snowflake", async () => {
    stubFetch({ sent: true, user_id: REAL_USER, display_name: "Bob" });
    const res = await sendInviteDM(GAME, REAL_USER);
    expect(res.sent).toBe(true);
    expect(calls[0]).toMatchObject({
      url: `/games/${GAME}/invites/dm`,
      method: "POST",
      body: { user_id: REAL_USER },
    });
  });
});
