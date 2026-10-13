// @vitest-environment jsdom
//
// The getting-started guide (#/guide), rendered: public, its sections and
// the places it links, the Discord parts only when the server names an
// invite, and the ways in from the footer, the front page and Help.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";
import { get } from "svelte/store";

import App from "../App.svelte";
import SiteFooter from "./components/SiteFooter.svelte";
import HelpMenu from "./components/HelpMenu.svelte";
import FrontPage from "../routes/FrontPage.svelte";
import Guide from "../routes/Guide.svelte";
import { GUIDE_HASH } from "./about";
import { _resetCommunityForTests, _setCommunityForTests } from "./community";
import { navigate, route } from "./router";
import { session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

vi.mock("../routes/Game.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));
vi.mock("../routes/Practice.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

const realFetch = globalThis.fetch;

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

beforeEach(() => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  globalThis.fetch = vi.fn(async () => new Response("{}", { status: 200 })) as never;
});

afterEach(() => {
  cleanup();
  _resetCommunityForTests();
  globalThis.fetch = realFetch;
  session.set(null);
  navigate("#/");
  vi.restoreAllMocks();
});

describe("the getting-started guide", () => {
  it("renders its sections and links the pages it names", async () => {
    const { container } = render(Guide as never, {} as never);
    await settle();
    const text = container.textContent!.replace(/\s+/g, " ");
    for (const s of ["What it is", "Getting in", "Playing a game", "At the table", "Decks"]) {
      expect(text).toContain(s);
    }
    expect(text).toContain("Elder Dragon Highlander");
    expect(text).toContain("Create a voice room");
    const hrefs = [...container.querySelectorAll("a")].map((a) => a.getAttribute("href"));
    for (const h of ["#/practice", "#/decks", "#/roadmap", "#/about"]) expect(hrefs).toContain(h);
  });

  it("shows the Discord section and buttons only when the server names an invite", async () => {
    _setCommunityForTests({ invite_url: "https://discord.gg/example", gate: true });
    const shown = render(Guide as never, {} as never);
    await settle();
    const join = shown.container.querySelector('[data-testid="guide-discord"] a.btn.discord')!;
    expect(join.getAttribute("href")).toBe("https://discord.gg/example");
    const section = shown.container.querySelector('[data-testid="guide-discord-section"]')!;
    expect(section.textContent).toContain("#looking-for-game");
    expect(section.textContent).toContain("/c2-deck-check");
    cleanup();

    _resetCommunityForTests();
    const hidden = render(Guide as never, {} as never);
    await settle();
    expect(hidden.container.querySelector('[data-testid="guide-discord"]')).toBeNull();
    expect(hidden.container.querySelector('[data-testid="guide-discord-section"]')).toBeNull();
  });

  it("is a public route reachable signed out", async () => {
    session.set(null);
    navigate(GUIDE_HASH);
    await settle();
    const { container } = render(App as never, {} as never);
    await settle();
    expect(get(route).name).toBe("guide");
    expect(container.textContent).toContain("Getting started");
  });

  it("is linked from the footer, the front page and Help", async () => {
    const footer = render(SiteFooter as never, {} as never);
    expect(footer.container.querySelector(`a[href="${GUIDE_HASH}"]`)).not.toBeNull();
    cleanup();
    const front = render(FrontPage as never, {} as never);
    expect(front.container.querySelector(`a[href="${GUIDE_HASH}"]`)).not.toBeNull();
    cleanup();
    const help = render(HelpMenu as never, {} as never);
    click(help.container.querySelector<HTMLButtonElement>("#help-button")!);
    await settle();
    const item = help.container.querySelector<HTMLAnchorElement>(
      `#help-menu a[href="${GUIDE_HASH}"]`,
    )!;
    expect(item.textContent?.trim()).toBe("Getting started");
    click(item);
    await settle();
    expect(get(route).name).toBe("guide");
  });
});
