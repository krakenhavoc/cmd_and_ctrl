import { expect, test } from "@playwright/test";
import { ADMIN_TOKEN } from "./env";
import { adminLogin, createGame } from "./lobby-api";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";

// ADR 0124 §8's nightly spec: the admin views, end to end. The token's
// session opens Live now, sees the table this spec created with its
// seat connected, opens Games, and finds the same table there.
//
// It selects on the four labels §7 adds to the labels contract
// (AGENTS.md §5): `navigation "admin views"` with its links Live now,
// Games and Accounts, `region "live now"`, `table "games"` and
// `table "accounts"`.
test.describe("admin views", () => {
  test("Live now shows a connected seat, and Games lists the same table", async ({
    browser,
    page,
    request,
  }) => {
    test.slow();

    const name = `Admin views ${Date.now()}`;
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, name);
    expect(game.invite_token).toBeTruthy();

    let alice: JoinedPlayer | null = null;
    try {
      // A guest seat with an open socket: the player sits on the table.
      alice = await joinAsPlayer(browser, game.id, game.invite_token!, "Alice");

      // The token signs in on #/admin; #/admin again now goes to Live now.
      await page.goto("/");
      await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
      await page.goto("/#/admin");
      await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
      await page.getByRole("button", { name: "log in" }).click();
      await expect(page).toHaveURL(/#\/lobby$/);

      // The header's Admin link, an admin's only, goes to Live now.
      await page
        .getByRole("navigation", { name: "Site" })
        .getByRole("link", { name: "Admin", exact: true })
        .click();
      await expect(page).toHaveURL(/#\/admin\/live$/);

      const tabs = page.getByRole("navigation", { name: "admin views" });
      await expect(tabs.getByRole("link", { name: "Live now" })).toHaveAttribute(
        "aria-current",
        "page",
      );
      await expect(tabs.getByRole("link", { name: "Games" })).toBeVisible();
      await expect(tabs.getByRole("link", { name: "Accounts" })).toBeVisible();

      // The table is listed with Alice's seat connected. Live now asks
      // again every 10 seconds, so a socket that opened after the first
      // ask shows up on the next.
      const live = page.getByRole("region", { name: "live now" });
      const card = live.getByRole("listitem").filter({ hasText: name }).first();
      await expect(card).toBeVisible({ timeout: 25_000 });
      const seat = card.getByRole("listitem").filter({ hasText: "Alice" });
      await expect(seat).toContainText(/connected\s*·\s*since/, { timeout: 25_000 });
      await expect(seat).not.toContainText("not connected");

      // Games lists the same table, newest first.
      await tabs.getByRole("link", { name: "Games" }).click();
      await expect(page).toHaveURL(/#\/admin\/games$/);
      const games = page.getByRole("table", { name: "games" });
      const row = games.getByRole("row").filter({ hasText: name });
      await expect(row).toBeVisible();
      await expect(row).toContainText("Alice");

      // The row opens the table's detail.
      await row.getByRole("link", { name }).click();
      await expect(page).toHaveURL(new RegExp(`#/admin/games/${game.id}$`));
      await expect(page.getByRole("link", { name: /Open table/ })).toBeVisible();

      // Accounts renders its table (the token's stack may have none in it).
      await tabs.getByRole("link", { name: "Accounts" }).click();
      await expect(page).toHaveURL(/#\/admin\/accounts$/);
      await expect(
        page.getByRole("table", { name: "accounts" }).or(page.getByText("No accounts match.")),
      ).toBeVisible();
    } finally {
      await closeAll(alice);
    }
  });

  test("a signed-out visitor is sent to sign in, and the token's form sends an admin back", async ({
    page,
  }) => {
    await page.goto("/");
    await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
    await page.goto("/#/admin/games?state=active&archived=false");
    await expect(page).toHaveURL(/#\/login$/);

    // The token's sign-in comes back to the view the link named.
    await page.goto("/#/admin");
    await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
    await page.getByRole("button", { name: "log in" }).click();
    await expect(page).toHaveURL(/#\/admin\/games\?state=active&archived=false$/);
    await expect(
      page.getByRole("table", { name: "games" }).or(page.getByText("No tables match.")),
    ).toBeVisible();
  });
});
