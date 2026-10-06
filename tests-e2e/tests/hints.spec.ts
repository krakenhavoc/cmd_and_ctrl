import {
  expect,
  test,
  type BrowserContext,
  type Locator,
  type Page,
} from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { ADMIN_TOKEN } from "./env";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";

// hints: the first-use hints on the real pages (ADR 0125 §8, #1085).
//
// Site pages, in one fresh browser context per kind of session: each
// page's hint appears beside its anchor, "Got it" dismisses it, a
// reload does not bring it back, and Help → Tips for this page shows it
// again. The Settings dialog has no Help button over it, so its hint
// comes back through Settings → Advanced → Show all tips again.
//
// The table: two guests at a table started through startGameAs. No
// hint is on screen while the mulligan is open, table.dock shows in
// the first quiet moment, and `i` (Go to the tip) reaches it.
//
// Every page fails the spec on a console line containing "has no
// anchor": a hint whose anchor the page does not render.
//
// Not covered, because the harness cannot sign anyone in with Discord:
// lobby.create and decks.library (a signed-in person only). The other
// table hints wait on game states (a stack item, a second turn, three
// seats) that this spec does not build; the tutorial walk and the unit
// tests cover the anchors they share with the tutorial.
//
// Labels come from the contract-label registry (client/src/lib/labels.ts),
// so a rename follows on its own.

/** How long a site page waits for its hint's anchor (lib/hints/queue.ts SITE_WINDOW_MS). */
const SITE_WINDOW_MS = 3_000;
/** "Tips show once": the site.help hint, offered on the first page with no hint of its own. */
const SITE_HELP = "Tips show once";

/** Collect "has no anchor" console lines from every page of a context. */
function watchAnchors(context: BrowserContext): string[] {
  const lines: string[] = [];
  context.on("console", (msg) => {
    const text = msg.text();
    if (text.includes("has no anchor")) lines.push(text);
  });
  return lines;
}

function tipOf(page: Page): Locator {
  return page.getByRole("complementary", { name: L.tip, exact: true });
}

/**
 * The tip says `title`, and sits beside `anchor`: the anchor is
 * described by the tip's body, and the card is next to it without
 * covering it.
 */
async function expectTipBeside(
  page: Page,
  title: string,
  anchor: Locator,
): Promise<void> {
  const tip = tipOf(page);
  // A tip is drawn only where it fits on screen beside its anchor, so
  // one below the fold waits until the person scrolls to it, as they
  // would to use the field.
  await anchor.scrollIntoViewIfNeeded();
  await expect(tip).toBeVisible();
  await expect(tip).toContainText(title);
  await expect(anchor).toHaveAttribute("aria-describedby", /hint-tip-body/);
  // The card follows its anchor on the layer's poll, so an anchor that
  // is still settling (a sheet folding away) is measured again.
  await expect(async () => {
    const [a, t] = await Promise.all([anchor.boundingBox(), tip.boundingBox()]);
    expect(a, "the anchor has a box").not.toBeNull();
    expect(t, "the tip has a box").not.toBeNull();
    const dx = Math.max(0, a!.x - (t!.x + t!.width), t!.x - (a!.x + a!.width));
    const dy = Math.max(
      0,
      a!.y - (t!.y + t!.height),
      t!.y - (a!.y + a!.height),
    );
    const overlaps =
      a!.x < t!.x + t!.width &&
      t!.x < a!.x + a!.width &&
      a!.y < t!.y + t!.height &&
      t!.y < a!.y + a!.height;
    expect(overlaps, `the "${title}" tip covers its anchor`).toBe(false);
    expect(
      Math.hypot(dx, dy),
      `the "${title}" tip is beside its anchor`,
    ).toBeLessThanOrEqual(24);
  }).toPass({ timeout: 5_000 });
}

/**
 * Dismiss the tip on screen with its primary button. A replay may put
 * the next tip up at once, so what goes is this tip's title.
 */
async function dismiss(
  page: Page,
  button: "Got it" | "Not now" = "Got it",
): Promise<void> {
  const tip = tipOf(page);
  const title = (await tip.locator(".hint-title").textContent()) ?? "";
  await tip.getByRole("button", { name: button, exact: true }).click();
  await expect(tip.filter({ hasText: title })).toHaveCount(0);
}

/**
 * Nothing shows on this visit: run the page's clock past the time a
 * site hint's anchor may take to appear (the context's clock is
 * installed, so this waits on nothing real), then look.
 */
async function expectNoTipThisVisit(
  page: Page,
  anchor: Locator,
): Promise<void> {
  await expect(anchor).toBeVisible();
  await anchor.scrollIntoViewIfNeeded();
  await page.clock.runFor(SITE_WINDOW_MS + 500);
  await expect(tipOf(page)).toHaveCount(0);
}

/** Help → Tips for this page. */
async function tipsForThisPage(page: Page): Promise<void> {
  await page.getByRole("button", { name: L.help, exact: true }).click();
  await page
    .getByRole("menuitem", { name: L.tipsForThisPage, exact: true })
    .click();
}

/**
 * One site page, the whole cycle: its hint beside its anchor; dismissed,
 * and not back after a reload; back from Help → Tips for this page,
 * followed by the site's own Help tip, which a replay also shows.
 */
async function pageCycle(
  page: Page,
  opts: {
    url?: string;
    title: string;
    anchor: Locator;
    dismissWith?: "Got it" | "Not now";
  },
): Promise<void> {
  if (opts.url) await page.goto(opts.url);
  await expectTipBeside(page, opts.title, opts.anchor);
  await dismiss(page, opts.dismissWith);

  await page.reload();
  await expectNoTipThisVisit(page, opts.anchor);

  await tipsForThisPage(page);
  await expectTipBeside(page, opts.title, opts.anchor);
  await dismiss(page, opts.dismissWith);
  if (opts.title !== SITE_HELP) {
    // The replay goes on to the site's hint, after the page's own.
    await expectTipBeside(
      page,
      SITE_HELP,
      page.getByRole("button", { name: L.help, exact: true }),
    );
    await dismiss(page);
  }
  await expect(tipOf(page)).toHaveCount(0);
}

test.describe("first-use hints", () => {
  test("site pages: each hint shows once beside its anchor, and Help shows it again", async ({
    browser,
    request,
  }) => {
    test.slow();
    const context = await browser.newContext();
    // Timers and Date run as usual; expectNoTipThisVisit only moves
    // them on, so a negative check needs no real wait.
    await context.clock.install();
    const noAnchor = watchAnchors(context);
    const page = await context.newPage();
    try {
      const help = page.getByRole("button", { name: L.help, exact: true });

      // Home, signed out: no hint of its own, so the site's Help tip.
      await pageCycle(page, { url: "/#/home", title: SITE_HELP, anchor: help });

      // The public pages.
      await pageCycle(page, {
        url: "/#/roadmap",
        title: "What comes next",
        anchor: page.getByRole("textbox", {
          name: L.searchRoadmap,
          exact: true,
        }),
      });
      await pageCycle(page, {
        url: "/#/decks",
        title: "Check a deck",
        anchor: page.getByRole("textbox", { name: L.deckLink, exact: true }),
      });

      // A guest seat, by an invite link, lands on the Lobby: the
      // tutorial offer, whose "Got it" reads "Not now".
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Hints ${Date.now()}`);
      await page.goto(
        `/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`,
      );
      await page.getByPlaceholder("your name").fill("Guest");
      await page.getByRole("button", { name: "join" }).click();
      await expect(page).toHaveURL(/#\/lobby$/);
      const title = page.getByRole("heading", {
        name: L.lobbyTitle,
        exact: true,
      });
      await expect(
        tipOf(page).getByRole("link", { name: L.startPractice, exact: true }),
      ).toHaveAttribute("href", "#/practice");
      await pageCycle(page, {
        title: "New here?",
        anchor: title,
        dismissWith: "Not now",
      });

      // The catalogue needs a session; the guest's will do.
      await pageCycle(page, {
        url: "/#/catalog",
        title: "Every card the game plays",
        anchor: page.getByRole("textbox", {
          name: L.searchCatalogue,
          exact: true,
        }),
      });

      // The Settings dialog is a visit of its own each time it opens.
      await page.goto("/#/lobby");
      await expectNoTipThisVisit(page, title);
      const settings = page.getByRole("dialog", { name: "settings" });
      const openSettings = async () => {
        await page.getByRole("button", { name: /^account menu:/ }).click();
        await page
          .getByRole("button", { name: "Settings", exact: true })
          .click();
        await expect(settings).toBeVisible();
      };
      const closeSettings = async () => {
        await settings.getByRole("button", { name: "close settings" }).click();
        await expect(settings).toHaveCount(0);
        // The hint layer notices the dialog closing on its next poll
        // (100 ms); a person takes longer than that to open it again.
        await page.clock.runFor(500);
      };
      const sections = settings.getByRole("navigation", {
        name: L.settingsSections,
        exact: true,
      });
      await openSettings();
      await expectTipBeside(page, "Skins are under Display", sections);
      await dismiss(page);
      await closeSettings();
      await openSettings();
      await expectNoTipThisVisit(page, sections);
      // Settings → Advanced: Show all tips again, and the next opening
      // offers it once more.
      await sections
        .getByRole("button", { name: "Advanced", exact: true })
        .click();
      await settings
        .getByRole("button", { name: L.showAllTipsAgain, exact: true })
        .click();
      await closeSettings();
      await openSettings();
      await expectTipBeside(page, "Skins are under Display", sections);
      await dismiss(page);
      await closeSettings();

      expect(noAnchor, "a hint pointed at nothing").toEqual([]);
    } finally {
      await context.close();
    }
  });

  test("the admin views: the admin token's hint", async ({ browser }) => {
    test.slow();
    const context = await browser.newContext();
    await context.clock.install();
    const noAnchor = watchAnchors(context);
    const page = await context.newPage();
    try {
      await page.goto("/#/admin");
      await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
      await page.getByRole("button", { name: "log in" }).click();
      await expect(page).toHaveURL(/#\/lobby$/);
      // A fresh context: the Lobby has no hint for the token (no
      // tutorial offer, no create form tip), so the site's Help tip.
      await expectTipBeside(
        page,
        SITE_HELP,
        page.getByRole("button", { name: L.help, exact: true }),
      );
      await dismiss(page);

      await pageCycle(page, {
        url: "/#/admin/live",
        title: "Admin views",
        anchor: page.getByRole("navigation", {
          name: L.adminViews,
          exact: true,
        }),
      });

      expect(noAnchor, "a hint pointed at nothing").toEqual([]);
    } finally {
      await context.close();
    }
  });

  test("the table: none over the mulligan, then the dock's in a quiet moment, reached with i", async ({
    browser,
    request,
  }) => {
    test.slow();
    let alice: JoinedPlayer | null = null;
    let bob: JoinedPlayer | null = null;
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(
        request,
        adminToken,
        `Hints table ${Date.now()}`,
      );
      alice = await joinAsPlayer(browser, game.id, game.invite_token!, "Alice");
      bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");
      const players = [alice, bob];
      const noAnchor = players.map((p) => watchAnchors(p.context));
      // Each table counts how long it has been on screen before a hint
      // may show (5 s). With the clock installed and the table opened
      // again under it, the spec moves that clock on rather than waiting.
      for (const p of players) {
        await p.context.clock.install();
        await p.page.reload();
        await expect(
          p.page.getByRole("region", { name: L.actions, exact: true }),
        ).toBeVisible();
      }

      const deck = makeCommanderDeck();
      await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
      await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
      // The helper finishes the opening roll, so both land in the mulligan.
      await startGameAs(request, adminToken, game.id);

      const mulliganOf = (p: Page) =>
        p
          .getByRole("region", { name: L.actions, exact: true })
          .getByRole("dialog", { name: L.mulligan, exact: true });
      const keepOf = (p: Page) =>
        mulliganOf(p).getByRole("button", { name: L.keepHand, exact: true });
      for (const p of players) await expect(mulliganOf(p.page)).toBeVisible();

      // Decisions go in turn order (CR 103.5), and the first turn was
      // rolled for: find who decides first.
      await expect
        .poll(
          async () =>
            (await keepOf(players[0].page).isEnabled()) ||
            (await keepOf(players[1].page).isEnabled()),
        )
        .toBe(true);
      const [first, second] = (await keepOf(alice.page).isEnabled())
        ? [alice, bob]
        : [bob, alice];

      // Both tables have settled, and neither shows a tip: the opening
      // hand is open on both (the second seat's waits its turn).
      for (const p of players) {
        await p.page.clock.runFor(6_000);
        await expect(mulliganOf(p.page)).toBeVisible();
        await expect(
          tipOf(p.page),
          `a tip over ${p.name}'s mulligan`,
        ).toHaveCount(0);
      }

      // The first seat keeps: its table is quiet, and the dock's tip
      // shows beside the dock. The second seat's hand is still open,
      // and still no tip there.
      await keepOf(first.page).click();
      await expect(mulliganOf(first.page)).toHaveCount(0);
      const dockOf = (p: Page) =>
        p.getByRole("region", { name: L.actions, exact: true });
      await expectTipBeside(first.page, "Your controls", dockOf(first.page));
      await second.page.clock.runFor(1_000);
      await expect(mulliganOf(second.page)).toBeVisible();
      await expect(tipOf(second.page), "a tip over the mulligan").toHaveCount(
        0,
      );

      // The second seat keeps, and its tip shows too.
      await expect(keepOf(second.page)).toBeEnabled();
      await keepOf(second.page).click();
      await expect(mulliganOf(second.page)).toHaveCount(0);
      await expectTipBeside(second.page, "Your controls", dockOf(second.page));

      // A tip never takes focus; Go to the tip (i) moves it there, and
      // Escape dismisses it (as Got it).
      const tip = tipOf(second.page);
      await expect(tip).not.toBeFocused();
      await second.page.keyboard.press("i");
      await expect(tip).toBeFocused();
      await second.page.keyboard.press("Escape");
      await expect(tip).toHaveCount(0);

      // The first seat's goes with Got it, pressed the way a person
      // presses it: held for longer than the layer's poll (#2422, #2372).
      // A held pointer is a gesture at the table, and a gesture makes a
      // tip step aside; a press on the tip itself must not, or the card
      // is gone before the click lands and back on release.
      const firstTip = tipOf(first.page);
      await firstTip
        .getByRole("button", { name: "Got it", exact: true })
        .hover();
      await first.page.mouse.down();
      await first.page.clock.runFor(400);
      await expect(firstTip, "the tip stays up under the press").toBeVisible();
      await first.page.mouse.up();
      await expect(firstTip).toHaveCount(0);
      // And it does not come back, past the table's gap between tips.
      await first.page.clock.runFor(25_000);
      await expect(
        tipOf(first.page).filter({ hasText: "Your controls" }),
        "a dismissed tip came back",
      ).toHaveCount(0);

      noAnchor.forEach((lines, i) =>
        expect(lines, `a hint pointed at nothing (${players[i].name})`).toEqual(
          [],
        ),
      );
    } finally {
      await closeAll(alice, bob);
    }
  });
});
