import { expect, test, type Locator, type Page } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";

// #2229 / ADR 0121 §5: "Roll a die", end to end.
//
// A seat rolls a d20 from the action dock's ⋯ menu before the first
// turn (the mulligan: startGameAs finishes the opening roll as the
// admin, ADR 0121 PR 5), and the other seat flips a coin during
// a turn. Each roll tumbles at the roller's seat on both screens, and
// both game logs say so: "<name> rolled a d20 at the table: N",
// "<name> flipped a coin at the table: heads|tails".
//
// This spec selects the menu items by their names — "Roll a d6",
// "Roll a d20", "Flip a coin" (ADR 0121 §8) — which is what makes them a
// label contract (AGENTS.md §5). Renaming one is a breaking change.

async function openLog(page: Page): Promise<Locator> {
  const log = page.locator('[aria-label="game log"]');
  if (!(await log.isVisible())) {
    await page.getByRole("button", { name: "open game log" }).click();
  }
  await expect(log).toBeVisible();
  return log;
}

async function rollFromMenu(page: Page, name: "Roll a d6" | "Roll a d20" | "Flip a coin") {
  const dock = page.getByRole("region", { name: "actions", exact: true });
  await dock.getByRole("button", { name: "more actions" }).click();
  const menu = dock.getByRole("menu", { name: "game actions" });
  await expect(menu.getByRole("menuitem", { name: "Roll a d6" })).toBeVisible();
  await expect(menu.getByRole("menuitem", { name: "Roll a d20" })).toBeVisible();
  await expect(menu.getByRole("menuitem", { name: "Flip a coin" })).toBeVisible();
  await menu.getByRole("menuitem", { name }).click();
  await expect(menu).toHaveCount(0);
}

test.describe("#2229 Roll a die", () => {
  test("a d20 in the mulligan and a coin mid-turn, from the ⋯ menu, in both logs", async ({
    browser,
    request,
  }) => {
    test.slow();

    let alice: JoinedPlayer | null = null;
    let bob: JoinedPlayer | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Table roll 2229 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");

      alice = await joinAsPlayer(browser, game.id, game.invite_token, "Alice");
      bob = await joinAsPlayer(browser, game.id, game.invite_token, "Bob");
      const deck = makeCommanderDeck();
      await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
      await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
      await startGameAs(request, adminToken, game.id);

      for (const p of [alice, bob]) {
        await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toBeVisible({
          timeout: 10_000,
        });
      }

      // --- before the first turn: Alice rolls a d20 --------------------
      await rollFromMenu(alice.page, "Roll a d20");
      const d20 = /^Alice rolled a d20 at the table: (\d+)$/;
      // The die tumbles at Alice's seat on Bob's screen too.
      await expect(bob.page.locator('.dice-layer [data-dice-key="table:1"]')).toBeAttached({
        timeout: 10_000,
      });
      for (const p of [alice, bob]) {
        const log = await openLog(p.page);
        await expect(log.getByText(d20)).toHaveCount(1, { timeout: 10_000 });
      }
      const aliceLine = await (await openLog(alice.page)).getByText(d20).textContent();
      const result = Number(aliceLine?.match(d20)?.[1]);
      expect(result).toBeGreaterThanOrEqual(1);
      expect(result).toBeLessThanOrEqual(20);
      await expect((await openLog(bob.page)).getByText(d20)).toHaveText(aliceLine!.trim());

      // It is not a game action: the mulligan is still waiting on both.
      // They decide in turn order (CR 103.5), so keep side by side.
      await Promise.all(
        [alice, bob].map(async (p) => {
          await p.page.getByRole("button", { name: "Keep hand" }).click();
          await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toHaveCount(0, {
            timeout: 10_000,
          });
        }),
      );

      // --- during a turn: Bob flips a coin ------------------------------
      await expect(bob.page.getByRole("region", { name: "actions", exact: true })).toBeVisible();
      await rollFromMenu(bob.page, "Flip a coin");
      const coin = /^Bob flipped a coin at the table: (heads|tails)$/;
      await expect(alice.page.locator('.dice-layer [data-dice-key="table:2"]')).toBeAttached({
        timeout: 10_000,
      });
      for (const p of [alice, bob]) {
        const log = await openLog(p.page);
        await expect(log.getByText(coin)).toHaveCount(1, { timeout: 10_000 });
        // Alice's d20 is still there, once.
        await expect(log.getByText(d20)).toHaveCount(1);
      }
    } finally {
      await closeAll(alice, bob);
    }
  });
});
