import { expect, test } from "@playwright/test";
import { adminLogin, createGame, uploadDeckAs, startGameAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { joinAsPlayer } from "./players";

// board-layout: smoke-tests the new HTML/CSS player panel that replaces
// the PixiJS canvas board. Asserts the wireframe is wired up — each
// player panel exposes its CREATURES / LANDS / right-column zones,
// the four pile buttons (EXILE / GRAVE / DECK / CMD), and a HAND strip.
// Then drives the DECK pile to confirm draw_card flows end-to-end and
// that the resulting card lands in the hand strip with a real image.
//
// Prereqs match full-game.spec.ts: server + Scryfall bulk dump.

test.describe("board layout", () => {
  test("self panel renders all wireframe zones; deck pile draws a card", async ({
    browser,
    request,
  }) => {
    test.slow();

    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Board ${Date.now()}`);
    expect(game.invite_token).toBeTruthy();

    const alice = await joinAsPlayer(browser, game.id, game.invite_token!, "Alice");
    const bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");

    const deck = makeCommanderDeck();
    await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
    await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
    await startGameAs(request, adminToken, game.id);

    // Both players keep their opening hand so the table view is the
    // primary UI on screen.
    for (const p of [alice, bob]) {
      await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toBeVisible({
        timeout: 10_000,
      });
      await p.page.getByRole("button", { name: "Keep hand" }).click();
      await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toHaveCount(0, {
        timeout: 10_000,
      });
    }

    // Self panel: every wireframe zone is present. The aria-label on
    // each BattlefieldRow / Hand / PileBar / PlayerHeader is the
    // contract used here.
    const page = alice.page;
    // Scope to the self panel: every seat renders the same row labels,
    // so an unscoped query matches the opponent's rows too.
    const selfBoard = page.getByRole("region", { name: "your board" });
    // toBeAttached, not toBeVisible: an empty row is a zero-height
    // flex container (Playwright: "hidden") until a card lands in it.
    await expect(selfBoard.getByRole("list", { name: "creatures" })).toBeAttached();
    await expect(selfBoard.getByRole("list", { name: "lands" })).toBeAttached();
    await expect(selfBoard.getByRole("list", { name: "enchant / artifact" })).toBeAttached();
    await expect(page.getByLabel("your hand")).toBeVisible();
    await expect(page.getByLabel("Alice piles")).toBeVisible();

    // Pile zones: EXILE / GRAVE / LIBRARY as PileButtons, plus the
    // command zone (its own affordance since the CMD pile button was
    // replaced by the cast UI). Scoped to the self board — the
    // opponent's panel renders the same piles.
    await expect(selfBoard.getByRole("button", { name: /exile: \d+/ })).toBeVisible();
    await expect(selfBoard.getByRole("button", { name: /grave: \d+/ })).toBeVisible();
    await expect(selfBoard.getByRole("button", { name: /library: \d+/ })).toBeVisible();
    await expect(selfBoard.getByLabel(/command zone, \d+ card/)).toBeVisible();

    // ADR 0111 §1 (PR 2): the action dock, bottom right. Its label is
    // the tutorial's step-8 anchor (ADR 0076 §2.4) and an e2e contract.
    // It holds the turn header, the priority toggles and the action bar
    // with next and Pass turn, each drawn once on the page.
    const dock = page.getByRole("region", { name: "actions", exact: true });
    await expect(dock).toBeVisible();
    await expect(dock.getByLabel("turn and phase indicator")).toBeVisible();
    await expect(dock.getByRole("group", { name: "priority controls" })).toBeVisible();
    await expect(dock.getByRole("button", { name: "next", exact: true })).toBeVisible();
    await expect(dock.getByRole("button", { name: "Pass turn" })).toBeVisible();
    await expect(page.getByRole("button", { name: "next", exact: true })).toHaveCount(1);
    await expect(page.getByRole("button", { name: "pass turn" })).toHaveCount(1);
    // Owner decision 1: the dock is in the screen's bottom-right corner,
    // and the self panel's rail (its piles) ends above it.
    const box = (await dock.boundingBox())!;
    const viewport = page.viewportSize()!;
    expect(viewport.width - (box.x + box.width)).toBeLessThan(40);
    expect(viewport.height - (box.y + box.height)).toBeLessThan(40);
    // The rail is the scroll container the piles live in; on a short
    // panel the piles scroll inside it rather than slide under the dock.
    const rail = (await selfBoard.locator(".rail").boundingBox())!;
    expect(rail.y + rail.height).toBeLessThanOrEqual(box.y + 1);

    // ADR 0111 §5: the bluff split button is always on the table, even
    // on a default install where neither bluff setting is on.
    await expect(dock.getByRole("button", { name: "bluff", exact: true })).toBeVisible();
    await expect(dock.getByRole("button", { name: "bluff options" })).toBeVisible();

    // Stack overlay should be hidden when nothing is on the stack.
    await expect(page.getByLabel(/stack: \d+ on the stack/)).toHaveCount(0);

    // Drawing a card via the deck pile button bumps the hand size by 1.
    // Read the count from the hand zone's child elements (one card per
    // entry); the toolbar's "draw" button works too but we want to
    // exercise the pile-button click path specifically.
    const handCardsBefore = await page.getByLabel("your hand").locator(".hand-slot").count();
    await selfBoard.getByRole("button", { name: /library: \d+/ }).click();
    await expect(page.getByLabel("your hand").locator(".hand-slot")).toHaveCount(
      handCardsBefore + 1,
      { timeout: 5_000 },
    );

    await alice.context.close();
    await bob.context.close();
  });
});
