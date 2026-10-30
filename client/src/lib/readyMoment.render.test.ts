// @vitest-environment jsdom
//
// ADR 0212 §2, the ready-check moment, on the lobby card: when this client
// sees the last seated player's deck land, the tile's check lands in seat
// order, the counter fills, Start game gets one soft ring, and the polite
// live region says "Every deck is in. You can start the game." once.
// Under any of the three reduced-motion signals (DESIGN.md §4) nothing
// moves and the line is the same. The rule for WHEN it fires is in
// readyMoment.test.ts; this file pins what the page does with it.

import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Lobby from "../routes/Lobby.svelte";
import { READY_ANNOUNCE_HOLD_MS, READY_MOMENT_MS } from "./readyMoment";
import { currentSession, setSession, type Session } from "./session";
import { settings } from "./settings";
import { DEFAULT_TABLE_SETTINGS } from "./tableSettings";
import { cleanup, flushSync, render } from "./test/render.svelte";

const USER = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const GAME = "11111111-2222-3333-4444-555555555555";
const ALICE = "a1a1a1a1-0000-4000-8000-000000000001";
const BOB = "b0b0b0b0-0000-4000-8000-000000000002";
const CARA = "c0c0c0c0-0000-4000-8000-000000000003";

function seated(playerID: string, role: Session["principal"]["role"] = "player"): Session {
  const exp = new Date(Date.now() + 86_400_000).toISOString();
  const admin = role === "admin";
  return {
    token: `tok-${playerID.slice(0, 4)}`,
    expiresAt: exp,
    principal: {
      role,
      user_id: admin ? undefined : USER,
      name: "Viewer",
      issued_at: new Date().toISOString(),
      expires_at: exp,
      game_id: admin ? undefined : GAME,
      player_id: admin ? undefined : playerID,
    },
    gameID: admin ? undefined : GAME,
    playerID: admin ? undefined : playerID,
  } as Session;
}

interface Who {
  id: string;
  name: string;
  seat: number;
  deck: boolean;
  host?: boolean;
}

const alice = (deck = true): Who => ({ id: ALICE, name: "Alice", seat: 0, deck, host: true });
const bob = (deck = true): Who => ({ id: BOB, name: "Bob", seat: 1, deck });
const cara = (deck = true): Who => ({ id: CARA, name: "Cara", seat: 2, deck });

// The table the server answers with, which a test changes between reads.
let current: Record<string, unknown> = {};

function setTable(who: Who[], over: Record<string, unknown> = {}): void {
  current = {
    id: GAME,
    name: "Friday",
    created_at: new Date().toISOString(),
    state: "lobby",
    is_creator: false,
    host_player_id: ALICE,
    players: who.map((w) => ({
      player_id: w.id,
      name: w.name,
      seat: w.seat,
      deck_uploaded: w.deck,
      deck_name: w.deck ? `${w.name}'s deck` : undefined,
      is_host: w.host === true,
    })),
    ...over,
  };
}

function stubServer(): void {
  const routes: Record<string, () => unknown> = {
    "GET /games": () => ({ games: [current] }),
    [`GET /games/${GAME}`]: () => ({ ...current, invite_token: "inv", spectator_invite: "spec" }),
    "GET /me": () => ({ admin: false }),
    "GET /me/setup": () => ({ setup: null }),
    "GET /decks": () => ({ decks: [] }),
    [`GET /games/${GAME}/settings`]: () => ({
      settings: { ...DEFAULT_TABLE_SETTINGS, starting_life: 40, undo_limit: 1 },
    }),
  };
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const route = routes[`${init.method ?? "GET"} ${input}`];
      const payload = route ? route() : {};
      return {
        ok: true,
        status: 200,
        statusText: "OK",
        json: async () => payload,
        clone() {
          return this;
        },
      } as unknown as Response;
    }),
  );
}

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    for (let j = 0; j < 5; j++) await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

// reread is what a poll does: the page lists the tables again.
async function reread(): Promise<void> {
  window.dispatchEvent(new Event("focus"));
  await settle();
}

const card = (c: HTMLElement): HTMLElement => c.querySelector<HTMLElement>("li.tcard")!;
const liveRegion = (c: HTMLElement): HTMLElement =>
  c.querySelector<HTMLElement>('p.sr-only[role="status"]')!;
const said = (c: HTMLElement): string => liveRegion(c).textContent?.trim() ?? "";
const primary = (c: HTMLElement): HTMLButtonElement =>
  card(c).querySelector<HTMLButtonElement>(".tfoot .btn.primary")!;
const counter = (c: HTMLElement): string =>
  card(c).querySelector(".tfoot .rcount")?.textContent?.trim() ?? "";
const tile = (c: HTMLElement, name: string): HTMLElement =>
  card(c).querySelector<HTMLElement>(`li[aria-label="${name}"]`)!;

// Everything the live region said, in order: each time its text became
// something other than empty. A screen reader announces an insertion, so
// this is what it would have read out.
function listen(c: HTMLElement): string[] {
  const lines: string[] = [];
  let last = "";
  new MutationObserver(() => {
    const now = said(c);
    if (now && now !== last) lines.push(now);
    // an emptied region and a refilled one with the same words is a new line
    if (!now) last = "";
    else last = now;
  }).observe(liveRegion(c), { childList: true, characterData: true, subtree: true });
  return lines;
}

const START_LINE = "Every deck is in. You can start the game.";

let savedSettings = get(settings);

beforeEach(() => {
  localStorage.clear();
  location.hash = "#/lobby";
  savedSettings = get(settings);
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
  Reflect.deleteProperty(window, "matchMedia");
  settings.set(savedSettings);
  setSession(null);
  localStorage.clear();
});

describe("the ready-check moment on the lobby card (ADR 0212 §2)", () => {
  it("says nothing when the page opens on a table that is already ready", async () => {
    setSession(seated(ALICE));
    setTable([alice(), bob()], { is_creator: true });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const lines = listen(container);

    expect(counter(container)).toBe("2/2");
    expect(primary(container).textContent?.trim()).toBe("Start game · 2 players");
    expect(said(container)).toBe("");
    expect(card(container).classList.contains("moment")).toBe(false);
    expect(primary(container).classList.contains("ringed")).toBe(false);

    // And not on the polls that follow.
    await reread();
    await reread();
    expect(lines).toEqual([]);
    expect(card(container).classList.contains("moment")).toBe(false);
  });

  it("when the last deck lands: rings that seat, the primary, and says one line, once", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    setSession(seated(ALICE));
    setTable([alice(), bob(false)], { is_creator: true });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const lines = listen(container);

    expect(counter(container)).toBe("1/2");
    expect(primary(container).textContent?.trim()).toBe("Waiting for 1 player");
    expect(said(container)).toBe("");
    expect(card(container).classList.contains("moment")).toBe(false);

    // Bob's deck lands: the next poll sees it.
    setTable([alice(), bob(true)], { is_creator: true });
    await reread();

    expect(counter(container)).toBe("2/2");
    expect(primary(container).textContent?.trim()).toBe("Start game · 2 players");
    expect(said(container)).toBe(START_LINE);
    expect(lines).toEqual([START_LINE]);
    expect(card(container).classList.contains("moment")).toBe(true);
    // The tile whose deck landed gets the ring; the others only pop their check.
    expect(tile(container, "seat 2: Bob").classList.contains("landed")).toBe(true);
    expect(tile(container, "seat 1: Alice").classList.contains("landed")).toBe(false);
    expect(tile(container, "seat 1: Alice").classList.contains("moment")).toBe(true);
    expect(primary(container).classList.contains("ringed")).toBe(true);

    // More polls of the same ready table: no second line, no second ring.
    await reread();
    await reread();
    expect(lines).toEqual([START_LINE]);

    // The classes go once the keyframes have played; the line is emptied
    // after its hold, so a screen reader's browse mode is not left a stale one.
    vi.advanceTimersByTime(READY_MOMENT_MS + 50);
    flushSync();
    expect(card(container).classList.contains("moment")).toBe(false);
    expect(tile(container, "seat 2: Bob").classList.contains("landed")).toBe(false);
    expect(primary(container).classList.contains("ringed")).toBe(false);
    expect(said(container)).toBe(START_LINE);
    vi.advanceTimersByTime(READY_ANNOUNCE_HOLD_MS);
    flushSync();
    expect(said(container)).toBe("");
    expect(lines).toEqual([START_LINE]);
  });

  it("says the line again when the table leaves all-ready and comes back", async () => {
    setSession(seated(ALICE));
    setTable([alice(), bob(false)], { is_creator: true });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const lines = listen(container);

    setTable([alice(), bob()], { is_creator: true });
    await reread();
    expect(lines).toEqual([START_LINE]);

    // Cara sits down with no deck yet: not ready, and nothing to say.
    setTable([alice(), bob(), cara(false)], { is_creator: true });
    await reread();
    expect(counter(container)).toBe("2/3");
    expect(lines).toEqual([START_LINE]);

    // Her deck lands: a full transition, so a second line.
    setTable([alice(), bob(), cara()], { is_creator: true });
    await reread();
    expect(counter(container)).toBe("3/3");
    expect(lines).toEqual([START_LINE, START_LINE]);
    expect(tile(container, "seat 3: Cara").classList.contains("landed")).toBe(true);
  });

  it("is silent when a seat leaves or joins ready and the table stays startable", async () => {
    setSession(seated(ALICE));
    setTable([alice(), bob(), cara()], { is_creator: true });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const lines = listen(container);

    setTable([alice(), bob()], { is_creator: true }); // Cara left
    await reread();
    setTable([alice(), bob(), cara()], { is_creator: true }); // and sat down again, deck in
    await reread();
    expect(lines).toEqual([]);
    expect(card(container).classList.contains("moment")).toBe(false);
  });

  it("tells a player who may not start who the table waits for, and rings nothing", async () => {
    setSession(seated(BOB));
    setTable([alice(), bob(), cara(false)]);
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const lines = listen(container);

    setTable([alice(), bob(), cara()]);
    await reread();

    expect(lines).toEqual(["Every deck is in. Waiting for Alice to start."]);
    expect(primary(container).textContent?.trim()).toBe("Waiting for Alice to start");
    expect(primary(container).disabled).toBe(true);
    // The wave and the landed seat's ring still play; the disabled primary is not ringed.
    expect(card(container).classList.contains("moment")).toBe(true);
    expect(primary(container).classList.contains("ringed")).toBe(false);
  });

  it("stays quiet for an admin who is not at the table, though the card still changes", async () => {
    setSession(seated(ALICE, "admin"));
    setTable([alice(), bob(false)]);
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    expect(currentSession()?.principal.role).toBe("admin");
    const lines = listen(container);

    setTable([alice(), bob()]);
    await reread();

    expect(counter(container)).toBe("2/2");
    expect(primary(container).textContent?.trim()).toBe("Start game · 2 players");
    expect(lines).toEqual([]);
    expect(said(container)).toBe("");
    expect(card(container).classList.contains("moment")).toBe(true);
  });

  describe("under reduced motion, nothing moves and the line is the same", () => {
    const signals: [string, () => void][] = [
      [
        "the setting accessibility.reduceMotion",
        () =>
          settings.update((s) => ({
            ...s,
            accessibility: { ...s.accessibility, reduceMotion: true },
          })),
      ],
      [
        "the animations master switch off",
        () => settings.update((s) => ({ ...s, animations: { ...s.animations, enabled: false } })),
      ],
      [
        "the OS's prefers-reduced-motion",
        () =>
          Object.defineProperty(window, "matchMedia", {
            configurable: true,
            value: (q: string) => ({
              matches: q.includes("prefers-reduced-motion"),
              media: q,
              addEventListener: () => undefined,
              removeEventListener: () => undefined,
            }),
          }),
      ],
    ];

    it.each(signals)("%s", async (_name, reduce) => {
      reduce();
      setSession(seated(ALICE));
      setTable([alice(), bob(false)], { is_creator: true });
      stubServer();
      const { container } = render(Lobby as never, {} as never);
      await settle();
      const lines = listen(container);

      setTable([alice(), bob()], { is_creator: true });
      await reread();

      // The checks, the counter and the button simply changed...
      expect(tile(container, "seat 2: Bob").textContent).toMatch(/Ready/);
      expect(counter(container)).toBe("2/2");
      expect(primary(container).textContent?.trim()).toBe("Start game · 2 players");
      // ...with no class to animate...
      expect(card(container).classList.contains("moment")).toBe(false);
      expect(card(container).querySelector(".landed, .ringed, .tile.moment")).toBeNull();
      // ...and the announcement is the same.
      expect(lines).toEqual([START_LINE]);
    });
  });

  it("names the table when the person takes part in more than one", async () => {
    setSession(seated(ALICE));
    const other = "99999999-2222-3333-4444-555555555555";
    const twin = (id: string, name: string, who: Who[]) => ({
      id,
      name,
      created_at: new Date().toISOString(),
      state: "lobby",
      is_creator: true,
      host_player_id: ALICE,
      players: who.map((w) => ({
        player_id: w.id,
        name: w.name,
        seat: w.seat,
        deck_uploaded: w.deck,
        is_host: w.host === true,
      })),
    });
    let second = twin(other, "Saturday", [alice(), bob(false)]);
    setTable([alice(), bob(false)], { is_creator: true });
    stubServer();
    const base = vi.mocked(fetch).getMockImplementation()!;
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: string, init?: RequestInit) => {
        if (input === "/games" && (init?.method ?? "GET") === "GET") {
          return {
            ok: true,
            status: 200,
            statusText: "OK",
            json: async () => ({ games: [current, second] }),
            clone() {
              return this;
            },
          } as unknown as Response;
        }
        return base(input, init);
      }),
    );
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const lines = listen(container);

    second = twin(other, "Saturday", [alice(), bob()]);
    await reread();

    expect(lines).toEqual(["Every deck is in at Saturday. Waiting for Alice to start."]);
  });
});
