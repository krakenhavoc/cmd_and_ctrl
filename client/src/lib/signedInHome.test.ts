import { describe, expect, it } from "vitest";

import type { GameMeta } from "./api";
import { parseHash, type Route } from "./router";
import type { Session } from "./session";
import {
  GUEST_CODE_MESSAGE,
  inviteHash,
  isPublicRoute,
  joinBoxFor,
  myTables,
  oauthCompleteTarget,
  routeRedirect,
} from "./signedInHome";

// ADR 0112 §1, the signed-in home: who stays on #/login, where the
// Discord round trip lands, which join box the Lobby offers, and which
// tables it lists. All pure, so the whole decision table is pinned here
// and the components only render the answers.

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const NIL = "00000000-0000-0000-0000-000000000000";

type Role = Session["principal"]["role"];

function sess(
  role: Role,
  userID?: string,
  opts: { gameID?: string; admin?: boolean } = {},
): Session {
  const gameID = role === "identified" || role === "admin" ? undefined : (opts.gameID ?? "g1");
  return {
    token: `${role}-tok`,
    expiresAt: new Date(Date.now() + 86_400_000).toISOString(),
    principal: {
      role,
      user_id: userID,
      name: "Alice",
      game_id: gameID,
      player_id: role === "player" ? "p1" : undefined,
      issued_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 86_400_000).toISOString(),
    },
    gameID,
    playerID: role === "player" ? "p1" : undefined,
    admin: opts.admin,
    // What /me says about an allowlisted person (ADR 0112 §2 item 9),
    // here in admin mode.
    ...(opts.admin
      ? { admin_allowed: true, admin_mode: true, admin_mode_ends_at: Date.now() + 86_400_000 }
      : {}),
  };
}

// Every kind of session the router can hold, named for the table in
// ADR 0112 §1 item 1.
const SESSIONS: Array<[string, Session | null]> = [
  ["none", null],
  ["identified, with a user", sess("identified", USER)],
  ["identified, no user database", sess("identified")],
  ["signed-in player", sess("player", USER)],
  ["signed-in spectator", sess("spectator", USER)],
  ["guest player", sess("player", NIL)],
  ["guest spectator", sess("spectator")],
  ["admin token", sess("admin")],
  ["allowlisted admin person", sess("player", USER, { admin: true })],
];

describe("routeRedirect: Login is for signed-out visitors only (§1 item 1)", () => {
  const login = parseHash("#/login");

  it("keeps a signed-out visitor on #/login", () => {
    expect(routeRedirect(login, null)).toBeNull();
  });

  for (const [name, s] of SESSIONS) {
    if (s === null) continue;
    it(`sends ${name} on #/login to the Lobby`, () => {
      expect(routeRedirect(login, s)).toBe("#/lobby");
    });
  }

  it("sends a signed-out visitor on a gated route to #/login", () => {
    for (const hash of ["#/lobby", "#/games/g1", "#/catalog", "#/my-games", "#/practice"]) {
      expect(routeRedirect(parseHash(hash), null)).toBe("#/login");
    }
  });

  it("leaves a session on a gated route alone", () => {
    for (const [, s] of SESSIONS) {
      if (s === null) continue;
      expect(routeRedirect(parseHash("#/lobby"), s)).toBeNull();
      expect(routeRedirect(parseHash("#/games/g1"), s)).toBeNull();
    }
  });
});

describe("routeRedirect: the ways in that do not go through Login (§1 item 2)", () => {
  // An invite link, a spectator link, a reclaim link and both shapes of
  // the Discord round trip are reachable whatever the browser holds:
  // none of them may be bounced, signed out or signed in.
  const ENTRIES: Array<[string, Route]> = [
    ["an invite link", parseHash("#/games/g2/join?t=invite-tok")],
    ["a spectator link", parseHash("#/games/g2/join?t=spec-tok&spectator=1")],
    ["a reclaim link", parseHash("#/games/g2/reclaim?t=ticket")],
    [
      "OAuth from an invite",
      parseHash("#/oauth-complete?token=t&expires_at=2099-01-01T00:00:00Z&game=g2&player_id=p2"),
    ],
    [
      "OAuth from the login page",
      parseHash("#/oauth-complete?token=t&expires_at=2099-01-01T00:00:00Z"),
    ],
  ];

  it("parses each entry to its own route, not to login", () => {
    expect(ENTRIES.map(([, r]) => r.name)).toEqual([
      "join",
      "join",
      "reclaim",
      "oauthComplete",
      "oauthComplete",
    ]);
    expect(ENTRIES[1][1]).toMatchObject({ name: "join", spectator: true });
  });

  for (const [entry, r] of ENTRIES) {
    for (const [name, s] of SESSIONS) {
      it(`opens ${entry} for ${name}`, () => {
        expect(isPublicRoute(r)).toBe(true);
        expect(routeRedirect(r, s)).toBeNull();
      });
    }
  }
});

describe("routeRedirect: #/admin (§2 item 8, the sessions told apart without /me)", () => {
  const admin = parseHash("#/admin");

  it("shows the token form signed out", () => {
    expect(routeRedirect(admin, null)).toBeNull();
  });

  it("sends the token session itself to the Lobby", () => {
    expect(routeRedirect(admin, sess("admin"))).toBe("#/lobby");
  });

  it("sends an allowlisted person to the Lobby, where the chip is, in either mode", () => {
    const inAdminMode = sess("player", USER, { admin: true });
    expect(routeRedirect(admin, inAdminMode)).toBe("#/lobby");
    const inPlayerMode = {
      ...sess("identified", USER),
      admin: false,
      admin_allowed: true,
      admin_mode: false,
    };
    expect(routeRedirect(admin, inPlayerMode)).toBe("#/lobby");
    // Before /me has answered, the person is not known to be on the
    // list, and sees the form until it does.
    expect(routeRedirect(admin, sess("identified", USER))).toBeNull();
  });

  it("shows the token form to any other session", () => {
    for (const [name, s] of SESSIONS) {
      if (s === null || s.principal.role === "admin" || name === "allowlisted admin person") {
        continue;
      }
      expect(routeRedirect(admin, s)).toBeNull();
    }
    // A stray admin_allowed on a session with no user is not a person.
    expect(routeRedirect(admin, { ...sess("player", NIL), admin_allowed: true })).toBeNull();
  });
});

describe("oauthCompleteTarget (§1 item 2)", () => {
  it("takes the invite flow to its table", () => {
    const s = sess("player", USER, { gameID: "g2" });
    expect(oauthCompleteTarget(s)).toBe("#/games/g2");
  });

  it("takes the login-page flow to the Lobby, not back to Login", () => {
    expect(oauthCompleteTarget(sess("identified", USER))).toBe("#/lobby");
    expect(oauthCompleteTarget(sess("identified"))).toBe("#/lobby");
  });
});

describe("joinBoxFor (§1 item 3)", () => {
  it("gives a signed-in person a code or a link", () => {
    expect(joinBoxFor(sess("identified", USER))).toBe("code");
    expect(joinBoxFor(sess("identified"))).toBe("code");
    expect(joinBoxFor(sess("player", USER))).toBe("code");
    expect(joinBoxFor(sess("spectator", USER))).toBe("code");
    expect(joinBoxFor(sess("player", USER, { admin: true }))).toBe("code");
  });

  it("gives a guest seat or guest spectator a link only", () => {
    expect(joinBoxFor(sess("player", NIL))).toBe("link");
    expect(joinBoxFor(sess("player"))).toBe("link");
    expect(joinBoxFor(sess("spectator"))).toBe("link");
  });

  it("gives the admin token no box, and nobody signed out one either", () => {
    expect(joinBoxFor(sess("admin"))).toBe("none");
    expect(joinBoxFor(null)).toBe("none");
  });

  it("tells a guest what to do with a bare code", () => {
    expect(GUEST_CODE_MESSAGE).toBe(
      "This browser is seated as a guest. Open the invite link instead, or link Discord from your table's menu.",
    );
  });
});

describe("inviteHash", () => {
  it("pulls the fragment out of a pasted link", () => {
    expect(inviteHash("https://cmd.labxp.io/#/games/g1/join?t=abc")).toBe("#/games/g1/join?t=abc");
    expect(inviteHash("#/games/g1/join?t=abc&spectator=1")).toBe(
      "#/games/g1/join?t=abc&spectator=1",
    );
  });

  it("returns empty for a bare code", () => {
    expect(inviteHash("abc123")).toBe("");
    expect(inviteHash("http://")).toBe("");
  });
});

function game(id: string, extra: Partial<GameMeta> = {}): GameMeta {
  return {
    id,
    name: id,
    created_at: new Date().toISOString(),
    players: [],
    state: "lobby",
    ...extra,
  };
}

describe("myTables: what the Lobby lists (§1 item 6, owner answer 3)", () => {
  const all = [
    game("g1"),
    game("mine-created", { is_creator: true }),
    game("seat-elsewhere"),
    game("other-pod"),
  ];
  const rejoinable = new Set(["seat-elsewhere"]);

  it("shows a signed-in person only their own tables", () => {
    const got = myTables(all, sess("player", USER), { admin: false, rejoinable });
    expect(got.map((g) => g.id)).toEqual(["g1", "mine-created", "seat-elsewhere"]);
  });

  it("shows an identified session its created and held tables, and nothing else", () => {
    const got = myTables(all, sess("identified", USER), { admin: false, rejoinable });
    expect(got.map((g) => g.id)).toEqual(["mine-created", "seat-elsewhere"]);
  });

  it("shows a signed-in spectator the table they watch", () => {
    const got = myTables(all, sess("spectator", USER), { admin: false, rejoinable: new Set() });
    expect(got.map((g) => g.id)).toEqual(["g1", "mine-created"]);
  });

  it("never falls back to every table for a signed-in person with none", () => {
    const others = [game("a"), game("b")];
    expect(myTables(others, sess("identified", USER), { admin: false, rejoinable })).toEqual([]);
  });

  it("shows an admin every table", () => {
    expect(myTables(all, sess("admin"), { admin: true, rejoinable })).toEqual(all);
    const person = sess("player", USER, { admin: true });
    expect(myTables(all, person, { admin: true, rejoinable })).toEqual(all);
  });

  it("keeps a guest's view as it was", () => {
    // A guest seat: its own table and the ones it created, else all.
    expect(
      myTables(all, sess("player", NIL), { admin: false, rejoinable }).map((g) => g.id),
    ).toEqual(["g1", "mine-created"]);
    const elsewhere = sess("player", NIL, { gameID: "gone" });
    const noneOfMine = [game("a"), game("b")];
    expect(myTables(noneOfMine, elsewhere, { admin: false, rejoinable })).toEqual(noneOfMine);
    // A guest spectator sees the room.
    expect(myTables(all, sess("spectator"), { admin: false, rejoinable })).toEqual(all);
  });
});
