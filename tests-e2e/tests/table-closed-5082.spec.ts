import {
  expect,
  test,
  type APIRequestContext,
  type Page,
} from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { adminLogin, createGame } from "./lobby-api";

// #5082, ADR 0211 §4 "Closed means closed": an archived table refuses
// every way in for anyone but an admin (#5042). Over HTTP the refusal is
// a 409 with code "game_closed", and the client says "This table was
// closed." for it. A refused WebSocket upgrade reaches the browser as a
// bare 1006, so after a failed dial the client asks GET /games/{id}
// whether the table was closed, and stops redialling when it was.

async function archive(
  request: APIRequestContext,
  adminToken: string,
  gameID: string,
) {
  const res = await request.post(`/games/${gameID}/archive`, {
    headers: { Authorization: `Bearer ${adminToken}` },
  });
  expect(res.ok(), `archive failed: ${res.status()} ${await res.text()}`).toBe(
    true,
  );
}

// countDials counts the table sockets the page opens from now on.
function countDials(page: Page): () => number {
  let dials = 0;
  page.on("websocket", (ws) => {
    if (ws.url().includes("/ws")) dials += 1;
  });
  return () => dials;
}

test.describe("#5082 a closed table says so", () => {
  test("join and spectate links to a closed table say it was closed", async ({
    page,
    request,
  }) => {
    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Closed Join ${Date.now()}`,
    );
    await archive(request, adminToken, game.id);

    await page.goto(
      `/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`,
    );
    await page.getByPlaceholder("your name").fill("Late Player");
    await page.getByRole("button", { name: "join" }).click();
    await expect(page.getByRole("alert")).toContainText(L.gameClosed);
    await expect(page).toHaveURL(/\/join\?t=/);

    await page.goto(
      `/#/games/${game.id}/join?t=${encodeURIComponent(game.spectator_invite!)}&spectator=1`,
    );
    await page.getByPlaceholder("your name (chat label)").fill("Late Watcher");
    await page.getByRole("button", { name: "watch" }).click();
    await expect(page.getByRole("alert")).toContainText(L.gameClosed);
    await expect(page).toHaveURL(/\/join\?t=/);
  });

  test("a seat at a table that is closed under it stops and says so", async ({
    page,
    request,
  }) => {
    // The cold-start allowance below needs more than the default 30s.
    test.slow();
    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Closed Seat ${Date.now()}`,
    );

    await page.goto(
      `/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`,
    );
    await page.getByPlaceholder("your name").fill("Seated Player");
    await page.getByRole("button", { name: "join" }).click();
    await expect(page).toHaveURL(/#\/lobby$/);
    await page.goto(`/#/games/${game.id}`);
    // The game route is the client's biggest module; a cold dev server
    // can take a while to serve it the first time.
    await expect(page.locator(".status-connected")).toBeVisible({
      timeout: 60_000,
    });

    // The archive closes the table's sockets with 1000, as a delete does;
    // the client asks which it was.
    await archive(request, adminToken, game.id);
    await expect(page.getByText(L.gameClosed, { exact: true })).toBeVisible();
    await expect(page.getByRole("link", { name: "Tables" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Reconnect" })).toHaveCount(
      0,
    );
  });

  test("a spectator who reopens a closed table is not redialled forever", async ({
    page,
    request,
  }) => {
    test.slow();
    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Closed Watch ${Date.now()}`,
    );

    await page.goto(
      `/#/games/${game.id}/join?t=${encodeURIComponent(game.spectator_invite!)}&spectator=1`,
    );
    await page.getByPlaceholder("your name (chat label)").fill("Watcher");
    await page.getByRole("button", { name: "watch" }).click();
    await expect(page).toHaveURL(new RegExp(`#/games/${game.id}$`));
    await expect(page.locator(".status-connected")).toBeVisible({
      timeout: 60_000,
    });

    await archive(request, adminToken, game.id);
    await expect(page.getByText(L.gameClosed, { exact: true })).toBeVisible();

    // Reopened, the spectator's upgrade is refused with 409, which the
    // browser reports as a bare 1006: the path the probe exists for.
    const dials = countDials(page);
    await page.reload();
    await expect(page.getByText(L.gameClosed, { exact: true })).toBeVisible();
    await expect(page.locator(".status-table_closed")).toBeVisible();
    const seen = dials();
    expect(seen).toBeGreaterThanOrEqual(1);
    // The ladder's first rungs are under two seconds apart; four seconds
    // of silence means it stopped.
    await page.waitForTimeout(4_000);
    expect(dials()).toBe(seen);
  });
});
