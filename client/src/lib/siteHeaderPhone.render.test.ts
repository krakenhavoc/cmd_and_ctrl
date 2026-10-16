// @vitest-environment jsdom
//
// siteHeaderPhone.render.test.ts — the site header's phone layout (#3877).
// Below 600px the nav links sit behind a menu button: a disclosure with
// aria-expanded, focus moved into the panel, Escape to close, and a
// close on a page change. jsdom applies no component CSS, so this
// covers the markup and the behaviour; tests-e2e's public-header-phone
// spec covers the one-row layout at phone widths in a real browser.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import SiteHeader from "./components/SiteHeader.svelte";
import { L } from "./labels";
import { session, type Session } from "./session";
import { _resetCommunityForTests, _setCommunityForTests } from "./community";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const INVITE = "https://discord.gg/abc123";
const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";

function identified(): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: "tok",
    expiresAt: expires,
    principal: {
      role: "identified",
      user_id: USER,
      name: "Alice",
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

const menuButton = (c: HTMLElement) =>
  c.querySelector<HTMLButtonElement>(`button[aria-label="${L.siteMenu}"]`)!;
const panel = (c: HTMLElement) => c.querySelector<HTMLElement>("nav#site-nav")!;

function press(key: string): void {
  window.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true }));
  flushSync();
}

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

describe("the site header's menu button", () => {
  it("is a closed disclosure for the nav, which it names", async () => {
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    const btn = menuButton(container);
    expect(btn.getAttribute("aria-expanded")).toBe("false");
    expect(btn.getAttribute("aria-controls")).toBe("site-nav");
    expect(panel(container).getAttribute("aria-label")).toBe("Site");
    expect(panel(container).classList.contains("open")).toBe(false);
  });

  it("opens on a click, puts focus on the first link, and closes on Escape with focus back on the button", async () => {
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    const btn = menuButton(container);
    click(btn);
    await settle();
    expect(btn.getAttribute("aria-expanded")).toBe("true");
    expect(panel(container).classList.contains("open")).toBe(true);
    expect(document.activeElement).toBe(panel(container).querySelector("a"));

    press("Escape");
    expect(btn.getAttribute("aria-expanded")).toBe("false");
    expect(panel(container).classList.contains("open")).toBe(false);
    expect(document.activeElement).toBe(btn);
  });

  it("closes on a click outside the panel, and on one of its links", async () => {
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    const btn = menuButton(container);
    click(btn);
    await settle();
    click(document.body);
    flushSync();
    expect(btn.getAttribute("aria-expanded")).toBe("false");

    click(btn);
    await settle();
    // A link to the page already open changes no route: the click closes it.
    const link = panel(container).querySelector<HTMLAnchorElement>("a")!;
    link.addEventListener("click", (e) => e.preventDefault());
    click(link);
    flushSync();
    expect(btn.getAttribute("aria-expanded")).toBe("false");
  });

  it("keeps every nav link, signed out and signed in, inside the panel", async () => {
    const out = render(SiteHeader as never, {} as never);
    await settle();
    expect(
      [...panel(out.container).querySelectorAll(".grp a")].map((a) => a.textContent?.trim()),
    ).toEqual(["Decks", "Catalog", "Roadmap", "Guide", "About"]);
    out.destroy();

    session.set(identified());
    const inn = render(SiteHeader as never, {} as never);
    await settle();
    expect(
      [...panel(inn.container).querySelectorAll(".grp a")].map((a) => a.textContent?.trim()),
    ).toEqual([L.navTables, "Stats", "Decks", "Catalog", "Roadmap"]);
  });

  it("keeps Sign in and the account button in the row, outside the panel", async () => {
    const out = render(SiteHeader as never, {} as never);
    await settle();
    const signIn = out.container.querySelector("a.sign-in")!;
    expect(signIn).not.toBeNull();
    expect(panel(out.container).contains(signIn)).toBe(false);
    out.destroy();

    session.set(identified());
    const inn = render(SiteHeader as never, {} as never);
    await settle();
    const acct = inn.container.querySelector('button[aria-controls="account-menu"]')!;
    expect(acct).not.toBeNull();
    expect(panel(inn.container).contains(acct)).toBe(false);
  });

  it("opening the account menu closes the nav panel", async () => {
    session.set(identified());
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    click(menuButton(container));
    await settle();
    click(container.querySelector('button[aria-controls="account-menu"]')!);
    await settle();
    expect(menuButton(container).getAttribute("aria-expanded")).toBe("false");
  });

  it("carries the C2 Discord as a labelled row in the panel only when the server names an invite", async () => {
    const none = render(SiteHeader as never, {} as never);
    await settle();
    expect(panel(none.container).querySelector(".panel-util")).toBeNull();
    none.destroy();

    _setCommunityForTests({ invite_url: INVITE, gate: false });
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    const row = panel(container).querySelector<HTMLAnchorElement>(
      `.panel-util a[href="${INVITE}"]`,
    );
    expect(row?.textContent?.trim()).toBe("C2 Discord");
    expect(row?.getAttribute("rel")).toContain("noopener");
  });
});
