// @vitest-environment jsdom
//
// helpMenu.render.test.ts — Help and the way in (ADR 0125 §6, Delivery
// PR 5): the site header's `button "help"` and its `menu "help"`, for
// everyone; Settings → Advanced's "Tips and the tutorial"; Home's
// Practice game tile; and the tutorial offer (`lobby.practice`), which
// goes to a guest or a signed-in person with no finished game, and never
// to the admin token.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get, writable } from "svelte/store";
import { tick } from "svelte";

import type { Route } from "./router";

const routeStore = vi.hoisted(() => ({
  value: null as null | import("svelte/store").Writable<Route>,
  navigate: null as null | ((hash: string) => void),
}));
vi.mock("./router", async (importOriginal) => {
  const orig = await importOriginal<typeof import("./router")>();
  routeStore.value = writable<Route>({ name: "lobby" });
  routeStore.navigate = vi.fn();
  return { ...orig, route: routeStore.value, navigate: routeStore.navigate };
});

import SiteHeader from "./components/SiteHeader.svelte";
import HelpMenu from "./components/HelpMenu.svelte";
import Settings from "./components/Settings.svelte";
import Home from "../routes/Home.svelte";
import Lobby from "../routes/Lobby.svelte";
import { L } from "./labels";
import { HINTS } from "./hints";
import { anchorOf, emptyContext, holds, type Hint, type HintContext } from "./hints/hint";
import { _resetHintsForTests, hintEngine } from "./hints/runtime";
import { _resetEndedGameForTests, anyEnded, hasEndedGameFor, noteMyGames } from "./hints/endedGame";
import { candidates } from "./hints/queue";
import { defaultSettings, openSettings, settings, settingsOpen } from "./settings";
import { shortcutsHelpOpen } from "./shortcutRuntime";
import { session, type Session } from "./session";
import { _resetForTests as resetModals } from "./modalLayers";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const OTHER = "6c1e7b4f-9a8a-4f1f-8c2b-1a4d2e3f5b6c";

function sess(role: Session["principal"]["role"], userID?: string): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: `${role}-tok`,
    expiresAt: expires,
    principal: {
      role,
      user_id: userID,
      name: role === "admin" ? "admin" : "Alice",
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
  } as Session;
}

const byID = (id: string): Hint => {
  const h = HINTS.find((x) => x.id === id);
  if (!h) throw new Error(`no hint ${id}`);
  return h;
};

const helpButton = () =>
  document.querySelector<HTMLButtonElement>(`button[aria-label="${L.help}"]`);
const helpMenu = () => document.querySelector<HTMLElement>("#help-menu");
const menuItems = () => [...(helpMenu()?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])];
const itemNamed = (name: string) =>
  menuItems().find((el) => (el.textContent ?? "").trim() === name);

async function settle(): Promise<void> {
  await tick();
  await tick();
  flushSync();
}

async function openHelp(): Promise<void> {
  click(helpButton()!);
  await settle();
}

beforeEach(() => {
  localStorage.clear();
  settings.set(defaultSettings());
  settingsOpen.set(false);
  shortcutsHelpOpen.set(false);
  routeStore.value!.set({ name: "lobby" });
  vi.mocked(routeStore.navigate!).mockClear();
  session.set(null);
  _resetHintsForTests({ log: () => {} });
  _resetEndedGameForTests();
  resetModals();
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => new Response("{}", { status: 404 })),
  );
});

afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  session.set(null);
  vi.unstubAllGlobals();
});

describe('the site header\'s Help: button "help" and its menu', () => {
  it("is there for everyone, signed in or not, between the site nav and the account control", () => {
    for (const s of [null, sess("identified", USER), sess("admin")]) {
      session.set(s);
      const r = render(SiteHeader as never, {} as never);
      const btn = helpButton()!;
      expect(btn, String(s?.principal.role)).not.toBeNull();
      expect(btn.getAttribute("aria-haspopup")).toBe("menu");
      expect(btn.getAttribute("aria-expanded")).toBe("false");
      const nav = r.container.querySelector("nav.site-nav")!;
      // After the nav in document order, and before the account control.
      expect(nav.compareDocumentPosition(btn) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
      const acct =
        r.container.querySelector(".acct-btn") ?? r.container.querySelector('a[href="#/login"]');
      expect(btn.compareDocumentPosition(acct!) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
      r.destroy();
    }
  });

  it("opens a menu named by its button, with the four items in order, focus on the first", async () => {
    session.set(sess("identified", USER));
    render(HelpMenu as never, {} as never);
    await openHelp();
    const menu = helpMenu()!;
    expect(menu.getAttribute("role")).toBe("menu");
    expect(menu.getAttribute("aria-labelledby")).toBe(helpButton()!.id);
    expect(helpButton()!.getAttribute("aria-expanded")).toBe("true");
    expect(menuItems().map((el) => (el.textContent ?? "").trim())).toEqual([
      "Tips for this page",
      "Show all tips again",
      "Practice game",
      "Keyboard shortcuts",
    ]);
    expect(document.activeElement).toBe(menuItems()[0]);
  });

  it("Practice game opens a practice table, and signed out goes to sign in first", async () => {
    session.set(sess("player", USER));
    const r = render(HelpMenu as never, {} as never);
    await openHelp();
    const signedIn = itemNamed("Practice game")!;
    expect(signedIn.getAttribute("href")).toBe("#/practice");
    click(signedIn);
    expect(routeStore.navigate).toHaveBeenLastCalledWith("#/practice");
    expect(helpMenu()).toBeNull();
    r.destroy();

    session.set(null);
    render(HelpMenu as never, {} as never);
    await openHelp();
    const signedOut = itemNamed("Practice game (sign in first)")!;
    expect(signedOut.getAttribute("href")).toBe("#/login");
    click(signedOut);
    expect(routeStore.navigate).toHaveBeenLastCalledWith("#/login");
  });

  it("Tips for this page forgets this page's tips and the site's, and replays them with tips off", async () => {
    const page: Hint = {
      id: "lobby.example",
      version: 2,
      place: "lobby",
      order: 0,
      anchor: { label: L.lobbyTitle },
      title: "A lobby tip",
      body: "It says where something is.",
    };
    const site = byID("site.help");
    const elsewhere: Hint = { ...page, id: "decks.example", place: "decks" };
    settings.update((s) => ({
      ...s,
      help: {
        tipsOff: true,
        seen: { "lobby.example": 2, "site.help": 1, "decks.example": 2 },
      },
    }));
    render(HelpMenu as never, { hints: [page, site, elsewhere] } as never);
    await openHelp();
    click(itemNamed("Tips for this page")!);
    flushSync();
    expect(helpMenu()).toBeNull();
    // This page's and the site's are forgotten; another page's are not,
    // and tips stay off: a replay is what the person asked for.
    expect(get(settings).help).toEqual({ tipsOff: true, seen: { "decks.example": 2 } });
    // The engine offers them, in order, despite tips off.
    const env = {
      hints: [page, site, elsewhere],
      ctx: emptyContext("lobby"),
      visit: "route#1",
      seen: get(settings).help.seen,
      tipsOff: true,
      now: 1,
      resolves: () => true,
    };
    expect(hintEngine().tick(env)?.id).toBe("lobby.example");
    hintEngine().dismiss(2);
    expect(hintEngine().tick({ ...env, now: 3 })?.id).toBe("site.help");
  });

  it("has no tips for a page that offers none (sign in), and says so", async () => {
    routeStore.value!.set({ name: "login" });
    render(HelpMenu as never, {} as never);
    await openHelp();
    const item = itemNamed("Tips for this page") as HTMLButtonElement;
    expect(item.disabled).toBe(true);
    expect(item.title).toBe("this page has no tips");
    // Home has none of its own, but the site's tip is replayed there.
    cleanup();
    document.body.innerHTML = "";
    routeStore.value!.set({ name: "home" });
    render(HelpMenu as never, {} as never);
    await openHelp();
    expect((itemNamed("Tips for this page") as HTMLButtonElement).disabled).toBe(false);
  });

  it("Show all tips again forgets every tip and switches tips back on", async () => {
    settings.update((s) => ({
      ...s,
      help: { tipsOff: true, seen: { "site.help": 1, "lobby.practice": 1 } },
    }));
    render(HelpMenu as never, {} as never);
    await openHelp();
    click(itemNamed("Show all tips again")!);
    flushSync();
    expect(get(settings).help).toEqual({ tipsOff: false, seen: {} });
    expect(helpMenu()).toBeNull();
  });

  it("Keyboard shortcuts opens the ? overlay", async () => {
    render(HelpMenu as never, {} as never);
    await openHelp();
    click(itemNamed("Keyboard shortcuts")!);
    flushSync();
    expect(get(shortcutsHelpOpen)).toBe(true);
    expect(helpMenu()).toBeNull();
  });

  it("closes on Escape with focus back on the button, and arrows move between items", async () => {
    render(HelpMenu as never, {} as never);
    await openHelp();
    const menu = helpMenu()!;
    menu.dispatchEvent(new KeyboardEvent("keydown", { key: "End", bubbles: true }));
    expect(document.activeElement).toBe(menuItems().at(-1));
    menu.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowDown", bubbles: true }));
    expect(document.activeElement).toBe(menuItems()[0]);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(helpMenu()).toBeNull();
    expect(document.activeElement).toBe(helpButton());
  });

  it("closes on a click outside it", async () => {
    render(HelpMenu as never, {} as never);
    await openHelp();
    document.body.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    flushSync();
    expect(helpMenu()).toBeNull();
  });
});

describe("Settings → Advanced: Tips and the tutorial", () => {
  function openAdvanced(): HTMLElement {
    render(Settings as never, {} as never);
    openSettings("advanced");
    flushSync();
    const h = [...document.querySelectorAll("h4")].find(
      (el) => el.textContent === "Tips and the tutorial",
    );
    expect(h).toBeDefined();
    return h!.parentElement!;
  }
  const button = (root: HTMLElement, name: string) =>
    [...root.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => (b.textContent ?? "").trim() === name,
    );

  it("comes first in Advanced, above Export", () => {
    const section = openAdvanced();
    const export_ = [...document.querySelectorAll("h4")].find((el) => el.textContent === "Export")!;
    expect(
      section.compareDocumentPosition(export_) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it("Show tips is the inverse of Hide tips", () => {
    settings.update((s) => ({ ...s, help: { ...s.help, tipsOff: true } }));
    const section = openAdvanced();
    const box = [...section.querySelectorAll("label")]
      .find((l) => (l.textContent ?? "").includes("Show tips"))!
      .querySelector<HTMLInputElement>('input[type="checkbox"]')!;
    expect(box.checked).toBe(false);
    box.click();
    flushSync();
    expect(get(settings).help.tipsOff).toBe(false);
    expect(box.checked).toBe(true);
    box.click();
    flushSync();
    expect(get(settings).help.tipsOff).toBe(true);
  });

  it("Show all tips again forgets every tip and switches tips back on", () => {
    settings.update((s) => ({ ...s, help: { tipsOff: true, seen: { "site.help": 1 } } }));
    const section = openAdvanced();
    click(button(section, "Show all tips again")!);
    expect(get(settings).help).toEqual({ tipsOff: false, seen: {} });
    expect(section.textContent).toContain("every tip will show again");
  });

  it("Replay the tutorial closes Settings and opens a practice table", () => {
    session.set(sess("identified", USER));
    const section = openAdvanced();
    click(button(section, "Replay the tutorial")!);
    expect(get(settingsOpen)).toBe(false);
    expect(routeStore.navigate).toHaveBeenLastCalledWith("#/practice");
  });

  it("signed out, Replay the tutorial goes to sign in first", () => {
    const section = openAdvanced();
    click(button(section, "Replay the tutorial (sign in first)")!);
    expect(routeStore.navigate).toHaveBeenLastCalledWith("#/login");
  });
});

describe("Home's Help group", () => {
  const tile = (name: string) =>
    [...document.querySelectorAll<HTMLAnchorElement>("a.tile")].find(
      (a) => a.querySelector(".tile-title")?.textContent === name,
    );

  it("has a Practice game tile, first, beside the bot guide", () => {
    session.set(sess("identified", USER));
    render(Home as never, {} as never);
    const help = [...document.querySelectorAll("section.group")].find(
      (s) => s.querySelector("h2")?.textContent === "Help",
    )!;
    const titles = [...help.querySelectorAll(".tile-title")].map((t) => t.textContent);
    expect(titles.slice(0, 2)).toEqual(["Practice game", "Bot guide"]);
    expect(tile("Practice game")!.getAttribute("href")).toBe("#/practice");
    expect(tile("Practice game")!.querySelector(".pill")).toBeNull();
  });

  it("signed out, it goes to sign in and says so", () => {
    render(Home as never, {} as never);
    const t = tile("Practice game")!;
    expect(t.getAttribute("href")).toBe("#/login");
    expect(t.querySelector(".pill")?.textContent).toBe("Sign in first");
  });
});

describe("the tutorial offer (lobby.practice)", () => {
  const offer = byID("lobby.practice");
  const ctx = (over: Partial<HintContext>): HintContext => ({
    ...emptyContext("lobby"),
    ...over,
  });
  const offered = (c: HintContext) => holds(offer, c) && anchorOf(offer, c) !== null;

  it("goes to a guest and to a signed-in person with no finished game", () => {
    expect(offered(ctx({}))).toBe(true);
    expect(offered(ctx({ signedIn: true, hasEndedGame: false }))).toBe(true);
    expect(offered(ctx({ signedIn: true, adminMode: true, hasEndedGame: false }))).toBe(true);
  });

  it("never to someone who has finished a game, nor to the admin token", () => {
    expect(holds(offer, ctx({ signedIn: true, hasEndedGame: true }))).toBe(false);
    expect(holds(offer, ctx({ adminToken: true }))).toBe(false);
  });

  it("waits for the read of /me/games rather than guessing, and is not passed over meanwhile", () => {
    const unknown = ctx({ signedIn: true, hasEndedGame: null });
    // Still a candidate (so the visit waits for it) but with no anchor yet.
    expect(holds(offer, unknown)).toBe(true);
    expect(anchorOf(offer, unknown)).toBeNull();
    const cands = candidates({ hints: HINTS, ctx: unknown, seen: {}, tipsOff: false });
    expect(cands[0]?.id).toBe("lobby.practice");
  });

  it("anchors to the Lobby's title, and its action starts practice with Not now beside it", () => {
    expect(anchorOf(offer, ctx({}))).toEqual({ label: L.lobbyTitle });
    expect(offer.action).toEqual({ label: "Start practice", href: "#/practice" });
  });

  it("site.help anchors to the Help button", () => {
    expect(anchorOf(byID("site.help"), ctx({}))).toEqual({ label: L.help });
  });
});

describe("whether the person has finished a game (lib/hints/endedGame.ts)", () => {
  it("reads an ended table from GET /me/games", () => {
    expect(anyEnded([])).toBe(false);
    expect(anyEnded([{ state: "lobby" }, { state: "active" }])).toBe(false);
    expect(anyEnded([{ state: "active" }, { state: "ended" }])).toBe(true);
  });

  it("is unknown until a read lands, kept per user, and null for a guest or the token", () => {
    const me = sess("identified", USER);
    expect(hasEndedGameFor(me)).toBeNull();
    noteMyGames(USER, [{ state: "ended" }]);
    expect(hasEndedGameFor(me)).toBe(true);
    expect(hasEndedGameFor(sess("player", USER))).toBe(true);
    // Another person on the same tab starts from unknown.
    expect(hasEndedGameFor(sess("identified", OTHER))).toBeNull();
    noteMyGames(USER, [{ state: "active" }]);
    expect(hasEndedGameFor(me)).toBe(false);
    expect(hasEndedGameFor(sess("admin"))).toBeNull();
    expect(hasEndedGameFor(null)).toBeNull();
    noteMyGames(null, [{ state: "ended" }]);
    expect(hasEndedGameFor(me)).toBe(false);
  });

  it("is filled by the Lobby's own read of /me/games, with no new request", async () => {
    const fetches: string[] = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string) => {
        fetches.push(url);
        const body =
          url === "/me/games"
            ? { games: [{ id: "g1", state: "ended", others: [] }] }
            : url === "/games"
              ? { games: [] }
              : {};
        return new Response(JSON.stringify(body), { status: 200 });
      }),
    );
    session.set(sess("identified", USER));
    render(Lobby as never, {} as never);
    for (let i = 0; i < 20; i++) await Promise.resolve();
    await settle();
    expect(fetches.filter((u) => u === "/me/games")).toHaveLength(1);
    expect(hasEndedGameFor(get(session))).toBe(true);
  });
});
