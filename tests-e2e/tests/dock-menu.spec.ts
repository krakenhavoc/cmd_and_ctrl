import { expect, test } from "@playwright/test";
import { adminLogin, createGame, uploadDeckAs, startGameAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { closeAll, joinAsPlayer } from "./players";

// dock-menu: ADR 0111 Delivery PR 7 (S56, #1958), owner decision 3. The
// ⋯ menu ("more actions") moved from the command bar into the action
// dock, as the last chip on its toggles row, opening upward. It holds
// the sandbox tools, life history, the table, the vote launcher and
// Concede with its confirm. A spectator has no dock and keeps a smaller
// menu on the command bar. The names here are the contract AGENTS.md
// describes ("Labels are a contract").
//
// Prereqs match board-layout.spec.ts: server + Scryfall bulk dump.

test.describe("the action dock's ⋯ menu", () => {
  test("opens from the dock, cancels a concede, and a spectator keeps one on the command bar", async ({
    browser,
    request,
  }) => {
    test.slow();

    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Dock menu ${Date.now()}`,
    );
    expect(game.invite_token).toBeTruthy();
    expect(game.spectator_invite).toBeTruthy();

    const alice = await joinAsPlayer(
      browser,
      game.id,
      game.invite_token!,
      "Alice",
    );
    const bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");
    const deck = makeCommanderDeck();
    await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
    await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
    await startGameAs(request, adminToken, game.id);

    for (const p of [alice, bob]) {
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
    }

    const page = alice.page;
    const dock = page.getByRole("region", { name: "actions", exact: true });
    const toggles = dock.getByRole("group", { name: "priority controls" });
    await expect(dock).toBeVisible();

    // One ⋯ on the page, in the dock's toggles row; none on the command
    // bar any more.
    await expect(
      page.getByRole("button", { name: "more actions" }),
    ).toHaveCount(1);
    await expect(
      page.locator("header.bar").getByRole("button", { name: "more actions" }),
    ).toHaveCount(0);
    const more = toggles.getByRole("button", { name: "more actions" });
    await expect(more).toBeVisible();
    await expect(more).toHaveAttribute("aria-expanded", "false");

    // It opens upward, out of the dock, with the entries a seat has.
    await more.click();
    const menu = dock.getByRole("menu", { name: "game actions" });
    await expect(menu).toBeVisible();
    await expect(more).toHaveAttribute("aria-expanded", "true");
    for (const name of [
      /^Draw a card/,
      "Untap all",
      "Shuffle library",
      "Life history",
      /^Table settings/,
      "Call a vote…",
      "Back to lobby",
      "Concede…",
    ]) {
      await expect(menu.getByRole("menuitem", { name })).toBeVisible();
    }
    await expect(
      menu.getByRole("spinbutton", { name: "mulligan hand size" }),
    ).toBeVisible();
    const menuBox = (await menu.boundingBox())!;
    const moreBox = (await more.boundingBox())!;
    expect(menuBox.y + menuBox.height).toBeLessThanOrEqual(moreBox.y + 1);

    // Escape closes it, and focus goes back to ⋯.
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);
    await expect(more).toBeFocused();

    // Concede asks first, where the menu was; Keep playing backs out and
    // the seat is still in the game.
    await more.click();
    await dock.getByRole("menuitem", { name: "Concede…" }).click();
    const confirm = page.getByRole("dialog", { name: "concede the game?" });
    await expect(confirm).toBeVisible();
    await confirm.getByRole("button", { name: "Keep playing" }).click();
    await expect(confirm).toHaveCount(0);
    await more.click();
    await expect(
      dock.getByRole("menuitem", { name: "Concede…" }),
    ).toBeEnabled();
    // A press outside it closes it (the play area's left gutter, which
    // holds no card and no control).
    await page.mouse.click(6, 200);
    await expect(dock.getByRole("menu", { name: "game actions" })).toHaveCount(
      0,
    );

    // The vote launcher is in the menu now, not the board's corner.
    await expect(page.getByRole("button", { name: "call a vote" })).toHaveCount(
      0,
    );

    // A spectator: no dock, and the menu stays on the command bar with
    // nothing that acts on a seat.
    const specContext = await browser.newContext();
    const spec = await specContext.newPage();
    await spec.goto(
      `/#/games/${game.id}/join?t=${encodeURIComponent(game.spectator_invite!)}&spectator=1`,
    );
    await spec.getByPlaceholder("your name (chat label)").fill("Watcher");
    await spec.getByRole("button", { name: "watch" }).click();
    await expect(spec).toHaveURL(new RegExp(`#/games/${game.id}$`));
    await expect(spec.getByText("spectating")).toBeVisible();
    await expect(
      spec.getByRole("region", { name: "actions", exact: true }),
    ).toHaveCount(0);
    const specMore = spec
      .locator("header.bar")
      .getByRole("button", { name: "more actions" });
    await expect(specMore).toBeVisible();
    await specMore.click();
    const specMenu = spec.getByRole("menu", { name: "game actions" });
    await expect(
      specMenu.getByRole("menuitem", { name: "Life history" }),
    ).toBeVisible();
    await expect(
      specMenu.getByRole("menuitem", { name: "Back to lobby" }),
    ).toBeVisible();
    await expect(
      specMenu.getByRole("menuitem", { name: "Concede…" }),
    ).toHaveCount(0);
    await expect(
      specMenu.getByRole("menuitem", { name: /^Draw a card/ }),
    ).toHaveCount(0);

    await specContext.close();
    await closeAll(alice, bob);
  });
});
