// @vitest-environment jsdom
//
// #5438 (ADR 0222 §5): the pure rules of the draft page, the seat token's
// storage and the HTTP calls' shapes.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  buildMinuteChoices,
  canConfirm,
  cardAccessibleName,
  clockText,
  createDraft,
  defaultPickMode,
  DraftApiError,
  draftErrorMessage,
  draftHash,
  draftImageURL,
  draftNewHash,
  draftSocketURL,
  forgetDraftSeat,
  groupPool,
  isAway,
  isDraftGone,
  isHost,
  isLand,
  isUrgent,
  joinDraft,
  leaveDraft,
  msLeftAtFrame,
  openSeats,
  passTarget,
  pickCount,
  pickModeName,
  poolGroupOf,
  POOL_GROUP_ORDER,
  previewDraft,
  progressLine,
  pruneSelection,
  rememberDraftSeat,
  seatChoices,
  seatStatus,
  secondsLeft,
  stackCopies,
  startDraft,
  stillPicking,
  storedDraftIDs,
  storedDraftSeat,
  takeOf,
  timerAnnouncement,
  toggleSelection,
  viewAnnouncement,
  waitingLine,
  waitingOn,
  type DraftCard,
  type DraftView,
} from "./draft";
import { L } from "./labels";

function card(id: string, over: Partial<DraftCard> = {}): DraftCard {
  return { id, name: id, colors: [], mana_value: 0, type_line: "Artifact", ...over };
}

function view(over: Partial<DraftView> = {}): DraftView {
  return {
    id: "d1",
    name: "Friday draft",
    set: "blb",
    set_name: "Bloomburrow",
    booster: "play",
    mode: "pick_one",
    seats: 4,
    state: "picking",
    seq: 5,
    round: 2,
    rounds: 3,
    pick: 5,
    picks_in_round: 14,
    take: 1,
    passing: "right",
    pick_seconds: 60,
    deadline: "2026-11-11T19:00:45Z",
    now: "2026-11-11T19:00:02Z",
    you: 0,
    players: [
      { seat: 0, name: "Ian", host: true, connected: true, picking: true, picked: 4 },
      { seat: 1, name: "Bot 1", bot: true, connected: true, picking: false, picked: 5 },
      { seat: 2, name: "Sam", connected: false, picking: true, picked: 4 },
      { seat: 3, name: "Bot 3", bot: true, connected: true, picking: true, picked: 4 },
    ],
    pack: [card("c1"), card("c2"), card("c3")],
    picks: [],
    ...over,
  };
}

describe("grouping picks", () => {
  const pool: DraftCard[] = [
    card("forest", { name: "Forest", type_line: "Basic Land — Forest", basic: true }),
    card("g3", { name: "Bear", colors: ["G"], mana_value: 3, type_line: "Creature — Bear" }),
    card("g1", { name: "Elf", colors: ["G"], mana_value: 1, type_line: "Creature — Elf" }),
    card("g3b", { name: "Ape", colors: ["G"], mana_value: 3, type_line: "Creature — Ape" }),
    card("w2", { name: "Knight", colors: ["W"], mana_value: 2, type_line: "Creature — Knight" }),
    card("ub", { name: "Spy", colors: ["U", "B"], mana_value: 2, type_line: "Creature — Rogue" }),
    card("art", { name: "Relic", colors: [], mana_value: 4, type_line: "Artifact" }),
    card("arbor", {
      name: "Dryad Arbor",
      colors: ["G"],
      type_line: "Land Creature — Forest Dryad",
    }),
    card("r1", { name: "Shock", colors: ["R"], mana_value: 1, type_line: "Instant" }),
    card("u2", { name: "Looter", colors: ["U"], mana_value: 2, type_line: "Creature — Rat" }),
    card("b5", { name: "Reaper", colors: ["B"], mana_value: 5, type_line: "Creature — Horror" }),
  ];

  it("orders the groups W, U, B, R, G, multicolour, colourless, lands", () => {
    const groups = groupPool(pool);
    expect(groups.map((g) => g.key)).toEqual([...POOL_GROUP_ORDER]);
    expect(groups.map((g) => g.name)).toEqual([
      L.draftGroupWhite,
      L.draftGroupBlue,
      L.draftGroupBlack,
      L.draftGroupRed,
      L.draftGroupGreen,
      L.draftGroupMulticolour,
      L.draftGroupColourless,
      L.draftGroupLands,
    ]);
    expect(groups.reduce((n, g) => n + g.count, 0)).toBe(pool.length);
  });

  it("sorts a group by mana value, then by name, and rows it by mana value", () => {
    const green = groupPool(pool).find((g) => g.key === "G")!;
    expect(green.cards.map((c) => c.name)).toEqual(["Elf", "Ape", "Bear"]);
    expect(green.byValue.map((r) => [r.manaValue, r.cards.map((c) => c.name)])).toEqual([
      [1, ["Elf"]],
      [3, ["Ape", "Bear"]],
    ]);
    expect(green.count).toBe(3);
  });

  it("keeps lands apart, whatever their colours", () => {
    const lands = groupPool(pool).find((g) => g.key === "land")!;
    expect(lands.cards.map((c) => c.name)).toEqual(["Dryad Arbor", "Forest"]);
    expect(isLand({ type_line: "Land" })).toBe(true);
    expect(isLand({ basic: true })).toBe(true);
    expect(isLand({ type_line: "Creature — Elf" })).toBe(false);
    // A spell with a land on its back face is drafted as the spell.
    expect(isLand({ type_line: "Sorcery // Land" })).toBe(false);
    expect(isLand({ type_line: "Landfall Enthusiast" })).toBe(false);
  });

  it("puts two or more colours in multicolour and none in colourless", () => {
    expect(poolGroupOf(card("x", { colors: ["U", "B"] }))).toBe("multi");
    expect(poolGroupOf(card("x", { colors: ["U", "U"] }))).toBe("U");
    expect(poolGroupOf(card("x", { colors: [] }))).toBe("colorless");
    expect(poolGroupOf(card("x", { colors: undefined }))).toBe("colorless");
  });

  it("leaves out empty groups and takes an absent list", () => {
    expect(groupPool(undefined)).toEqual([]);
    expect(groupPool([pool[4]]).map((g) => g.key)).toEqual(["W"]);
  });

  it("does not reorder the list it was given", () => {
    const before = pool.map((c) => c.id);
    groupPool(pool);
    expect(pool.map((c) => c.id)).toEqual(before);
  });

  it("folds neighbouring copies into one row", () => {
    const rows = stackCopies([
      card("a", { name: "Forest" }),
      card("b", { name: "Forest" }),
      card("c", { name: "Island" }),
    ]);
    expect(rows.map((r) => [r.card.name, r.count])).toEqual([
      ["Forest", 2],
      ["Island", 1],
    ]);
  });
});

describe("the pick timer", () => {
  const frame = { deadline: "2026-11-11T19:00:45Z", now: "2026-11-11T19:00:02Z" };

  it("counts from the server's clock, not the viewer's", () => {
    expect(msLeftAtFrame(frame)).toBe(43_000);
    expect(secondsLeft(frame, 0)).toBe(43);
    expect(secondsLeft(frame, 999)).toBe(43);
    expect(secondsLeft(frame, 1000)).toBe(42);
    expect(secondsLeft(frame, 42_500)).toBe(1);
    // The viewer's wall clock is never read: only time since the frame.
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2031-01-01T00:00:00Z"));
    expect(secondsLeft(frame, 3000)).toBe(40);
    vi.useRealTimers();
  });

  it("stops at zero and never goes negative", () => {
    expect(secondsLeft(frame, 43_000)).toBe(0);
    expect(secondsLeft(frame, 90_000)).toBe(0);
    expect(secondsLeft({ deadline: frame.now, now: frame.deadline }, 0)).toBe(0);
    expect(secondsLeft(frame, -5000)).toBe(43);
  });

  it("is null with no timer running", () => {
    expect(secondsLeft({}, 0)).toBeNull();
    expect(secondsLeft({ deadline: frame.deadline }, 0)).toBeNull();
    expect(secondsLeft({ deadline: "soon", now: frame.now }, 0)).toBeNull();
  });

  it("turns urgent in the last 10 seconds", () => {
    expect(isUrgent(11)).toBe(false);
    expect(isUrgent(10)).toBe(true);
    expect(isUrgent(0)).toBe(true);
    expect(isUrgent(null)).toBe(false);
  });

  it("is announced at 30 and 10 seconds, not every second", () => {
    const said: string[] = [];
    let prev: number | null = null;
    for (let s = 60; s >= 0; s--) {
      const line = timerAnnouncement(prev, s);
      if (line) said.push(line);
      prev = s;
    }
    expect(said).toEqual(["30 seconds left to pick.", "10 seconds left to pick."]);
    // A new pick resets the clock upward: nothing to say.
    expect(timerAnnouncement(3, 60)).toBe("");
    expect(timerAnnouncement(null, 10)).toBe("");
    // A frame that lands late may skip the exact second.
    expect(timerAnnouncement(12, 9)).toBe("10 seconds left to pick.");
  });

  it("reads as a clock", () => {
    expect(clockText(65)).toBe("1:05");
    expect(clockText(9)).toBe("0:09");
    expect(clockText(0)).toBe("0:00");
  });
});

describe("the create form's defaults", () => {
  it("offers pick two first at four seats and pick one otherwise", () => {
    expect(defaultPickMode(4)).toBe("pick_two");
    for (const n of [2, 3, 5, 6, 7, 8]) expect(defaultPickMode(n)).toBe("pick_one");
  });

  it("offers 2 to 8 seats", () => {
    expect(seatChoices()).toEqual([2, 3, 4, 5, 6, 7, 8]);
  });

  it("names the modes from the registry", () => {
    expect(pickModeName("pick_one")).toBe(L.draftPickOne);
    expect(pickModeName("pick_two")).toBe(L.draftPickTwo);
    expect(pickCount(1)).toBe("Pick 1 card");
    expect(pickCount(2)).toBe("Pick 2 cards");
  });
});

describe("choosing cards", () => {
  it("with one to take, a click moves the choice and a second click lets go", () => {
    expect(toggleSelection([], "a", 1)).toEqual(["a"]);
    expect(toggleSelection(["a"], "b", 1)).toEqual(["b"]);
    expect(toggleSelection(["b"], "b", 1)).toEqual([]);
  });

  it("with two to take, a third choice replaces the oldest", () => {
    expect(toggleSelection([], "a", 2)).toEqual(["a"]);
    expect(toggleSelection(["a"], "b", 2)).toEqual(["a", "b"]);
    expect(toggleSelection(["a", "b"], "c", 2)).toEqual(["b", "c"]);
    expect(toggleSelection(["b", "c"], "b", 2)).toEqual(["c"]);
  });

  it("chooses nothing when there is nothing to take", () => {
    expect(toggleSelection(["a"], "b", 0)).toEqual([]);
  });

  it("confirms only exactly `take` different cards of the pack", () => {
    const pack = [card("a"), card("b"), card("c")];
    expect(canConfirm(["a"], pack, 1)).toBe(true);
    expect(canConfirm([], pack, 1)).toBe(false);
    expect(canConfirm(["a"], pack, 2)).toBe(false);
    expect(canConfirm(["a", "b"], pack, 2)).toBe(true);
    expect(canConfirm(["a", "a"], pack, 2)).toBe(false);
    expect(canConfirm(["a", "z"], pack, 2)).toBe(false);
    expect(canConfirm(["a"], undefined, 1)).toBe(false);
    expect(canConfirm([], pack, 0)).toBe(false);
  });

  it("drops choices a new pack does not hold", () => {
    expect(pruneSelection(["a", "b"], [card("b"), card("c")], 2)).toEqual(["b"]);
    expect(pruneSelection(["a", "b"], [card("a"), card("b")], 1)).toEqual(["b"]);
    expect(pruneSelection(["a"], undefined, 1)).toEqual([]);
  });

  it("takes what the server says, never more than the pack holds", () => {
    expect(takeOf(view({ take: 2, mode: "pick_two" }))).toBe(2);
    expect(takeOf(view({ take: 2, pack: [card("last")] }))).toBe(1);
    expect(takeOf(view({ pack: undefined, take: 0 }))).toBe(0);
    expect(takeOf(view({ state: "lobby" }))).toBe(0);
    expect(takeOf(view({ state: "done" }))).toBe(0);
  });
});

describe("the seats", () => {
  it("lists who is still picking, in seat order", () => {
    expect(stillPicking(view()).map((p) => p.name)).toEqual(["Ian", "Sam", "Bot 3"]);
    expect(waitingOn(view()).map((p) => p.name)).toEqual(["Sam", "Bot 3"]);
    expect(stillPicking(view({ state: "done" }))).toEqual([]);
    expect(stillPicking(view({ state: "lobby" }))).toEqual([]);
  });

  it("names up to three seats it waits on, then counts them", () => {
    expect(waitingLine(view())).toBe("Waiting for Sam and Bot 3.");
    const one = view();
    one.players[3].picking = false;
    expect(waitingLine(one)).toBe("Waiting for Sam.");
    const none = view();
    for (const p of none.players) p.picking = false;
    expect(waitingLine(none)).toContain("Everyone has picked");
    const many = view({
      seats: 8,
      players: Array.from({ length: 8 }, (_, i) => ({ seat: i, name: `P${i}`, picking: i > 0 })),
    });
    expect(waitingLine(many)).toBe("Waiting for 7 players.");
  });

  it("calls a person with no socket away, and never a bot", () => {
    const v = view();
    expect(isAway(v.players[2])).toBe(true);
    expect(isAway(v.players[0])).toBe(false);
    expect(isAway({ seat: 1, name: "Bot", bot: true, connected: false })).toBe(false);
    expect(seatStatus(v, v.players[2])).toBe("away");
    expect(seatStatus(v, v.players[0])).toBe("picking");
    expect(seatStatus(v, v.players[1])).toBe("picked");
    expect(seatStatus(view({ state: "lobby" }), v.players[0])).toBe("seated");
  });

  it("knows the host and the open seats", () => {
    expect(isHost(view())).toBe(true);
    expect(isHost(view({ you: 2 }))).toBe(false);
    const lobby = view({ state: "lobby", seats: 8, players: view().players.slice(0, 3) });
    expect(openSeats(lobby)).toBe(5);
    expect(openSeats(view())).toBe(0);
  });

  it("passes left to the next higher seat and right to the next lower, wrapping", () => {
    expect(passTarget(view({ passing: "left", you: 0 }))?.name).toBe("Bot 1");
    expect(passTarget(view({ passing: "left", you: 3 }))?.name).toBe("Ian");
    expect(passTarget(view({ passing: "right", you: 0 }))?.name).toBe("Bot 3");
    expect(passTarget(view({ passing: "right", you: 2 }))?.name).toBe("Bot 1");
    expect(passTarget(view({ state: "done" }))).toBeUndefined();
  });

  it("says where the pick is", () => {
    expect(progressLine(view())).toBe("Pack 2 of 3 · pick 5 of 14 · passing right");
    expect(progressLine(view({ state: "lobby", round: 0 }))).toBe("");
  });
});

describe("what a screen reader hears", () => {
  it("names a card by its name first", () => {
    expect(
      cardAccessibleName(
        card("c41", {
          name: "Shoreline Looter",
          mana_cost: "{1}{U}",
          type_line: "Creature — Rat Rogue",
          rarity: "uncommon",
        }),
      ),
    ).toMatch(/^Shoreline Looter, costs .*blue.*, Creature — Rat Rogue, uncommon$/);
    expect(cardAccessibleName(card("f", { name: "Forest", type_line: undefined }))).toBe("Forest");
  });

  it("announces the start, a new pack, a landed pick and the end, once each", () => {
    const lobby = view({ state: "lobby", round: 0, pack: undefined });
    const first = view({ round: 1, pick: 1, passing: "left" });
    expect(viewAnnouncement(lobby, first)).toBe(
      "The draft has started. Pack 1 of 3 · pick 1 of 14 · passing left. Pick 1 card.",
    );
    // The same pack again (another seat picked): nothing.
    expect(viewAnnouncement(first, { ...first, seq: 9 })).toBe("");
    const waiting = { ...first, pack: undefined, take: 0 };
    waiting.players = first.players.map((p) => ({ ...p, picking: p.seat === 2 }));
    expect(viewAnnouncement(first, waiting)).toBe("Pick made. Waiting for Sam.");
    expect(viewAnnouncement(waiting, { ...waiting, seq: 10 })).toBe("");
    const next = view({ round: 1, pick: 2, take: 2 });
    expect(viewAnnouncement(waiting, next)).toContain("pick 2 of 14");
    expect(viewAnnouncement(waiting, next)).toContain("Pick 2 cards.");
    const done = view({ state: "done", pack: undefined });
    expect(viewAnnouncement(next, done)).toBe(`${L.draftFinished}.`);
    expect(viewAnnouncement(done, done)).toBe("");
    // A reload into the middle of a pick says where it is.
    expect(viewAnnouncement(null, first)).toContain("Pack 1 of 3");
  });
});

describe("links and urls", () => {
  it("builds the draft's hashes", () => {
    expect(draftHash("abc")).toBe("#/draft/abc");
    expect(draftHash("abc", "in vite")).toBe("#/draft/abc?invite=in%20vite");
    expect(draftNewHash()).toBe("#/draft");
    expect(draftNewHash("blb")).toBe("#/draft?set=blb");
  });

  it("dials /draft/{id}/ws with the seat token", () => {
    const https = { protocol: "https:", host: "c2.example" } as Location;
    expect(draftSocketURL("d 1", "t&k", https)).toBe("wss://c2.example/draft/d%201/ws?token=t%26k");
    const http = { protocol: "http:", host: "localhost:5173" } as Location;
    expect(draftSocketURL("d", "t", http)).toBe("ws://localhost:5173/draft/d/ws?token=t");
  });

  it("draws a card from the public draft image route, never a session-gated one", () => {
    expect(draftImageURL(card("c", { scryfall_id: "abc" }))).toBe("/draft/image/abc?size=normal");
    expect(draftImageURL({ scryfall_id: "abc" }, "small", 1)).toBe(
      "/draft/image/abc?size=small&face=1",
    );
    expect(draftImageURL(card("c"))).toBeNull();
  });
});

describe("the seat token", () => {
  beforeEach(() => localStorage.clear());

  it("is kept under the draft's id and read back", () => {
    rememberDraftSeat("d1", { token: "tok", expires_at: "2099-01-01T00:00:00Z" });
    expect(storedDraftSeat("d1")).toEqual({ token: "tok", expiresAt: "2099-01-01T00:00:00Z" });
    expect(storedDraftSeat("d2")).toBeNull();
    expect(storedDraftIDs()).toEqual(["d1"]);
  });

  it("forgets an expired seat", () => {
    rememberDraftSeat("d1", { token: "tok", expires_at: "2026-01-01T00:00:00Z" });
    const now = Date.parse("2026-06-01T00:00:00Z");
    expect(storedDraftSeat("d1", now)).toBeNull();
    expect(storedDraftIDs(now)).toEqual([]);
    expect(localStorage.getItem("cmdctrl.draft.d1")).toBeNull();
  });

  it("forgets on request, and ignores junk", () => {
    rememberDraftSeat("d1", { token: "tok", expires_at: "2099-01-01T00:00:00Z" });
    forgetDraftSeat("d1");
    expect(storedDraftSeat("d1")).toBeNull();
    localStorage.setItem("cmdctrl.draft.bad", "{not json");
    expect(storedDraftSeat("bad")).toBeNull();
    localStorage.setItem("cmdctrl.draft.empty", JSON.stringify({ token: "" }));
    expect(storedDraftSeat("empty")).toBeNull();
  });

  it("survives a browser with no storage", () => {
    const get = vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("denied");
    });
    const setItem = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("denied");
    });
    expect(() => rememberDraftSeat("d1", { token: "t", expires_at: "" })).not.toThrow();
    expect(storedDraftSeat("d1")).toBeNull();
    expect(() => forgetDraftSeat("d1")).not.toThrow();
    get.mockRestore();
    setItem.mockRestore();
  });
});

describe("the draft routes", () => {
  const realFetch = globalThis.fetch;
  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  function serve(body: unknown, status = 200): ReturnType<typeof vi.fn> {
    const spy = vi.fn(
      async () =>
        new Response(status === 204 ? null : JSON.stringify(body), {
          status,
          headers: { "Content-Type": "application/json" },
        }),
    );
    globalThis.fetch = spy as never;
    return spy;
  }
  const call = (spy: ReturnType<typeof vi.fn>, i = 0) => {
    const [url, init] = spy.mock.calls[i] as [string, RequestInit];
    return {
      url,
      method: init.method,
      headers: init.headers as Record<string, string>,
      body: init.body ? (JSON.parse(init.body as string) as Record<string, unknown>) : undefined,
      credentials: init.credentials,
    };
  };

  it("creates with the chosen values, the cookie, and no bearer", async () => {
    const spy = serve({ token: "tok", expires_at: "x", draft: view({ state: "lobby" }) }, 201);
    const got = await createDraft({ set: "blb", mode: "pick_two", seats: 4, name: "  Friday  " });
    expect(got.token).toBe("tok");
    const c = call(spy);
    expect(c.url).toBe("/draft");
    expect(c.method).toBe("POST");
    expect(c.body).toEqual({ set: "blb", mode: "pick_two", seats: 4, name: "Friday" });
    expect(c.credentials).toBe("same-origin");
    expect(c.headers.Authorization).toBeUndefined();
  });

  it("leaves a blank name out, so the server generates one", async () => {
    const spy = serve({ token: "t", expires_at: "x", draft: view() }, 201);
    await createDraft({ set: "blb", mode: "pick_one", seats: 8, name: "   " });
    expect(call(spy).body).toEqual({ set: "blb", mode: "pick_one", seats: 8 });
  });

  it("sends the build timer choice: untimed, or a length, never both", async () => {
    const spy = serve({ token: "t", expires_at: "x", draft: view() }, 201);
    await createDraft({ set: "blb", mode: "pick_one", seats: 8, build_untimed: true });
    expect(call(spy).body).toEqual({ set: "blb", mode: "pick_one", seats: 8, build_untimed: true });
    await createDraft({
      set: "blb",
      mode: "pick_one",
      seats: 8,
      build_untimed: true,
      build_seconds: 600,
    });
    expect(call(spy, 1).body).toEqual({
      set: "blb",
      mode: "pick_one",
      seats: 8,
      build_untimed: true,
    });
    await createDraft({ cube: "c1", mode: "pick_one", seats: 4, build_seconds: 600 });
    expect(call(spy, 2).body).toEqual({
      cube: "c1",
      mode: "pick_one",
      seats: 4,
      build_seconds: 600,
    });
    expect(buildMinuteChoices()[0]).toBe(2);
    expect(buildMinuteChoices().at(-1)).toBe(30);
  });

  it("previews by invite, and by seat token", async () => {
    const spy = serve({ id: "d1", name: "n", set: "blb", mode: "pick_one", seats: 8, taken: 1 });
    await previewDraft("d1", { invite: "a b" });
    expect(call(spy).url).toBe("/draft/d1?invite=a%20b");
    await previewDraft("d1", { token: "seat" });
    expect(call(spy, 1).url).toBe("/draft/d1");
    expect(call(spy, 1).headers.Authorization).toBe("Bearer seat");
  });

  it("joins with the invite and a guest's name", async () => {
    const spy = serve({ token: "t", expires_at: "x", draft: view() });
    await joinDraft("d1", "inv", " Sam ");
    expect(call(spy).url).toBe("/draft/d1/join");
    expect(call(spy).body).toEqual({ invite: "inv", name: "Sam" });
    await joinDraft("d1", "inv");
    expect(call(spy, 1).body).toEqual({ invite: "inv" });
  });

  it("starts and leaves with the seat token as a bearer", async () => {
    let spy = serve({ draft: view({ seq: 6 }) });
    expect((await startDraft("d1", "seat")).seq).toBe(6);
    expect(call(spy).url).toBe("/draft/d1/start");
    expect(call(spy).headers.Authorization).toBe("Bearer seat");
    spy = serve(null, 204);
    await leaveDraft("d1", "seat");
    expect(call(spy).url).toBe("/draft/d1/leave");
    expect(call(spy).method).toBe("POST");
    expect(call(spy).headers.Authorization).toBe("Bearer seat");
  });

  it("carries the server's sentence, status and code on a refusal", async () => {
    serve({ error: "The server is full right now.", code: "server_full" }, 503);
    const err = await createDraft({ set: "blb", mode: "pick_one", seats: 8 }).catch((e) => e);
    expect(err).toBeInstanceOf(DraftApiError);
    expect(err.status).toBe(503);
    expect(err.code).toBe("server_full");
    expect(draftErrorMessage(err, "fallback")).toBe("The server is full right now.");
    expect(isDraftGone(err)).toBe(false);
    expect(draftErrorMessage(new TypeError("network"), "fallback")).toBe("fallback");
  });

  it("reads 404, 401 and 403 as a draft or a seat that is gone", async () => {
    for (const status of [404, 401, 403]) {
      serve({ error: "nope" }, status);
      const err = await previewDraft("d1", { token: "t" }).catch((e) => e);
      expect(isDraftGone(err), String(status)).toBe(true);
    }
    serve({ error: "started", code: "draft_started" }, 409);
    expect(isDraftGone(await joinDraft("d1", "i").catch((e) => e))).toBe(false);
  });
});
