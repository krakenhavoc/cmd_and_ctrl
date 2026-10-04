// @vitest-environment jsdom
//
// #2191: the unofficial-fan-content notice renders on every page outside
// the table, with its Scryfall and policy links, and never on the game
// or practice table.

import { afterEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import App from "../App.svelte";
import SiteFooter from "./components/SiteFooter.svelte";
import { navigate } from "./router";
import { session } from "./session";
import { cleanup, flushSync, render } from "./test/render.svelte";

vi.mock("../routes/Game.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));
vi.mock("../routes/Practice.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

afterEach(() => {
  cleanup();
  session.set(null);
  navigate("#/home");
});

describe("SiteFooter", () => {
  it("renders the notice, the Scryfall link and the policy link", () => {
    const { container } = render(SiteFooter as never, {} as never);
    const footer = container.querySelector("footer");
    expect(footer).not.toBeNull();
    expect(footer!.textContent).toContain(
      "Unofficial fan content, not approved or endorsed by Wizards of the Coast.",
    );
    expect(footer!.textContent).not.toContain("permitted under the Fan Content Policy");
    const hrefs = [...footer!.querySelectorAll("a")].map((a) => a.getAttribute("href"));
    expect(hrefs).toContain("https://scryfall.com");
    expect(hrefs).toContain("https://company.wizards.com/en/legal/fancontentpolicy");
    expect(footer!.querySelector("details summary")?.getAttribute("aria-label")).toBe(
      "Legal notice",
    );
  });
});

describe("the footer in the app shell", () => {
  async function footerOn(hash: string): Promise<boolean> {
    // A signed-in player, so no route is redirected to the login page.
    session.set({
      token: "tok",
      expiresAt: new Date(Date.now() + 3_600_000).toISOString(),
      principal: { role: "player", user_id: "u1", name: "Owner", game_id: "g1", player_id: "p1" },
    } as never);
    navigate(hash);
    // hashchange is delivered asynchronously.
    await new Promise((r) => setTimeout(r, 0));
    const { container } = render(App as never, {} as never);
    flushSync();
    await tick();
    return container.querySelector("footer.site-footer") !== null;
  }

  it("shows on a non-game route", async () => {
    expect(await footerOn("#/roadmap")).toBe(true);
  });

  it("is absent on the game table", async () => {
    expect(await footerOn("#/games/g1")).toBe(false);
  });

  it("shows on the practice door page, which is not the table (the table is the game route)", async () => {
    expect(await footerOn("#/practice")).toBe(true);
  });
});
