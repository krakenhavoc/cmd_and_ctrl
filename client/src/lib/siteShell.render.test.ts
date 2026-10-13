// @vitest-environment jsdom
//
// siteShell.render.test.ts — the one site shell (ADR 0156 §3, §6): the
// header's grouped links for each kind of session, the C2 Discord in the
// header and the footer only when the server names an invite, the
// footer's one row, and the page frame's head.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createRawSnippet, tick } from "svelte";

import SiteHeader from "./components/SiteHeader.svelte";
import SiteFooter from "./components/SiteFooter.svelte";
import Page from "./components/Page.svelte";
import { L } from "./labels";
import { session, type Session } from "./session";
import { _resetCommunityForTests, _setCommunityForTests } from "./community";
import { TIP_SLOT_ATTR } from "./hints/runtime";
import { cleanup, flushSync, render } from "./test/render.svelte";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const INVITE = "https://discord.gg/abc123";

function sess(role: Session["principal"]["role"], userID?: string): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: `${role}-tok`,
    expiresAt: expires,
    principal: {
      role,
      user_id: userID,
      name: role === "admin" ? "admin" : "Alice",
      game_id: role === "player" ? "g1" : undefined,
      player_id: role === "player" ? "p1" : undefined,
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
  } as Session;
}

async function settle(): Promise<void> {
  await tick();
  await tick();
  flushSync();
}

const navGroups = (c: HTMLElement) =>
  [...c.querySelectorAll('nav[aria-label="Site"] .grp')].map((g) =>
    [...g.querySelectorAll("a")].map((a) => a.textContent?.trim()),
  );
const discordLink = (c: HTMLElement) => c.querySelector<HTMLAnchorElement>(`a[href="${INVITE}"]`);

beforeEach(() => {
  localStorage.clear();
  session.set(null);
  _resetCommunityForTests();
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => new Response("{}", { status: 404 })),
  );
});

afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  session.set(null);
  _resetCommunityForTests();
  vi.unstubAllGlobals();
});

describe("the site header's links", () => {
  it("signed out: Decks, Roadmap and About, then Sign in; no Home and no Tables", async () => {
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    expect(navGroups(container)).toEqual([["Decks", "Roadmap", "About"]]);
    const signIn = container.querySelector("a.sign-in");
    expect(signIn?.textContent?.trim()).toBe("Sign in");
    // No Discord sign-in on this stub server: Sign in goes to the front page.
    expect(signIn?.getAttribute("href")).toBe("#/");
    expect(container.querySelector('nav a[href="#/home"]')).toBeNull();
    expect(container.querySelector('button[aria-controls="account-menu"]')).toBeNull();
  });

  it("signed in with Discord: Tables and My games | Decks, Catalog and Roadmap", async () => {
    session.set(sess("identified", USER));
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    expect(navGroups(container)).toEqual([
      [L.navTables, "My games"],
      ["Decks", "Catalog", "Roadmap"],
    ]);
    // The route is still the Lobby's; only the words changed.
    const tables = [...container.querySelectorAll("nav a")].find(
      (a) => a.textContent?.trim() === L.navTables,
    );
    expect(tables?.getAttribute("href")).toBe("#/lobby");
    expect(container.querySelector("a.sign-in")).toBeNull();
    expect(container.querySelector('button[aria-controls="account-menu"]')).not.toBeNull();
  });

  it("a guest seat and the admin token get no My games, which needs a Discord account", async () => {
    for (const s of [sess("player"), sess("admin")]) {
      session.set(s);
      const { container, destroy } = render(SiteHeader as never, {} as never);
      await settle();
      expect(navGroups(container)).toEqual([[L.navTables], ["Decks", "Catalog", "Roadmap"]]);
      destroy();
    }
  });

  it("marks the current page", async () => {
    session.set(sess("identified", USER));
    location.hash = "#/catalog";
    await new Promise((r) => setTimeout(r, 0));
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    const current = container.querySelector('nav a[aria-current="page"]');
    expect(current?.textContent?.trim()).toBe("Catalog");
    location.hash = "";
  });
});

describe("the C2 Discord in the shell", () => {
  it("is hidden in the header and the footer while the server names no invite", async () => {
    const header = render(SiteHeader as never, {} as never);
    const footer = render(SiteFooter as never, {} as never);
    await settle();
    expect(header.container.querySelector('[aria-label="the C2 Discord server"]')).toBeNull();
    expect(footer.container.textContent).not.toContain("Discord");
  });

  it("is the header's icon link and the footer's first link when there is one", async () => {
    _setCommunityForTests({ invite_url: INVITE, gate: false });
    const header = render(SiteHeader as never, {} as never);
    const footer = render(SiteFooter as never, {} as never);
    await settle();
    const icon = discordLink(header.container);
    expect(icon?.getAttribute("aria-label")).toBe("the C2 Discord server");
    expect(icon?.getAttribute("target")).toBe("_blank");
    expect(icon?.getAttribute("rel")).toContain("noopener");
    const links = [...footer.container.querySelectorAll("nav a")].map((a) => a.textContent?.trim());
    expect(links).toEqual(["Discord", "About", "Privacy", "GitHub"]);
    expect(discordLink(footer.container)).not.toBeNull();
  });
});

describe("the page frame", () => {
  const body = createRawSnippet(() => ({ render: () => "<p class='body'>the page</p>" }));
  const actions = createRawSnippet(() => ({ render: () => "<button>Do it</button>" }));

  it("has the header, an eyebrow, one h1, the lede, the actions, a tip slot and the body", async () => {
    const { container } = render(
      Page as never,
      {
        eyebrow: "Play",
        title: "Tables",
        titleLabel: L.lobbyTitle,
        lede: "Create a table.",
        actions,
        children: body,
      } as never,
    );
    await settle();
    expect(container.querySelector("header.site-header")).not.toBeNull();
    expect(container.querySelector(".page-eyebrow")?.textContent).toBe("Play");
    const h1s = container.querySelectorAll("h1");
    expect(h1s).toHaveLength(1);
    expect(h1s[0].textContent).toBe("Tables");
    expect(h1s[0].getAttribute("aria-label")).toBe(L.lobbyTitle);
    expect(container.querySelector(".page-lede")?.textContent).toBe("Create a table.");
    expect(container.querySelector(".page-actions button")?.textContent).toBe("Do it");
    expect(container.querySelector(`[${TIP_SLOT_ATTR}]`)).not.toBeNull();
    expect(container.querySelector("main.site-page.wide p.body")).not.toBeNull();
  });

  it("is prose width when asked, and has no eyebrow or lede it was not given", async () => {
    const { container } = render(
      Page as never,
      { title: "About", width: "prose", children: body } as never,
    );
    await settle();
    expect(container.querySelector("main.site-page.prose")).not.toBeNull();
    expect(container.querySelector(".page-eyebrow")).toBeNull();
    expect(container.querySelector(".page-lede")).toBeNull();
    expect(container.querySelector(".page-actions")).toBeNull();
  });
});
