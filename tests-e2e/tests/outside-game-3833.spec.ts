import { expect, test } from "@playwright/test";
import { buildDeck } from "./s19-deck-fixture";
import { adminMoveByName, setupS19Game, type S19Setup } from "./s19-helpers";

// #3833, #4054 (ADR 0191, amended 2026-10-19): a deck's sideboard is not on
// the table. It seats outside the game, as before, but the player rail has
// no sideboard pile and the zone browser no sideboard tab, for the owner or
// anyone else. Wish's "name a card to play" lives in the action dock, and
// only while the right is live (render-tested; it needs a cast Wish).
//
// #4109 rides along: hovering a card with an Adventure lists both halves'
// rules text, and asks for no second scan (an adventurer has one image).

const SIDEBOARD = "\nSideboard:\n1 Burning Wish\n1 Living Wish\n";

test.describe("the sideboard is off the table", () => {
  test.describe.configure({ mode: "serial" });
  let setup: S19Setup | null = null;

  test.afterEach(async () => {
    if (setup) {
      await setup.shutdown();
      setup = null;
    }
  });

  test("no pile in the rail and no tab in the zone browser, though the deck seats a sideboard", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request, {
      casterDeck: buildDeck(["Bonecrusher Giant", "Lightning Bolt"], "Forest", 97) + SIDEBOARD,
    });
    const { admin, caster, opponent } = setup;

    // The deck did seat its sideboard: the server counts it.
    const seat = admin.snapshot().seats.find((s) => s.id === caster.playerID) as unknown as {
      sideboard_count?: number;
    };
    expect(seat.sideboard_count).toBe(2);

    for (const who of [caster, opponent]) {
      await expect(who.page.getByRole("region", { name: "your board" })).toBeVisible();
      await expect(who.page.locator('[data-pile="sideboard"]')).toHaveCount(0);
      await expect(who.page.getByRole("button", { name: /^sideboard/i })).toHaveCount(0);
    }

    // The graveyard browser offers its three zones and no more.
    const board = caster.page.getByRole("region", { name: "your board" });
    await board.getByRole("button", { name: /^grave: / }).click();
    const dialog = caster.page.getByRole("dialog", { name: /graveyard/i });
    await expect(dialog).toBeVisible();
    await expect(dialog.getByRole("tab")).toHaveText(["graveyard", "exile", "command"]);
    await expect(dialog).not.toContainText(/sideboard/i);
    await caster.page.keyboard.press("Escape");

    // No right is live, so the dock has no outside-the-game button either.
    await expect(caster.page.getByTestId("name-outside-card")).toHaveCount(0);
    await expect(caster.page.getByTestId("open-outside-cards")).toHaveCount(0);
  });

  test("hover zoom of an adventurer lists the creature and its Adventure", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request, {
      casterDeck: buildDeck(["Bonecrusher Giant", "Lightning Bolt"], "Forest", 97),
    });
    const { admin, caster } = setup;
    const imageRequests: string[] = [];
    caster.page.on("request", (r) => imageRequests.push(r.url()));

    await adminMoveByName(admin, caster.playerID, "Bonecrusher Giant", "library", "battlefield");
    const card = caster.page
      .getByRole("region", { name: "your board" })
      .getByRole("img", { name: "Bonecrusher Giant", exact: true })
      .first();
    await expect(card).toBeVisible();
    await card.hover();

    const other = caster.page.locator(".overlay .other-side");
    await expect(other).toBeVisible();
    await expect(other.locator("h3")).toHaveText("Adventure");
    await expect(other).toContainText("Stomp");
    await expect(other).toContainText("Damage can't be prevented this turn");
    // The creature's own text is in the usual place.
    await expect(caster.page.locator(".overlay .info > .oracle")).toContainText(
      "becomes the target of a spell",
    );
    // One image: no inset, and no ?face=1 scan was ever requested.
    await expect(caster.page.locator(".overlay .other-face")).toHaveCount(0);
    expect(imageRequests.filter((u) => /\/image\?.*face=1/.test(u))).toEqual([]);
  });
});
