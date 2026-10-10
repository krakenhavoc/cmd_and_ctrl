import { expect, test, type Page } from "@playwright/test";
import { adminLogin, createGame, uploadDeckAs, startGameAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { closeAll, joinAsPlayer } from "./players";

// #2919: when a game ends, every player stays on the ended table, with
// the final board and the game-over prompt, until they press Back to
// lobby themselves. That holds for a player eliminated earlier, for the
// one whose concession ended it, and for the winner, and it holds again
// after a page reload.
//
// Prereqs match board-layout.spec.ts: server + Scryfall bulk dump.

async function concede(page: Page): Promise<void> {
  const dock = page.getByRole("region", { name: "actions", exact: true });
  await dock.getByRole("button", { name: "more actions" }).click();
  await dock.getByRole("menuitem", { name: "Concede…" }).click();
  const confirm = page.getByRole("dialog", { name: "concede the game?" });
  await confirm.getByRole("button", { name: "Concede" }).click();
  await expect(confirm).toHaveCount(0);
}

test.describe("an ended table", () => {
  test("keeps every player on it until they leave, across a reload", async ({
    browser,
    request,
  }) => {
    test.slow();

    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Game end ${Date.now()}`,
    );
    const alice = await joinAsPlayer(
      browser,
      game.id,
      game.invite_token!,
      "Alice",
    );
    const bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");
    const carol = await joinAsPlayer(
      browser,
      game.id,
      game.invite_token!,
      "Carol",
    );
    const players = [alice, bob, carol];
    const deck = makeCommanderDeck();
    for (const p of players) {
      await uploadDeckAs(request, adminToken, game.id, p.playerID, deck);
    }
    await startGameAs(request, adminToken, game.id);

    await Promise.all(
      players.map(async (p) => {
        await expect(
          p.page.getByRole("dialog", { name: /keep or mulligan/i }),
        ).toBeVisible({
          timeout: 10_000,
        });
        await p.page.getByRole("button", { name: "Keep hand" }).click();
        await expect(
          p.page.getByRole("dialog", { name: /keep or mulligan/i }),
        ).toHaveCount(0, {
          timeout: 10_000,
        });
      }),
    );

    // Alice leaves first; the game goes on. Then Carol's concession
    // ends it, and Bob wins.
    await concede(alice.page);
    await expect(
      alice.page.getByText("You have been eliminated. Spectating."),
    ).toBeVisible();
    await concede(carol.page);

    const onTable = new RegExp(`#/games/${game.id}$`);
    const gameOver = (page: Page) =>
      page
        .getByRole("region", { name: "actions", exact: true })
        .getByRole("dialog", { name: "game over" });

    for (const p of players) {
      await expect(gameOver(p.page)).toBeVisible({ timeout: 10_000 });
    }
    // Nobody is moved off the table, however long they look at it.
    await players[0].page.waitForTimeout(3_000);
    for (const p of players) {
      await expect(p.page).toHaveURL(onTable);
      await expect(gameOver(p.page)).toBeVisible();
    }

    // A reload lands back on the ended table.
    for (const p of players) {
      await p.page.reload();
      await expect(p.page).toHaveURL(onTable);
      await expect(gameOver(p.page)).toBeVisible({ timeout: 10_000 });
    }

    // Leaving is the player's own choice, and it works, even while the
    // game-end fanfare is showing over the table (#2934).
    await expect(bob.page.getByTestId("game-end-fanfare")).toBeVisible();
    await gameOver(bob.page)
      .getByRole("button", { name: "Back to lobby" })
      .click();
    await expect(bob.page).toHaveURL(/#\/lobby$/);
    for (const p of [alice, carol]) {
      await expect(p.page).toHaveURL(onTable);
    }

    await closeAll(...players);
  });
});
