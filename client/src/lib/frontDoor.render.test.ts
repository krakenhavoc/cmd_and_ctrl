// @vitest-environment jsdom
//
// frontDoor.render.test.ts — ADR 0156 §4–6, rendered: the front page
// with the gate on and off and with no invite; the gate page and its
// ?retry=1; the Lobby's empty state; every Discord placement hidden
// when GET /community names no invite; and the guest practice game a
// signed-out visitor opens from #/practice.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import FrontPage from "../routes/FrontPage.svelte";
import RequestAccess from "../routes/RequestAccess.svelte";
import Lobby from "../routes/Lobby.svelte";
import Practice from "../routes/Practice.svelte";
import SiteHeader from "./components/SiteHeader.svelte";
import SiteFooter from "./components/SiteFooter.svelte";
import { _resetCommunityForTests, _setCommunityForTests, type Community } from "./community";
import { currentSession, setSession, type Session } from "./session";
import { endPractice } from "./practiceTable";
import { L } from "./labels";
import { cleanup, flushSync, render } from "./test/render.svelte";

const INVITE = "https://discord.gg/z7x9Ehpnjw";
const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const GAME = "77777777-7777-4777-8777-777777777777";

type Handler = () => { status?: number; body: unknown };
let routes: Record<string, Handler> = {};
let calls: Array<{ method: string; url: string; auth: string | null }> = [];

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      calls.push({ method, url: input, auth: new Headers(init.headers).get("Authorization") });
      const h = routes[`${method} ${input}`];
      const out = h ? h() : { status: 404, body: { error: "not found" } };
      const status = out.status ?? 200;
      return {
        ok: status < 400,
        status,
        statusText: String(status),
        json: async () => out.body,
        clone() {
          return this;
        },
      } as unknown as Response;
    }),
  );
}

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    for (let j = 0; j < 5; j++) await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

function community(c: Community): void {
  _setCommunityForTests(c);
  flushSync();
}

const text = (el: ParentNode) => (el.textContent ?? "").replace(/\s+/g, " ");
const links = (el: ParentNode) => [...el.querySelectorAll("a")];
const linkNamed = (el: ParentNode, name: string) => links(el).find((a) => text(a).trim() === name);
const discordLinks = (el: ParentNode) => links(el).filter((a) => a.getAttribute("href") === INVITE);

beforeEach(() => {
  routes = {
    "GET /auth/discord/config": () => ({ body: { enabled: true } }),
    "GET /games": () => ({ body: { games: [] } }),
    "GET /me/games": () => ({ body: { games: [] } }),
  };
  calls = [];
  localStorage.clear();
  _resetCommunityForTests();
  stubServer();
});

afterEach(async () => {
  endPractice();
  cleanup();
  setSession(null);
  _resetCommunityForTests();
  vi.unstubAllGlobals();
  localStorage.clear();
  location.hash = "";
  await settle();
});

describe("the front page (§4)", () => {
  it("leads with sign-in and the Discord server side by side, and the three steps with the gate on", async () => {
    community({ invite_url: INVITE, gate: true });
    const { container } = render(FrontPage as never, {} as never);
    await settle();

    const ctas = container.querySelector(".ctas")!;
    expect(links(ctas).map((a) => text(a).trim())).toEqual([
      L.signInWithDiscord,
      L.joinDiscordServer,
    ]);
    expect(linkNamed(ctas, L.signInWithDiscord)!.getAttribute("href")).toBe("/auth/discord/start");
    expect(linkNamed(ctas, L.joinDiscordServer)!.getAttribute("href")).toBe(INVITE);
    expect(linkNamed(container, L.tryPracticeGame)!.getAttribute("href")).toBe("#/practice");

    const steps = [...container.querySelectorAll(".steps h2")].map((h) => h.textContent);
    expect(steps).toEqual([
      "Join the C2 Discord",
      "Request account access",
      "Sign in and sit down",
    ]);
    expect(text(container)).toContain("Elder Dragon Highlander");

    // The four cards, the invite box first, with its labels unchanged.
    const cards = [...container.querySelectorAll(".cards h2")].map((h) => h.textContent);
    expect(cards).toEqual(["Have an invite?", "Check a deck", "Roadmap", "New here?"]);
    expect(container.querySelector('input[aria-label="invite code or link"]')).not.toBeNull();
    expect(linkNamed(container, "Decks")!.getAttribute("href")).toBe("#/decks");
    expect(container.querySelector('a[href="#/privacy"]')).not.toBeNull();
    expect(container.querySelector('input[placeholder="admin token"]')).toBeNull();
  });

  it("with the gate off shows only the last step, and a line inviting people to the server", async () => {
    community({ invite_url: INVITE, gate: false });
    const { container } = render(FrontPage as never, {} as never);
    await settle();
    expect(container.querySelector(".steps")).toBeNull();
    const panel = container.querySelector(".steps-panel")!;
    expect([...panel.querySelectorAll("h2")].map((h) => h.textContent)).toEqual([
      "Sign in and sit down",
    ]);
    expect(text(panel)).not.toContain("Elder Dragon Highlander");
    expect(discordLinks(panel)).toHaveLength(1);
  });

  it("says nothing about Discord with no invite and the gate off", async () => {
    community({ invite_url: null, gate: false });
    const { container } = render(FrontPage as never, {} as never);
    await settle();
    expect(discordLinks(container)).toHaveLength(0);
    expect(linkNamed(container, L.joinDiscordServer)).toBeUndefined();
    expect(text(container.querySelector(".steps-panel")!)).not.toMatch(/Discord/);
    // Sign-in and the practice game stay.
    expect(linkNamed(container, L.signInWithDiscord)).toBeDefined();
    expect(linkNamed(container, L.tryPracticeGame)).toBeDefined();
  });

  it("offers no sign-in button when the server has no Discord sign-in", async () => {
    routes["GET /auth/discord/config"] = () => ({ body: { enabled: false } });
    const { container } = render(FrontPage as never, {} as never);
    await settle();
    expect(linkNamed(container, L.signInWithDiscord)).toBeUndefined();
    expect(container.querySelector(".ctas")).toBeNull();
  });
});

describe("the header's Sign in, signed out", () => {
  it("starts Discord sign-in directly, as the hero does", async () => {
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    expect(container.querySelector("a.sign-in")?.getAttribute("href")).toBe("/auth/discord/start");
  });

  it("goes to the front page when the server has no Discord sign-in", async () => {
    routes["GET /auth/discord/config"] = () => ({ body: { enabled: false } });
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    expect(container.querySelector("a.sign-in")?.getAttribute("href")).toBe("#/");
  });
});

describe("the gate page (§5)", () => {
  it("walks the four onboarding steps, with Join, Check again and a practice game", async () => {
    community({ invite_url: INVITE, gate: true });
    const { container } = render(RequestAccess as never, {} as never);
    await settle();
    expect(container.querySelector("h1")?.textContent).toBe("Ask for an account on our Discord");
    const steps = [...container.querySelectorAll(".steps h2")].map((h) => h.textContent);
    expect(steps).toEqual([
      "Join the C2 Discord",
      "Answer yes to “request account access”",
      "Ask in the access-request channel",
      "Check again",
    ]);
    expect(text(container)).toContain("Channels & Roles");
    expect(text(container)).toContain("Elder Dragon Highlander");
    expect(discordLinks(container.querySelector(".steps")!)).toHaveLength(1);
    expect(linkNamed(container, "Check again")!.getAttribute("href")).toBe("/auth/discord/start");
    expect(linkNamed(container, L.tryPracticeGame)!.getAttribute("href")).toBe("#/practice");
    expect(text(container)).toMatch(/invite link from a host still works/);
  });

  it("leads with Discord not answering on ?retry=1", async () => {
    const { container } = render(RequestAccess as never, { retry: true } as never);
    await settle();
    expect(container.querySelector("h1")?.textContent).toBe("Discord didn't answer — try again");
    // Check again in the head, as well as in the steps.
    const head = container.querySelector(".page-head")!;
    expect(linkNamed(head, "Check again")!.getAttribute("href")).toBe("/auth/discord/start");
  });

  it("hides the Join button with no invite", async () => {
    community({ invite_url: null, gate: true });
    const { container } = render(RequestAccess as never, {} as never);
    await settle();
    expect(discordLinks(container)).toHaveLength(0);
    expect(linkNamed(container, "Join")).toBeUndefined();
  });
});

describe("the Lobby's empty state (§6)", () => {
  function signedIn(): Session {
    const exp = new Date(Date.now() + 86_400_000).toISOString();
    return {
      token: "id-tok",
      expiresAt: exp,
      principal: {
        role: "identified",
        user_id: USER,
        name: "Alice",
        issued_at: new Date().toISOString(),
        expires_at: exp,
      },
    } as Session;
  }

  it("has Practice game, Bring a deck and No pod yet?, which links the invite", async () => {
    community({ invite_url: INVITE, gate: true });
    setSession(signedIn());
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const cards = container.querySelector(".empty-cards")!;
    expect([...cards.querySelectorAll("h3")].map((h) => h.textContent)).toEqual([
      "Practice game",
      "Bring a deck",
      "No pod yet?",
    ]);
    const pod = linkNamed(cards, L.findAPod)!;
    expect(pod.getAttribute("href")).toBe(INVITE);
    expect(pod.getAttribute("target")).toBe("_blank");
  });

  it("drops the No pod yet? card with no invite", async () => {
    setSession(signedIn());
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const cards = container.querySelector(".empty-cards")!;
    expect([...cards.querySelectorAll("h3")].map((h) => h.textContent)).toEqual([
      "Practice game",
      "Bring a deck",
    ]);
    expect(discordLinks(container)).toHaveLength(0);
  });
});

describe("every Discord placement hides with no invite (§6)", () => {
  it("in the header and the footer", async () => {
    community({ invite_url: INVITE, gate: true });
    let h = render(SiteHeader as never, {} as never);
    let f = render(SiteFooter as never, {} as never);
    await settle();
    expect(discordLinks(h.container)).toHaveLength(1);
    expect(discordLinks(f.container)).toHaveLength(1);
    h.destroy();
    f.destroy();

    community({ invite_url: null, gate: true });
    h = render(SiteHeader as never, {} as never);
    f = render(SiteFooter as never, {} as never);
    await settle();
    expect(h.container.querySelector('a[target="_blank"][href*="discord"]')).toBeNull();
    expect(f.container.querySelector('a[href*="discord"]')).toBeNull();
  });
});

describe("a guest practice game (§2)", () => {
  it("opens from #/practice with no session, through POST /games/practice/guest", async () => {
    const exp = new Date(Date.now() + 43_200_000).toISOString();
    routes["POST /games/practice/guest"] = () => ({
      status: 201,
      body: {
        token: "guest-practice-tok",
        expires_at: exp,
        principal: {
          role: "player",
          name: "Player",
          game_id: GAME,
          player_id: "p1",
          issued_at: new Date().toISOString(),
          expires_at: exp,
        },
        game: { id: GAME, practice: true },
        player_id: "p1",
      },
    });
    render(Practice as never, {} as never);
    await settle();

    const post = calls.filter((c) => c.method === "POST");
    expect(post.map((c) => c.url)).toEqual(["/games/practice/guest"]);
    expect(post[0].auth).toBeNull();
    expect(currentSession()?.token).toBe("guest-practice-tok");
    expect(location.hash).toBe(`#/games/${GAME}`);
  });

  it("uses the session's own route when there is one", async () => {
    const exp = new Date(Date.now() + 43_200_000).toISOString();
    setSession({
      token: "seat-tok",
      expiresAt: exp,
      principal: {
        role: "identified",
        user_id: USER,
        name: "Alice",
        issued_at: new Date().toISOString(),
        expires_at: exp,
      },
    } as Session);
    routes["POST /games/practice"] = () => ({
      status: 201,
      body: {
        token: "practice-tok",
        expires_at: exp,
        principal: { role: "player", game_id: GAME, player_id: "p1" },
        game: { id: GAME },
        player_id: "p1",
      },
    });
    render(Practice as never, {} as never);
    await settle();
    expect(calls.filter((c) => c.method === "POST").map((c) => c.url)).toEqual(["/games/practice"]);
  });

  it("says why and links back to the front page when it fails", async () => {
    routes["POST /games/practice/guest"] = () => ({
      status: 429,
      body: { error: "too many practice games from here; try again in a few minutes" },
    });
    const { container } = render(Practice as never, {} as never);
    await settle();
    expect(container.querySelector('[role="alert"]')?.textContent).toMatch(/too many practice/);
    expect(linkNamed(container, "Back to the front page")!.getAttribute("href")).toBe("#/");
    expect(currentSession()).toBeNull();
  });
});
