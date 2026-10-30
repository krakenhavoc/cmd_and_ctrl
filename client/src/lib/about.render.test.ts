// @vitest-environment jsdom
//
// #2981: the About page, rendered. It is public, links the places it
// names, reuses LegalNotice rather than copying it, and any card count it
// shows comes from the server (GET /roadmap), never from the source.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";
import { get } from "svelte/store";

import App from "../App.svelte";
import SiteFooter from "./components/SiteFooter.svelte";
import FrontPage from "../routes/FrontPage.svelte";
import About from "../routes/About.svelte";
import { ABOUT_HASH } from "./about";
import { _resetCommunityForTests, _setCommunityForTests } from "./community";
import { navigate, route } from "./router";
import { session } from "./session";
import { cleanup, flushSync, render } from "./test/render.svelte";

vi.mock("../routes/Game.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));
vi.mock("../routes/Practice.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

const realFetch = globalThis.fetch;

function mockRoadmap(total: number | null) {
  globalThis.fetch = vi.fn(async (url: unknown) => {
    if (String(url).includes("/roadmap") && total !== null) {
      return new Response(
        JSON.stringify({
          counts: {
            implemented: 1,
            partial: 1,
            missing: 1,
            cards: { total, full: 1, caveats: 1, unreviewed: 0 },
          },
          next: [],
          items: [],
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }
    return new Response("{}", { status: total === null ? 500 : 200 });
  }) as never;
}

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

beforeEach(() => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  mockRoadmap(12345);
});

afterEach(() => {
  cleanup();
  _resetCommunityForTests();
  globalThis.fetch = realFetch;
  session.set(null);
  navigate("#/");
  vi.restoreAllMocks();
});

describe("the about page", () => {
  it("renders its sections, links, and reuses the legal notice", async () => {
    const { container } = render(About as never, {} as never);
    await settle();
    const text = container.textContent!.replace(/\s+/g, " ");
    for (const s of ["Free and unofficial", "Known limits", "Feedback", "Credits"]) {
      expect(text).toContain(s);
    }
    expect(text).toContain("no payments");
    expect(text).toContain("resolve it by hand");
    const hrefs = [...container.querySelectorAll("a")].map((a) => a.getAttribute("href"));
    for (const h of ["#/catalog", "#/roadmap", "#/privacy"]) expect(hrefs).toContain(h);
    // The Fan Content wording is LegalNotice's own, not a second copy here.
    expect(text).toContain("Magic: The Gathering is a trademark of Wizards of the Coast LLC");
    expect(text).toContain("Scryfall");
  });

  it("puts the Discord invite on its own line, from the community store, and hides it with none", async () => {
    _setCommunityForTests({ invite_url: "https://discord.gg/example", gate: false });
    const shown = render(About as never, {} as never);
    await settle();
    const block = shown.container.querySelector('[data-testid="about-discord"]')!;
    const a = block.querySelector("a.btn.discord")!;
    expect(a.getAttribute("href")).toBe("https://discord.gg/example");
    expect(a.textContent).toContain("Join the CMD & CTRL Discord");
    // Not in the Feedback list any more.
    expect(shown.container.querySelector("ul a[href='https://discord.gg/example']")).toBeNull();
    cleanup();

    _resetCommunityForTests();
    const hidden = render(About as never, {} as never);
    await settle();
    expect(hidden.container.querySelector('[data-testid="about-discord"]')).toBeNull();
    expect(hidden.container.textContent).not.toContain("Discord");
  });

  it("shows a card count only when the server supplied it", async () => {
    const { container } = render(About as never, {} as never);
    await settle();
    expect(globalThis.fetch).toHaveBeenCalledWith("/roadmap", expect.anything());
    expect(container.querySelector('[data-testid="about-card-count"]')!.textContent).toBe("12345");
    cleanup();

    mockRoadmap(null);
    const failed = render(About as never, {} as never);
    await settle();
    expect(failed.container.querySelector('[data-testid="about-card-count"]')).toBeNull();
    expect(failed.container.textContent).toContain("Free and unofficial");
  });

  it("is a public route reachable signed out", async () => {
    session.set(null);
    navigate(ABOUT_HASH);
    await settle();
    expect(get(route).name).toBe("about");
    const { container } = render(App as never, {} as never);
    await settle();
    expect(get(route).name).toBe("about");
    expect(container.textContent).toContain("Free and unofficial");
  });

  it("is linked from the footer and the front page", () => {
    const footer = render(SiteFooter as never, {} as never);
    expect(footer.container.querySelector(`a[href="${ABOUT_HASH}"]`)).not.toBeNull();
    cleanup();
    const login = render(FrontPage as never, {} as never);
    expect(login.container.querySelector(`a[href="${ABOUT_HASH}"]`)).not.toBeNull();
  });
});
