// adminViews.test.ts — the pure half of ADR 0124 §7, the admin views:
// the routes and their filters (a filter built into a hash parses back
// to itself, and a value the server does not know is dropped, never
// sent), the seat label, the outcome line, relative times, the accounts
// list's sort and search, and Live now's poll schedule.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  accountsHash,
  accountsQuery,
  adminViewHash,
  archiveConfirm,
  avatarSrc,
  canArchive,
  canReplay,
  canUnarchive,
  externalLink,
  filterAccounts,
  gameHash,
  gamesHash,
  gamesQuery,
  LIVE_IDLE_MS,
  LIVE_POLL_MS,
  LivePoller,
  outcomeLine,
  relativeTime,
  revokeConfirm,
  seatKind,
  seatLabel,
  sortAccounts,
  tabOf,
  type AccountsFilter,
  type AdminAccount,
  type AdminSeat,
  type GamesFilter,
} from "./adminViews";
import { parseHash } from "./router";

const GAME = "9c2f0e4c-1d1e-4c3b-9d7e-2f8a1b2c3d4e";
const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";

describe("parseHash: the admin views (§7)", () => {
  it("keeps #/admin alone as the token's form", () => {
    expect(parseHash("#/admin")).toEqual({ name: "adminLogin" });
    expect(parseHash("#/admin/")).toEqual({ name: "adminLogin" });
    expect(parseHash("#/admin?x=1")).toEqual({ name: "adminLogin" });
  });

  it("keeps a path under #/admin that names no view as the token's form", () => {
    for (const hash of ["#/admin/nowhere", "#/admin/live/extra", "#/admin/games/g1/more"]) {
      expect(parseHash(hash)).toEqual({ name: "adminLogin" });
    }
  });

  it("parses Live now", () => {
    expect(parseHash("#/admin/live")).toEqual({ name: "adminViews", view: { view: "live" } });
  });

  it("parses Games and its filters", () => {
    expect(parseHash("#/admin/games")).toEqual({
      name: "adminViews",
      view: { view: "games", filter: {} },
    });
    expect(parseHash("#/admin/games?state=active&archived=false")).toEqual({
      name: "adminViews",
      view: { view: "games", filter: { state: "active", archived: "false" } },
    });
    expect(parseHash("#/admin/games?practice=only")).toEqual({
      name: "adminViews",
      view: { view: "games", filter: { practice: "only" } },
    });
    expect(parseHash(`#/admin/games?user=${USER}&cursor=abc_123-=`)).toEqual({
      name: "adminViews",
      view: { view: "games", filter: { user: USER, cursor: "abc_123-=" } },
    });
  });

  it("parses one table and one account", () => {
    expect(parseHash(`#/admin/games/${GAME}`)).toEqual({
      name: "adminViews",
      view: { view: "game", id: GAME },
    });
    expect(parseHash(`#/admin/accounts/${USER}`)).toEqual({
      name: "adminViews",
      view: { view: "account", id: USER },
    });
  });

  it("parses Accounts with played and sort", () => {
    expect(parseHash("#/admin/accounts?played=7d&sort=first_seen")).toEqual({
      name: "adminViews",
      view: { view: "accounts", filter: { played: "7d", sort: "first_seen" } },
    });
  });

  it("parses every Grafana link of §9", () => {
    const links = [
      "#/admin/games?state=active&archived=false",
      "#/admin/games?state=lobby&archived=false",
      "#/admin/live",
      "#/admin/games?practice=only",
      "#/admin/accounts",
      "#/admin/accounts?played=1d",
      "#/admin/accounts?played=30d",
      "#/admin/games",
      "#/admin/accounts?sort=first_seen",
      "#/admin/games?state=ended&archived=false",
    ];
    for (const link of links) {
      const r = parseHash(link);
      expect(r.name).toBe("adminViews");
      if (r.name === "adminViews") expect(adminViewHash(r.view)).toBe(link);
    }
  });

  it("drops a value the server does not know, and keeps the rest", () => {
    expect(
      parseHash("#/admin/games?state=finished&archived=yes&practice=all&user=bob&cursor=a%20b"),
    ).toEqual({ name: "adminViews", view: { view: "games", filter: {} } });
    expect(parseHash("#/admin/games?state=ended&archived=maybe")).toEqual({
      name: "adminViews",
      view: { view: "games", filter: { state: "ended" } },
    });
    expect(parseHash("#/admin/accounts?played=2d&sort=age")).toEqual({
      name: "adminViews",
      view: { view: "accounts", filter: {} },
    });
  });
});

describe("filters round trip (§8)", () => {
  const GAMES: GamesFilter[] = [
    {},
    { state: "lobby" },
    { state: "active", archived: "false" },
    { state: "ended", archived: "true", practice: "include" },
    { practice: "only" },
    { practice: "exclude", user: USER },
    { state: "active", cursor: "eyJ0IjoxfQ==" },
  ];
  const ACCOUNTS: AccountsFilter[] = [
    {},
    { played: "1d" },
    { played: "7d", sort: "name" },
    { played: "30d", sort: "games" },
    { sort: "last_played" },
  ];

  for (const f of GAMES) {
    it(`games ${JSON.stringify(f)} parses back to itself`, () => {
      const r = parseHash(gamesHash(f));
      expect(r).toEqual({ name: "adminViews", view: { view: "games", filter: f } });
    });
  }

  for (const f of ACCOUNTS) {
    it(`accounts ${JSON.stringify(f)} parses back to itself`, () => {
      const r = parseHash(accountsHash(f));
      expect(r).toEqual({ name: "adminViews", view: { view: "accounts", filter: f } });
    });
  }

  it("never puts an unknown value in a hash or a query", () => {
    const bad = { state: "finished", archived: "1", practice: "all", user: "x", cursor: "a b" };
    expect(gamesHash(bad as unknown as GamesFilter)).toBe("#/admin/games");
    expect(gamesQuery(bad as unknown as GamesFilter)).toBe("/admin/games");
    const badAccounts = { played: "2d", sort: "age" };
    expect(accountsHash(badAccounts as unknown as AccountsFilter)).toBe("#/admin/accounts");
    expect(accountsQuery(badAccounts as unknown as AccountsFilter)).toBe("/admin/users");
  });

  it("builds the server's queries, with sort never sent", () => {
    expect(gamesQuery({})).toBe("/admin/games");
    expect(gamesQuery({ state: "active", archived: "false" })).toBe(
      "/admin/games?state=active&archived=false",
    );
    expect(gamesQuery({ practice: "only", user: USER, cursor: "c1" })).toBe(
      `/admin/games?practice=only&user=${USER}&cursor=c1`,
    );
    expect(accountsQuery({})).toBe("/admin/users");
    expect(accountsQuery({ played: "7d", sort: "name" })).toBe("/admin/users?played=7d");
    expect(accountsQuery({ sort: "games" })).toBe("/admin/users");
  });

  it("builds the detail hashes and maps each view to its tab", () => {
    expect(gameHash(GAME)).toBe(`#/admin/games/${GAME}`);
    expect(tabOf({ view: "game", id: GAME })).toBe("games");
    expect(tabOf({ view: "account", id: USER })).toBe("accounts");
    expect(tabOf({ view: "live" })).toBe("live");
  });
});

describe("the seat label (§8)", () => {
  const base: AdminSeat = { seat: 0, kind: "human", host: false };

  it("names an account, a guest, a pending Discord seat, a bot and an agent", () => {
    const account = { ...base, account: { id: USER, name: "Ann" } };
    expect(seatLabel(account)).toBe("Ann (account)");
    expect(seatLabel({ ...base, guest_name: "Gus" })).toBe("Gus (guest)");
    expect(seatLabel({ ...base, guest_name: "Dee", discord_pending: true })).toBe(
      "Dee (Discord, not signed in yet)",
    );
    expect(seatLabel({ ...base, kind: "bot", guest_name: "Bot 1", bot_tier: "heuristic" })).toBe(
      "Bot 1 (bot · heuristic)",
    );
    expect(
      seatLabel({ ...base, kind: "agent", guest_name: "Claude", agent_client: "claude-code" }),
    ).toBe("Claude (agent · claude-code)");
  });

  it("falls back to the seat number, and to 'unknown' for an agent with no client", () => {
    expect(seatLabel({ ...base, seat: 2 })).toBe("seat 3 (guest)");
    expect(seatKind({ ...base, kind: "agent" })).toBe("agent · unknown");
    expect(seatKind({ ...base, kind: "bot" })).toBe("bot");
  });
});

describe("the outcome line (§8)", () => {
  const seats: AdminSeat[] = [
    { seat: 0, kind: "human", host: true, account: { id: USER, name: "Ann" } },
    { seat: 1, kind: "bot", host: false, guest_name: "Bot 1", bot_tier: "random" },
  ];

  it("names the winner, a draw and a close with no result", () => {
    expect(outcomeLine({ state: "ended", outcome: "win", winner_seat: 1, seats })).toBe(
      "Won by Bot 1",
    );
    expect(outcomeLine({ state: "ended", outcome: "win", winner_seat: 3, seats })).toBe(
      "Won by seat 4",
    );
    expect(outcomeLine({ state: "ended", outcome: "win", seats })).toBe("Won");
    expect(outcomeLine({ state: "ended", outcome: "draw", seats })).toBe("Draw");
    expect(outcomeLine({ state: "ended", outcome: "closed", seats })).toBe("Closed, no result");
  });

  it("says where an unfinished table stands", () => {
    expect(outcomeLine({ state: "lobby", seats })).toBe("Waiting to start");
    expect(outcomeLine({ state: "active", seats })).toBe("In progress");
    expect(outcomeLine({ state: "ended", seats })).toBe("Ended");
  });
});

describe("relative times (§7)", () => {
  const now = Date.UTC(2026, 9, 5, 12, 0, 0);

  it("reads as 'just now', minutes, hours, days, then the date", () => {
    expect(relativeTime(now - 10_000, now)).toBe("just now");
    expect(relativeTime(now - 4 * 60_000, now)).toBe("4 min ago");
    expect(relativeTime(now - 59 * 60_000, now)).toBe("59 min ago");
    expect(relativeTime(now - 3 * 3_600_000, now)).toBe("3 h ago");
    expect(relativeTime(now - 2 * 86_400_000, now)).toBe("2 d ago");
    expect(relativeTime(now - 90 * 86_400_000, now)).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });

  it("reads a future time (clock skew) as 'just now', and absent as a dash", () => {
    expect(relativeTime(now + 60_000, now)).toBe("just now");
    expect(relativeTime(undefined, now)).toBe("—");
    expect(relativeTime(0, now)).toBe("—");
  });
});

describe("the accounts list: sort and search", () => {
  const A: AdminAccount[] = [
    {
      id: "1",
      name: "carol",
      games_played: 2,
      playing_now: false,
      first_seen_at: 300,
      last_played_at: 50,
    },
    { id: "2", name: "Ann", games_played: 9, playing_now: true, first_seen_at: 100 },
    {
      id: "3",
      name: "bob",
      games_played: 9,
      playing_now: false,
      first_seen_at: 200,
      last_played_at: 80,
    },
  ];
  const names = (l: AdminAccount[]): string[] => l.map((a) => a.name);

  it("keeps the server's order with no sort", () => {
    expect(names(sortAccounts(A, undefined))).toEqual(["carol", "Ann", "bob"]);
  });

  it("sorts by each column, ties by name", () => {
    expect(names(sortAccounts(A, "name"))).toEqual(["Ann", "bob", "carol"]);
    expect(names(sortAccounts(A, "first_seen"))).toEqual(["carol", "bob", "Ann"]);
    expect(names(sortAccounts(A, "games"))).toEqual(["Ann", "bob", "carol"]);
    expect(names(sortAccounts(A, "last_played"))).toEqual(["Ann", "bob", "carol"]);
    // The list given is never reordered in place.
    expect(names(A)).toEqual(["carol", "Ann", "bob"]);
  });

  it("filters by name as you type, any case", () => {
    expect(names(filterAccounts(A, "AN"))).toEqual(["Ann"]);
    expect(names(filterAccounts(A, "  "))).toEqual(names(A));
    expect(filterAccounts(A, "zed")).toEqual([]);
  });
});

describe("links, avatars and the linked actions", () => {
  it("uses only a same-origin avatar path, with the token", () => {
    expect(avatarSrc("/avatars/123/abc.png", "t k")).toBe("/avatars/123/abc.png?token=t%20k");
    expect(avatarSrc("/avatars/123/abc.png", undefined)).toBe("/avatars/123/abc.png");
    expect(avatarSrc("https://evil.example/x.png", "t")).toBeNull();
    expect(avatarSrc(undefined, "t")).toBeNull();
  });

  it("uses only http(s) external links", () => {
    expect(externalLink("https://moxfield.com/decks/a")).toBe("https://moxfield.com/decks/a");
    expect(externalLink("javascript:alert(1)")).toBeNull();
    expect(externalLink(undefined)).toBeNull();
  });

  it("offers archive, unarchive and replay on the right tables", () => {
    expect(canArchive({ practice: false })).toBe(true);
    expect(canArchive({ practice: false, archived_at: 1 })).toBe(false);
    expect(canArchive({ practice: true })).toBe(false);
    expect(canUnarchive({ archived_at: 1 })).toBe(true);
    expect(canUnarchive({})).toBe(false);
    expect(canReplay({ state: "ended", practice: false })).toBe(true);
    expect(canReplay({ state: "active", practice: false })).toBe(false);
  });

  it("says that archiving a running table ends it, and names the person to revoke", () => {
    expect(archiveConfirm({ name: "Friday", state: "active" })).toContain("ends the game");
    expect(archiveConfirm({ name: "Friday", state: "lobby" })).not.toContain("ends the game");
    const r = revokeConfirm("Ann");
    expect(r).toContain("Ann");
    expect(r).toContain("until they sign in with Discord again");
  });
});

describe("Live now's poll schedule (§7)", () => {
  let visible: boolean;
  let loads: number;
  let states: string[];

  function poller(): LivePoller {
    return new LivePoller({
      load: () => loads++,
      now: () => Date.now(),
      isVisible: () => visible,
      onState: (s) => states.push(s),
    });
  }

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(Date.UTC(2026, 9, 5, 12, 0, 0));
    visible = true;
    loads = 0;
    states = [];
  });
  afterEach(() => vi.useRealTimers());

  it("loads at once, then every 10 seconds while visible", () => {
    const p = poller();
    p.start();
    expect(loads).toBe(1);
    vi.advanceTimersByTime(LIVE_POLL_MS - 1);
    expect(loads).toBe(1);
    vi.advanceTimersByTime(1);
    expect(loads).toBe(2);
    vi.advanceTimersByTime(3 * LIVE_POLL_MS);
    expect(loads).toBe(5);
    expect(p.state).toBe("polling");
    p.stop();
    vi.advanceTimersByTime(10 * LIVE_POLL_MS);
    expect(loads).toBe(5);
  });

  it("stops while hidden and loads once when visible again", () => {
    const p = poller();
    p.start();
    visible = false;
    p.visibilityChanged();
    expect(p.state).toBe("hidden");
    vi.advanceTimersByTime(30 * LIVE_POLL_MS);
    expect(loads).toBe(1);
    visible = true;
    p.visibilityChanged();
    expect(p.state).toBe("polling");
    expect(loads).toBe(2);
    vi.advanceTimersByTime(LIVE_POLL_MS);
    expect(loads).toBe(3);
    p.stop();
  });

  it("does not load at start in a hidden tab", () => {
    visible = false;
    const p = poller();
    p.start();
    expect(loads).toBe(0);
    expect(p.state).toBe("hidden");
    p.stop();
  });

  it("pauses after 10 minutes with no input, and Resume asks again", () => {
    const p = poller();
    p.start();
    vi.advanceTimersByTime(LIVE_IDLE_MS + LIVE_POLL_MS);
    expect(p.state).toBe("paused");
    const at = loads;
    // One load at start, then one per tick before the idle mark.
    expect(at).toBe(1 + LIVE_IDLE_MS / LIVE_POLL_MS - 1);
    vi.advanceTimersByTime(60 * LIVE_POLL_MS);
    expect(loads).toBe(at);
    // Input alone does not un-pause; the page's Resume does.
    p.input();
    vi.advanceTimersByTime(LIVE_POLL_MS);
    expect(loads).toBe(at);
    p.resume();
    expect(p.state).toBe("polling");
    expect(loads).toBe(at + 1);
    vi.advanceTimersByTime(LIVE_POLL_MS);
    expect(loads).toBe(at + 2);
    expect(states).toEqual(["paused", "polling"]);
    p.stop();
  });

  it("keeps polling while there is input", () => {
    const p = poller();
    p.start();
    for (let i = 0; i < 120; i++) {
      vi.advanceTimersByTime(LIVE_POLL_MS);
      if (i % 30 === 0) p.input();
    }
    expect(p.state).toBe("polling");
    expect(loads).toBe(121);
    p.stop();
  });

  it("stays paused when the tab is hidden and shown again", () => {
    const p = poller();
    p.start();
    vi.advanceTimersByTime(LIVE_IDLE_MS + LIVE_POLL_MS);
    expect(p.state).toBe("paused");
    const at = loads;
    visible = false;
    p.visibilityChanged();
    visible = true;
    p.visibilityChanged();
    expect(p.state).toBe("paused");
    expect(loads).toBe(at);
    p.stop();
  });
});
