// @vitest-environment jsdom
//
// #5438 (ADR 0222 §5): the booster draft's pages. The create form, the
// invite page, the draft's lobby, the pick and the finished pool, drawn
// from the wire as docs/lobby.md and docs/protocol.md give it.

import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import Draft from "./Draft.svelte";
import { rememberDraftSeat, storedDraftSeat, type DraftCard, type DraftView } from "../lib/draft";
import { PACKS_ARE_APPROXIMATIONS } from "../lib/draftSets";
import { L } from "../lib/labels";
import { cleanup, flushSync, render } from "../lib/test/render.svelte";

const ID = "6f1c2a9e-0000-4000-8000-000000000000";
const realFetch = globalThis.fetch;

// ---- A fetch that answers by path, and remembers what it was asked ----

interface Call {
  url: string;
  method: string;
  body: Record<string, unknown> | undefined;
  headers: Record<string, string>;
  credentials: RequestCredentials | undefined;
}
type Answer = { status?: number; body?: unknown };
let calls: Call[] = [];

function serve(routes: Record<string, Answer>): void {
  calls = [];
  globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init: RequestInit = {}) => {
    const url = String(input);
    const method = init.method ?? "GET";
    calls.push({
      url,
      method,
      body: init.body ? (JSON.parse(String(init.body)) as Record<string, unknown>) : undefined,
      headers: (init.headers ?? {}) as Record<string, string>,
      credentials: init.credentials,
    });
    // The site header asks for things of its own; they get an empty answer.
    const a = routes[`${method} ${url}`] ?? routes[`${method} ${url.split("?")[0]}`] ?? {};
    const status = a.status ?? 200;
    return new Response(status === 204 ? null : JSON.stringify(a.body ?? {}), {
      status,
      headers: { "Content-Type": "application/json" },
    });
  }) as never;
}
const sent = (method: string, path: string): Call[] =>
  calls.filter((c) => c.method === method && c.url.split("?")[0] === path);

// ---- A WebSocket that the test drives ----

class FakeSocket {
  static opened: FakeSocket[] = [];
  static readonly OPEN = 1;
  readyState = 1;
  sent: string[] = [];
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};
  constructor(readonly url: string) {
    FakeSocket.opened.push(this);
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(type: string, fn: (ev: unknown) => void): void {
    this.listeners[type] = (this.listeners[type] ?? []).filter((f) => f !== fn);
  }
  send(data: string): void {
    this.sent.push(data);
  }
  close(): void {
    this.readyState = 3;
  }
  emit(type: string, ev: unknown): void {
    for (const fn of [...(this.listeners[type] ?? [])]) fn(ev);
  }
  frame(kind: string, payload: unknown, id = "srv"): void {
    this.emit("message", { data: JSON.stringify({ v: 0, kind, id, payload }) });
    flushSync();
  }
  drop(code: number, reason = ""): void {
    this.readyState = 3;
    this.emit("close", { code, reason });
    flushSync();
  }
  picks(): { id: string; payload: { cards: string[] } }[] {
    return this.sent.map((s) => JSON.parse(s)).filter((f) => f.kind === "draft_pick");
  }
}
const socket = (): FakeSocket => FakeSocket.opened[FakeSocket.opened.length - 1];

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

beforeEach(() => {
  localStorage.clear();
  FakeSocket.opened = [];
  vi.stubGlobal("WebSocket", FakeSocket);
  location.hash = "";
});

afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

const text = (el: Element | null | undefined): string =>
  (el?.textContent ?? "").replace(/\s+/g, " ").trim();
const main = (c: HTMLElement): HTMLElement => c.querySelector("main")!;
const button = (c: HTMLElement, name: string): HTMLButtonElement => {
  const b = [...c.querySelectorAll<HTMLButtonElement>("main button")].find(
    (el) => text(el) === name || el.getAttribute("aria-label") === name,
  );
  if (!b) throw new Error(`no button named "${name}"`);
  return b;
};
const cardButtons = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>(`[aria-label="${L.draftPack}"] button`),
];
const click = (el: Element): void => {
  el.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
  flushSync();
};

// ---- Fixtures ----

const SETS = {
  sets: [
    {
      code: "new",
      name: "Newest Set",
      released: "2026-09-01",
      booster: "play",
      cards: 270,
      draftable: false,
      blockers: ["Card One"],
    },
    {
      code: "blb",
      name: "Bloomburrow",
      released: "2024-08-02",
      booster: "play",
      cards: 261,
      draftable: true,
      blockers: [],
    },
    {
      code: "dmu",
      name: "Dominaria United",
      released: "2022-09-09",
      booster: "draft",
      cards: 250,
      draftable: true,
      blockers: [],
    },
  ],
  draftable: 2,
};

function card(id: string, name: string, over: Partial<DraftCard> = {}): DraftCard {
  return {
    id,
    name,
    scryfall_id: `sf-${id}`,
    rarity: "common",
    colors: ["U"],
    mana_cost: "{1}{U}",
    mana_value: 2,
    type_line: "Creature — Rat Rogue",
    ...over,
  };
}

const PACK = [
  card("c1", "Shoreline Looter"),
  card("c2", "Bonecache Overseer", { colors: ["B"], mana_cost: "{2}{B}", mana_value: 3 }),
  card("c3", "Fountainport Bell", {
    colors: [],
    mana_cost: "{1}",
    mana_value: 1,
    type_line: "Artifact",
  }),
];

function view(over: Partial<DraftView> = {}): DraftView {
  return {
    id: ID,
    name: "Friday draft",
    set: "blb",
    set_name: "Bloomburrow",
    booster: "play",
    mode: "pick_one",
    seats: 4,
    solo: false,
    state: "picking",
    seq: 12,
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
    pack: PACK,
    picks: [card("p1", "Thundertrap Trainer")],
    ...over,
  };
}

/** Mount the draft's own page for a browser that holds a seat, and feed it a view. */
async function room(v: DraftView | null = view()): Promise<HTMLElement> {
  serve({});
  rememberDraftSeat(ID, { token: "seat-token", expires_at: "2099-01-01T00:00:00Z" });
  const { container } = render(Draft as never, { id: ID } as never);
  await settle();
  if (v) socket().frame("draft", v);
  return container;
}

// ---- The create form ----

describe("the create form", () => {
  it("lists only the draftable sets, newest first, and says what a pack is", async () => {
    serve({ "GET /draft/sets": { body: SETS } });
    const { container } = render(Draft as never, {} as never);
    await settle();

    expect(text(container.querySelector("h1"))).toBe(L.draftNewTitle);
    const options = [...container.querySelectorAll<HTMLOptionElement>("#draft-set option")];
    expect(options.map((o) => o.value)).toEqual(["blb", "dmu"]);
    expect(text(options[0])).toContain("Bloomburrow");
    expect(text(options[0])).toContain(L.draftPlayBooster);
    expect(text(options[1])).toContain(L.draftDraftBooster);
    expect(text(main(container))).not.toContain("Newest Set");

    expect(text(container.querySelector('[data-testid="packs-approximate"]'))).toBe(
      PACKS_ARE_APPROXIMATIONS,
    );
    // A visitor with no sign-in is told what they get: bots in every other seat.
    const note = text(container.querySelector('[data-testid="bots-note"]'));
    expect(note).toContain("solo draft");
    expect(note).toContain("every other seat is a bot");
    // The seats run 2 to 8, and both modes are explained.
    expect(
      [...container.querySelectorAll<HTMLOptionElement>("#draft-seats option")].map((o) => o.value),
    ).toEqual(["2", "3", "4", "5", "6", "7", "8"]);
    const modes = [...container.querySelectorAll(".mode")].map(text);
    expect(modes[0]).toContain(L.draftPickOne);
    expect(modes[0]).toContain("one card from each pack");
    expect(modes[1]).toContain(L.draftPickTwo);
    expect(modes[1]).toContain("two cards from each pack");
    expect(text(main(container))).not.toMatch(/\bpod\b/i);
  });

  it("offers pick two first at four seats, and keeps a choice the person made", async () => {
    serve({ "GET /draft/sets": { body: SETS } });
    const { container } = render(Draft as never, {} as never);
    await settle();
    const radio = (v: string) =>
      container.querySelector<HTMLInputElement>(`input[name="draft-mode"][value="${v}"]`)!;
    const seats = container.querySelector<HTMLSelectElement>("#draft-seats")!;
    const setSeats = (n: number): void => {
      seats.value = String(n);
      seats.dispatchEvent(new Event("change", { bubbles: true }));
      flushSync();
    };
    expect(seats.value).toBe("8");
    expect(radio("pick_one").checked).toBe(true);
    setSeats(4);
    expect(radio("pick_two").checked).toBe(true);
    setSeats(6);
    expect(radio("pick_one").checked).toBe(true);
    // Either mode is allowed at any size: a choice sticks.
    radio("pick_two").checked = true;
    radio("pick_two").dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    setSeats(8);
    expect(radio("pick_two").checked).toBe(true);
  });

  it("posts the chosen values, keeps the seat token and opens the draft", async () => {
    serve({
      "GET /draft/sets": { body: SETS },
      "POST /draft": {
        status: 201,
        body: {
          token: "seat-token",
          expires_at: "2099-01-01T00:00:00Z",
          draft: view({ state: "lobby" }),
        },
      },
    });
    const { container } = render(Draft as never, { set: "dmu" } as never);
    await settle();

    // The sets page's link chose the set.
    const set = container.querySelector<HTMLSelectElement>("#draft-set")!;
    expect(set.value).toBe("dmu");
    const seats = container.querySelector<HTMLSelectElement>("#draft-seats")!;
    seats.value = "4";
    seats.dispatchEvent(new Event("change", { bubbles: true }));
    const name = container.querySelector<HTMLInputElement>("#draft-name")!;
    name.value = "Friday draft";
    name.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();

    expect(text(button(container, L.draftCreate))).toBe(L.draftCreate);
    container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();

    const post = sent("POST", "/draft");
    expect(post).toHaveLength(1);
    expect(post[0].body).toEqual({ set: "dmu", mode: "pick_two", seats: 4, name: "Friday draft" });
    // The cookie rides along, so a signed-in person is recognised; the
    // seat token is the answer's, never a bearer on the way in.
    expect(post[0].credentials).toBe("same-origin");
    expect(post[0].headers.Authorization).toBeUndefined();
    expect(storedDraftSeat(ID)?.token).toBe("seat-token");
    expect(location.hash).toBe(`#/draft/${ID}`);
  });

  it("says so when the link's set cannot be drafted, and picks one that can", async () => {
    serve({ "GET /draft/sets": { body: SETS } });
    const { container } = render(Draft as never, { set: "new" } as never);
    await settle();
    expect(text(main(container))).toContain("That set can’t be drafted yet");
    expect(container.querySelector<HTMLSelectElement>("#draft-set")!.value).toBe("blb");
  });

  it("shows the server's sentence when the server is full", async () => {
    serve({
      "GET /draft/sets": { body: SETS },
      "POST /draft": {
        status: 503,
        body: {
          error: "The server is full right now. Please try again in a few minutes.",
          code: "server_full",
        },
      },
    });
    const { container } = render(Draft as never, {} as never);
    await settle();
    container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(text(container.querySelector('[role="alert"]'))).toContain(
      "The server is full right now. Please try again in a few minutes.",
    );
    expect(location.hash).toBe("");
    expect(button(container, L.draftCreate).disabled).toBe(false);
  });

  it("points at the sets page when the server says the set is not draftable", async () => {
    serve({
      "GET /draft/sets": { body: SETS },
      "POST /draft": {
        status: 422,
        body: { error: "that set cannot be drafted yet", code: "set_not_draftable" },
      },
    });
    const { container } = render(Draft as never, {} as never);
    await settle();
    container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    const alert = container.querySelector('[role="alert"]')!;
    expect(text(alert)).toContain("that set cannot be drafted yet");
    expect(alert.querySelector("a")?.getAttribute("href")).toBe("#/draft/sets");
  });

  it("has no form when no set is ready", async () => {
    serve({ "GET /draft/sets": { body: { sets: [SETS.sets[0]], draftable: 0 } } });
    const { container } = render(Draft as never, {} as never);
    await settle();
    expect(container.querySelector("form")).toBeNull();
    expect(text(main(container))).toContain("No set can be drafted yet.");
  });

  it("offers the way back to a draft this browser holds a seat in", async () => {
    serve({ "GET /draft/sets": { body: SETS } });
    rememberDraftSeat(ID, { token: "t", expires_at: "2099-01-01T00:00:00Z" });
    const { container } = render(Draft as never, {} as never);
    await settle();
    const back = [...container.querySelectorAll<HTMLAnchorElement>("main a")].find((a) =>
      text(a).includes(L.draftReturn),
    );
    expect(back?.getAttribute("href")).toBe(`#/draft/${ID}`);
  });
});

// ---- The invite page ----

describe("the invite page", () => {
  const PREVIEW = {
    id: ID,
    name: "Friday draft",
    set: "blb",
    set_name: "Bloomburrow",
    mode: "pick_two",
    seats: 8,
    taken: 3,
    state: "lobby",
    host: "Ian",
  };

  it("shows the draft, asks a guest for a name, joins and connects to the seat", async () => {
    serve({
      [`GET /draft/${ID}`]: { body: PREVIEW },
      [`POST /draft/${ID}/join`]: {
        body: {
          token: "joined-token",
          expires_at: "2099-01-01T00:00:00Z",
          draft: view({ state: "lobby" }),
        },
      },
    });
    const { container } = render(Draft as never, { id: ID, invite: "AbC" } as never);
    await settle();

    expect(sent("GET", `/draft/${ID}`)[0].url).toBe(`/draft/${ID}?invite=AbC`);
    expect(text(container.querySelector("h1"))).toBe("Friday draft");
    const facts = text(container.querySelector('[data-testid="draft-preview"]'));
    expect(facts).toContain("Bloomburrow");
    expect(facts).toContain(L.draftPickTwo);
    expect(facts).toContain("3 of 8 taken");
    expect(facts).toContain("Ian");
    expect(FakeSocket.opened).toHaveLength(0);

    const name = container.querySelector<HTMLInputElement>("#draft-guest-name")!;
    expect(container.querySelector('label[for="draft-guest-name"]')?.textContent).toBe(
      L.draftGuestName,
    );
    name.value = "Sam";
    name.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(text(button(container, L.draftJoin))).toBe(L.draftJoin);
    container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();

    expect(sent("POST", `/draft/${ID}/join`)[0].body).toEqual({ invite: "AbC", name: "Sam" });
    expect(storedDraftSeat(ID)?.token).toBe("joined-token");
    expect(socket().url).toContain(`/draft/${ID}/ws?token=joined-token`);
  });

  it("says plainly when the invite leads nowhere", async () => {
    serve({ [`GET /draft/${ID}`]: { status: 404, body: { error: "not found" } } });
    const { container } = render(Draft as never, { id: ID, invite: "old" } as never);
    await settle();
    expect(text(container.querySelector("h1"))).toBe(L.draftGone);
    expect(container.querySelector("form")).toBeNull();
    expect(text(main(container))).toContain(L.draftNewDraft);
  });

  it("shows the server's reason when the draft is full", async () => {
    serve({
      [`GET /draft/${ID}`]: { body: { ...PREVIEW, taken: 8 } },
      [`POST /draft/${ID}/join`]: {
        status: 409,
        body: { error: "every seat is taken", code: "draft_full" },
      },
    });
    const { container } = render(Draft as never, { id: ID, invite: "AbC" } as never);
    await settle();
    container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(text(container.querySelector('[role="alert"]'))).toContain("every seat is taken");
    expect(FakeSocket.opened).toHaveLength(0);
  });

  it("has nothing to join with for a link with no invite and no seat", async () => {
    serve({});
    const { container } = render(Draft as never, { id: ID } as never);
    await settle();
    expect(text(container.querySelector("h1"))).toBe(L.draftGone);
    expect(FakeSocket.opened).toHaveLength(0);
    expect(sent("GET", `/draft/${ID}`)).toHaveLength(0);
  });
});

// ---- The draft's lobby ----

describe("the draft lobby", () => {
  const lobby = (over: Partial<DraftView> = {}): DraftView =>
    view({
      state: "lobby",
      round: 0,
      pick: 0,
      take: 0,
      pack: undefined,
      picks: [],
      deadline: undefined,
      now: undefined,
      seats: 4,
      invite: "AbC",
      players: [
        { seat: 0, name: "Ian", host: true, connected: true },
        { seat: 1, name: "Sam", connected: true },
      ],
      ...over,
    });

  it("dials the draft socket with the stored seat token", async () => {
    await room(null);
    expect(FakeSocket.opened).toHaveLength(1);
    expect(socket().url).toMatch(new RegExp(`^wss?://[^/]+/draft/${ID}/ws\\?token=seat-token$`));
  });

  it("shows the seats taken and open, and gives the host the invite link and Start", async () => {
    const c = await room(lobby());
    const seats = [...c.querySelectorAll(`[aria-label="${L.draftSeats}"] li`)].map(text);
    expect(seats).toHaveLength(4);
    expect(seats[0]).toContain("Ian");
    expect(seats[0]).toContain("you");
    expect(seats[0]).toContain("host");
    expect(seats[1]).toContain("Sam");
    expect(seats[2]).toContain("Open seat");
    expect(seats[2]).toContain("a bot will sit here");

    const link = c.querySelector<HTMLInputElement>(`[aria-label="${L.draftInviteLink}"]`)!;
    expect(link.value).toContain(`#/draft/${ID}?invite=AbC`);
    expect(link.readOnly).toBe(true);

    const write = vi.fn(async () => {});
    vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText: write } });
    click(button(c, L.draftCopyInvite));
    await settle();
    expect(write).toHaveBeenCalledWith(link.value);
    expect(text(main(c))).toContain("Copied.");

    click(button(c, L.draftStart));
    await settle();
    const start = sent("POST", `/draft/${ID}/start`);
    expect(start).toHaveLength(1);
    expect(start[0].headers.Authorization).toBe("Bearer seat-token");
    expect(text(c.querySelector('[data-testid="packs-approximate"]'))).toBe(
      PACKS_ARE_APPROXIMATIONS,
    );
  });

  it("gives a solo draft no invite link", async () => {
    const c = await room(
      lobby({
        solo: true,
        invite: undefined,
        players: [{ seat: 0, name: "Player", host: true, connected: true }],
      }),
    );
    expect(c.querySelector(`[aria-label="${L.draftInviteLink}"]`)).toBeNull();
    expect(text(main(c))).not.toContain(L.draftCopyInvite);
    expect(text(main(c))).toContain("solo draft");
    expect(text(button(c, L.draftStart))).toBe(L.draftStart);
  });

  it("gives a guest of the draft neither the link nor Start", async () => {
    const c = await room(lobby({ you: 1, invite: undefined }));
    expect(c.querySelector(`[aria-label="${L.draftInviteLink}"]`)).toBeNull();
    expect(text(main(c))).not.toContain(L.draftStart);
    expect(text(main(c))).toContain("Ian starts the draft");
  });

  it("leaves: posts with the seat token, forgets it and goes back to the form", async () => {
    const c = await room(lobby({ you: 1, invite: undefined }));
    click(button(c, L.draftLeave));
    await settle();
    const leave = sent("POST", `/draft/${ID}/leave`);
    expect(leave).toHaveLength(1);
    expect(leave[0].headers.Authorization).toBe("Bearer seat-token");
    expect(storedDraftSeat(ID)).toBeNull();
    expect(location.hash).toBe("#/draft");
  });
});

// ---- The pick ----

describe("the pick screen", () => {
  it("draws the pack as named buttons with the cards' images", async () => {
    const c = await room();
    const cards = cardButtons(c);
    expect(cards.map((b) => b.getAttribute("aria-label")!.split(",")[0])).toEqual([
      "Shoreline Looter",
      "Bonecache Overseer",
      "Fountainport Bell",
    ]);
    expect(cards.every((b) => b.getAttribute("aria-pressed") === "false")).toBe(true);
    // The public draft image route, never a session-gated one.
    const imgs = [...c.querySelectorAll<HTMLImageElement>(`[aria-label="${L.draftPack}"] img`)];
    expect(imgs.map((i) => i.getAttribute("src"))).toEqual(
      PACK.map((p) => `/draft/image/${p.scryfall_id}?size=normal`),
    );
    expect(text(c.querySelector('[data-testid="draft-progress"]'))).toBe(
      "Pack 2 of 3 · pick 5 of 14 · passing right to Bot 3",
    );
  });

  it("sends a draft_pick with the chosen card once it is confirmed", async () => {
    const c = await room();
    const confirm = button(c, "Pick 1 card");
    expect(confirm.disabled).toBe(true);
    click(cardButtons(c)[1]);
    expect(cardButtons(c)[1].getAttribute("aria-pressed")).toBe("true");
    expect(socket().picks()).toHaveLength(0);
    expect(confirm.disabled).toBe(false);
    click(confirm);
    const picks = socket().picks();
    expect(picks).toHaveLength(1);
    expect(picks[0].payload).toEqual({ cards: ["c2"] });
    // In flight: no second pick can be sent.
    expect(button(c, "Pick 1 card").disabled).toBe(true);
  });

  it("moves the choice with one card to take", async () => {
    const c = await room();
    click(cardButtons(c)[0]);
    click(cardButtons(c)[2]);
    expect(cardButtons(c).map((b) => b.getAttribute("aria-pressed"))).toEqual([
      "false",
      "false",
      "true",
    ]);
    click(button(c, "Pick 1 card"));
    expect(socket().picks()[0].payload).toEqual({ cards: ["c3"] });
  });

  it("picks at once on a double-click, and on Enter on the chosen card", async () => {
    let c = await room();
    cardButtons(c)[0].dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));
    flushSync();
    expect(socket().picks()[0].payload).toEqual({ cards: ["c1"] });
    cleanup();

    FakeSocket.opened = [];
    c = await room();
    const enter = (): boolean =>
      cardButtons(c)[1].dispatchEvent(
        new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
      );
    // Enter on a card that is not chosen is left to the button (it chooses).
    expect(enter()).toBe(true);
    expect(socket().picks()).toHaveLength(0);
    click(cardButtons(c)[1]);
    expect(enter()).toBe(false);
    flushSync();
    expect(socket().picks()[0].payload).toEqual({ cards: ["c2"] });
  });

  it("needs two cards in pick two", async () => {
    const c = await room(view({ mode: "pick_two", take: 2 }));
    const confirm = button(c, "Pick 2 cards");
    expect(confirm.disabled).toBe(true);
    click(cardButtons(c)[0]);
    expect(confirm.disabled).toBe(true);
    click(confirm);
    expect(socket().picks()).toHaveLength(0);
    // A double-click never short-cuts a two-card pick.
    cardButtons(c)[1].dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));
    flushSync();
    expect(socket().picks()).toHaveLength(0);

    click(cardButtons(c)[2]);
    expect(confirm.disabled).toBe(false);
    // A third choice replaces the oldest.
    click(cardButtons(c)[1]);
    expect(cardButtons(c).map((b) => b.getAttribute("aria-pressed"))).toEqual([
      "false",
      "true",
      "true",
    ]);
    click(confirm);
    expect(socket().picks()[0].payload).toEqual({ cards: ["c3", "c2"] });
  });

  it("takes one when one card is left in a pick-two pack", async () => {
    const c = await room(view({ mode: "pick_two", take: 1, pack: [PACK[0]] }));
    click(cardButtons(c)[0]);
    click(button(c, "Pick 1 card"));
    expect(socket().picks()[0].payload).toEqual({ cards: ["c1"] });
  });

  it("shows a refusal and lets the person choose again", async () => {
    const c = await room();
    click(cardButtons(c)[0]);
    click(button(c, "Pick 1 card"));
    socket().frame(
      "error",
      { code: "draft_pick_refused", message: "that card is not in your pack" },
      socket().picks()[0].id,
    );
    expect(text(c.querySelector('[role="alert"]'))).toContain("that card is not in your pack");
    expect(cardButtons(c)[0].disabled).toBe(false);
    expect(button(c, "Pick 1 card").disabled).toBe(false);
  });

  it("waits for the others once the pick has landed, and names them", async () => {
    const c = await room();
    click(cardButtons(c)[0]);
    click(button(c, "Pick 1 card"));
    const after = view({
      seq: 13,
      pack: undefined,
      take: 0,
      picks: [card("p1", "Thundertrap Trainer"), PACK[0]],
    });
    after.players = after.players.map((p) => ({ ...p, picking: p.seat >= 2 }));
    socket().frame("draft", after);
    await settle();

    expect(c.querySelector(`[aria-label="${L.draftPack}"]`)).toBeNull();
    const waiting = c.querySelector('[data-testid="draft-waiting"]')!;
    expect(text(waiting)).toContain(L.draftWaiting);
    expect(text(waiting)).toContain("Waiting for Sam and Bot 3.");
    // Focus follows the pick to the line that says what happens next.
    expect(document.activeElement).toBe(waiting.querySelector("h3"));
    expect(text(c.querySelector('[data-testid="draft-announce"]'))).toBe(
      "Pick made. Waiting for Sam and Bot 3.",
    );
    // The picks list has the new card and a count.
    expect(text(c.querySelector("#draft-picks-h"))).toBe(`${L.draftYourPicks} 2`);
    expect(text(c.querySelector('[data-testid="draft-pool"]'))).toContain("Shoreline Looter");
  });

  it("groups the picks by colour, with mana value and a count", async () => {
    const picks = [
      card("p1", "Thundertrap Trainer", { mana_value: 2 }),
      card("p2", "Mindwhisker", { mana_value: 3, mana_cost: "{2}{U}" }),
      card("p3", "Forest", {
        colors: [],
        mana_cost: undefined,
        mana_value: 0,
        type_line: "Basic Land — Forest",
        basic: true,
      }),
      card("p4", "Forest", {
        colors: [],
        mana_cost: undefined,
        mana_value: 0,
        type_line: "Basic Land — Forest",
        basic: true,
      }),
    ];
    const c = await room(view({ picks }));
    const groups = [...c.querySelectorAll('[data-testid="draft-pool"] section')];
    expect(groups.map((g) => g.getAttribute("data-group"))).toEqual(["U", "land"]);
    expect(text(groups[0].querySelector("h3"))).toBe(`${L.draftGroupBlue} 2`);
    const rows = [...groups[0].querySelectorAll("li")];
    expect(rows.map((li) => text(li.querySelector(".mv")))).toEqual([
      "mana value 2",
      "mana value 3",
    ]);
    expect(rows.map((li) => text(li.querySelector(".name")))).toEqual([
      "Thundertrap Trainer",
      "Mindwhisker",
    ]);
    // The cost is drawn as pips with a name of their own.
    expect(rows[1].querySelector('[role="img"]')?.getAttribute("data-cost")).toBe("{2}{U}");
    expect(text(groups[1])).toContain("2× Forest");
    expect(text(c.querySelector("#draft-picks-h"))).toBe(`${L.draftYourPicks} 4`);
  });

  it("lists every seat with who is picking, who has picked and who is away", async () => {
    const c = await room();
    const rows = [...c.querySelectorAll<HTMLElement>(`[aria-label="${L.draftSeats}"] li`)];
    expect(rows.map((r) => r.dataset.status)).toEqual(["picking", "picked", "away", "picking"]);
    expect(text(rows[0])).toContain("Ian");
    expect(text(rows[1])).toContain("bot");
    expect(text(rows[1])).toContain("picked");
    expect(text(rows[1])).toContain("5 cards");
    expect(text(rows[2])).toContain("away");
  });

  it("counts the timer down from the server's clock and turns urgent in the last 10 seconds", async () => {
    vi.useFakeTimers({ toFake: ["setInterval", "clearInterval", "performance", "Date"] });
    serve({});
    rememberDraftSeat(ID, { token: "seat-token", expires_at: "2099-01-01T00:00:00Z" });
    // The viewer's clock is years off; the count must not care.
    vi.setSystemTime(new Date("2031-05-05T00:00:00Z"));
    const { container: c } = render(Draft as never, { id: ID } as never);
    flushSync();
    socket().frame("draft", view());

    const timer = (): HTMLElement =>
      c.querySelector<HTMLElement>(`[aria-label="${L.draftTimer}"]`)!;
    expect(timer().getAttribute("role")).toBe("timer");
    expect(timer().dataset.seconds).toBe("43");
    expect(text(timer())).toContain("0:43");
    expect(timer().classList.contains("urgent")).toBe(false);

    vi.advanceTimersByTime(13_000);
    flushSync();
    expect(timer().dataset.seconds).toBe("30");
    expect(text(c.querySelector('[data-testid="draft-announce"]'))).toBe(
      "30 seconds left to pick.",
    );

    vi.advanceTimersByTime(20_000);
    flushSync();
    expect(timer().dataset.seconds).toBe("10");
    expect(timer().classList.contains("urgent")).toBe(true);
    expect(text(timer())).toContain("pick now");
    expect(text(c.querySelector('[data-testid="draft-announce"]'))).toBe(
      "10 seconds left to pick.",
    );

    // Between the marks the live region is left alone.
    vi.advanceTimersByTime(4_000);
    flushSync();
    expect(text(c.querySelector('[data-testid="draft-announce"]'))).toBe(
      "10 seconds left to pick.",
    );
    vi.advanceTimersByTime(60_000);
    flushSync();
    expect(timer().dataset.seconds).toBe("0");
  });

  it("asks before leaving a draft that is picking", async () => {
    const c = await room();
    click(button(c, L.draftLeave));
    await settle();
    expect(sent("POST", `/draft/${ID}/leave`)).toHaveLength(0);
    expect(text(main(c))).toContain("A bot takes your seat for the rest of the draft");
    click(button(c, L.draftStay));
    expect(text(main(c))).not.toContain("A bot takes your seat");
    click(button(c, L.draftLeave));
    click(button(c, L.draftLeaveConfirm));
    await settle();
    expect(sent("POST", `/draft/${ID}/leave`)).toHaveLength(1);
    expect(location.hash).toBe("#/draft");
  });
});

// ---- Hidden information ----

describe("another seat's cards (CR 905.1c)", () => {
  it("renders nothing about another seat but its name, status and count", async () => {
    // The server never sends these. A view that carried them anyway
    // must still draw none of them: the page reads only its own
    // `pack`, `picks` and `pool`.
    const leaky = view();
    const secret = [card("s1", "Secret Rare"), card("s2", "Hidden Bomb")];
    Object.assign(leaky.players[1], { pack: secret, picks: secret, pool: secret, cards: secret });
    Object.assign(leaky, { packs: [secret], others: secret });
    const c = await room(leaky);
    expect(text(c)).not.toContain("Secret Rare");
    expect(text(c)).not.toContain("Hidden Bomb");
    expect(c.innerHTML).not.toContain("sf-s1");
    const bot = text(c.querySelector(`[aria-label="${L.draftSeats}"] li[data-seat="1"]`));
    expect(bot).toBe("2 Bot 1 bot picked 5 cards");

    socket().frame("draft", { ...leaky, state: "done", seq: 99, pack: undefined, pool: [PACK[0]] });
    expect(text(c)).not.toContain("Secret Rare");
    expect(text(c)).toContain("Shoreline Looter");
  });

  it("reads cards only from the viewer's own pack, picks and pool", () => {
    // vitest runs from client/; under jsdom import.meta.url is no file URL.
    const here = join(process.cwd(), "src/routes");
    const dir = join(here, "../lib/components/draft");
    const files = [
      join(here, "Draft.svelte"),
      ...readdirSync(dir).map((f) => join(dir, f)),
      join(here, "../lib/draft.ts"),
      join(here, "../lib/draftSocket.ts"),
    ];
    expect(files.length).toBeGreaterThan(6);
    // Every `.pack`, `.picks` or `.pool` read is on the view itself (or
    // the previous and next view), never on a player or anything else.
    const VIEWS = new Set(["view", "prev", "next", "v"]);
    const reads: string[] = [];
    for (const file of files) {
      const src = readFileSync(file, "utf8");
      for (const m of src.matchAll(/([A-Za-z_$][\w$]*)\??\.(pack|picks|pool)\b/g)) {
        if (!VIEWS.has(m[1])) reads.push(`${file}: ${m[0]}`);
      }
      // A player row is never asked for cards under any name.
      expect(src, file).not.toMatch(
        /\b(p|player|players\[[^\]]*\])\??\.(cards|pack|picks|pool|hand)\b/,
      );
    }
    expect(reads).toEqual([]);
    // The seat list reads no card field at all.
    const seats = readFileSync(join(dir, "DraftSeats.svelte"), "utf8");
    expect(seats).not.toMatch(/\.(pack|picks|pool)\b|DraftCard\b|scryfall/);
  });
});

// ---- Done ----

describe("a finished draft", () => {
  const done = (): DraftView =>
    view({
      state: "done",
      seq: 80,
      pack: undefined,
      take: 0,
      deadline: undefined,
      now: undefined,
      picks: PACK,
      pool: [
        ...PACK,
        card("p9", "Forest", {
          colors: [],
          mana_cost: undefined,
          mana_value: 0,
          type_line: "Basic Land — Forest",
          basic: true,
        }),
      ],
    });

  it("shows the pool, grouped, and says deckbuilding comes next without faking it", async () => {
    const c = await room(done());
    const panel = c.querySelector('[data-testid="draft-done"]')!;
    expect(text(panel.querySelector("h2"))).toBe(L.draftFinished);
    expect(text(panel)).toContain(L.draftDeckbuildingNext);
    expect(text(panel)).toContain("not ready yet");
    expect(text(panel)).toContain("You drafted 4 cards from Bloomburrow.");
    // It is coming, not here: nothing on the page pretends to build a deck.
    expect(main(c).querySelectorAll("button")).toHaveLength(0);
    const links = [...main(c).querySelectorAll("a")].map(text);
    expect(links.some((l) => /deck/i.test(l))).toBe(false);
    expect(links.some((l) => l.includes(L.draftNewDraft))).toBe(true);

    expect(text(c.querySelector("#draft-pool-h"))).toBe(`${L.draftYourPool} 4`);
    const groups = [...c.querySelectorAll('[data-testid="draft-pool"] section')];
    expect(groups.map((g) => g.getAttribute("data-group"))).toEqual([
      "U",
      "B",
      "colorless",
      "land",
    ]);
    const imgs = [...c.querySelectorAll<HTMLImageElement>('[data-testid="draft-pool"] img')];
    expect(imgs.map((i) => i.alt)).toEqual([
      "Shoreline Looter",
      "Bonecache Overseer",
      "Fountainport Bell",
      "Forest",
    ]);
    expect(imgs[0].getAttribute("src")).toBe("/draft/image/sf-c1?size=normal");
    expect(c.querySelector(`[aria-label="${L.draftTimer}"]`)).toBeNull();
    expect(text(c.querySelector('[data-testid="draft-announce"]'))).toBe(`${L.draftFinished}.`);
  });

  it("keeps the pool on screen when the finished draft closes", async () => {
    const c = await room(done());
    socket().drop(1000, "draft closed");
    expect(text(c.querySelector("h1"))).toBe("Friday draft");
    expect(text(main(c))).toContain("This draft has closed.");
    expect(text(c.querySelector('[data-testid="draft-pool"]'))).toContain("Shoreline Looter");
  });
});

// ---- Losing the draft, or the connection ----

describe("a draft that is gone, and a lost connection", () => {
  it("says the draft is gone when the server closes it, and does not redial", async () => {
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "setInterval", "clearInterval"] });
    serve({});
    rememberDraftSeat(ID, { token: "seat-token", expires_at: "2099-01-01T00:00:00Z" });
    const { container: c } = render(Draft as never, { id: ID } as never);
    flushSync();
    socket().frame("draft", view());
    socket().drop(1000, "draft closed");
    expect(text(c.querySelector("h1"))).toBe(L.draftGone);
    expect(text(main(c))).toContain("This draft was closed");
    expect(c.querySelector(`[aria-label="${L.draftPack}"]`)).toBeNull();
    expect(storedDraftSeat(ID)).toBeNull();
    vi.advanceTimersByTime(120_000);
    expect(FakeSocket.opened).toHaveLength(1);
  });

  it("forgets the seat and says so plainly when the server no longer knows it", async () => {
    serve({ [`GET /draft/${ID}`]: { status: 404, body: { error: "not found" } } });
    rememberDraftSeat(ID, { token: "seat-token", expires_at: "2099-01-01T00:00:00Z" });
    const { container: c } = render(Draft as never, { id: ID } as never);
    await settle();
    // The upgrade is refused; a browser sees only 1006.
    socket().drop(1006);
    await settle();
    const asked = sent("GET", `/draft/${ID}`);
    expect(asked).toHaveLength(1);
    expect(asked[0].headers.Authorization).toBe("Bearer seat-token");
    expect(storedDraftSeat(ID)).toBeNull();
    expect(text(c.querySelector("h1"))).toBe(L.draftGone);
    expect(text(main(c))).toContain("This browser holds no seat in that draft");
  });

  it("offers the invite page again when a dead seat's link still carries its invite", async () => {
    serve({ [`GET /draft/${ID}`]: { status: 401, body: { error: "unauthorized" } } });
    rememberDraftSeat(ID, { token: "dead", expires_at: "2099-01-01T00:00:00Z" });
    const { container: c } = render(Draft as never, { id: ID, invite: "AbC" } as never);
    await settle();
    socket().drop(1006);
    await settle();
    expect(storedDraftSeat(ID)).toBeNull();
    // The join page asked with the invite this time.
    expect(sent("GET", `/draft/${ID}`).map((x) => x.url)).toContain(`/draft/${ID}?invite=AbC`);
    expect(c.querySelector(`[aria-label="${L.draftPack}"]`)).toBeNull();
  });

  it("says it is reconnecting, keeps the picks, and blocks the pick until it is back", async () => {
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "setInterval", "clearInterval"] });
    serve({});
    rememberDraftSeat(ID, { token: "seat-token", expires_at: "2099-01-01T00:00:00Z" });
    const { container: c } = render(Draft as never, { id: ID } as never);
    flushSync();
    socket().frame("draft", view());
    click(cardButtons(c)[0]);
    socket().drop(1006);

    const note = c.querySelector('[data-testid="draft-reconnecting"]')!;
    expect(note.getAttribute("role")).toBe("status");
    expect(text(note)).toContain("Connection lost. Reconnecting…");
    expect(text(note)).toContain("a card is picked for you when the timer runs out");
    expect(text(c.querySelector('[data-testid="draft-pool"]'))).toContain("Thundertrap Trainer");
    expect(button(c, "Pick 1 card").disabled).toBe(true);

    vi.advanceTimersByTime(1000);
    expect(FakeSocket.opened).toHaveLength(2);
    socket().frame("draft", view({ seq: 13 }));
    expect(c.querySelector('[data-testid="draft-reconnecting"]')).toBeNull();
    // The choice made before the drop is still there.
    expect(cardButtons(c)[0].getAttribute("aria-pressed")).toBe("true");
    expect(button(c, "Pick 1 card").disabled).toBe(false);
  });

  it("shows the server's sentence when the socket is refused at the connection cap", async () => {
    const c = await room(null);
    socket().frame("error", {
      code: "server_full",
      message: "The server is full right now. Please try again in a few minutes.",
    });
    expect(text(main(c))).toContain(
      "The server is full right now. Please try again in a few minutes.",
    );
  });
});
