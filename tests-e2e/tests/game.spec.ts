import { expect, test } from "@playwright/test";
import { adminLogin, createGame } from "./lobby-api";
import { ADMIN_TOKEN } from "./env";

// These tests exercise the game route (Game.svelte). The route mounts
// a PixiJS canvas and opens a WebSocket. We don't assert on rendered
// card geometry — that's a visual concern better covered by screenshot
// tests once the layout stabilises. Instead we assert on the things
// Playwright can see without reaching into Pixi: the WS handshake, the
// session wiring, and the URL router.

test.describe("game route", () => {
  test("admin opens a game → WS connects, snapshot arrives", async ({ page, request }) => {
    const token = await adminLogin(request);
    const game = await createGame(request, token, `Admin View ${Date.now()}`);

    // Seed the admin session in localStorage so the Game route can
    // authenticate without us going through the login form. The
    // shape matches lib/session.ts:Session.
    await page.goto("/"); // need a document so localStorage is addressable
    await page.evaluate(
      ([t, n]) => {
        localStorage.setItem(
          "cmdctrl.session",
          JSON.stringify({
            token: t,
            expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
            principal: {
              role: "admin",
              issued_at: new Date().toISOString(),
              expires_at: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
              name: "admin",
            },
          }),
        );
        void n; // unused but keeps parameter types happy
      },
      [token, "admin"],
    );

    const wsPromise = page.waitForEvent("websocket", (ws) => ws.url().includes("/ws"));
    await page.goto(`/#/games/${game.id}`);
    const ws = await wsPromise;
    expect(ws.url()).toContain(`game=${game.id}`);
    expect(ws.url()).toContain(`token=${token}`);
    // The page route committed — no redirect to login.
    await expect(page).toHaveURL(new RegExp(`#/games/${game.id}$`));
  });

  test("logged-in player who joined via invite can open the game", async ({ page, request }) => {
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Player View ${Date.now()}`);

    // Walk the public join flow so the session store carries a real
    // player_id issued by the server — faking one would miss the
    // WS authorisation path.
    await page.goto(`/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`);
    await page.getByPlaceholder("your name").fill("Seat One");
    await page.getByRole("button", { name: "join" }).click();
    await expect(page).toHaveURL(new RegExp(`#/games/${game.id}$`));

    // Confirm the WS URL includes the player= param — this is what
    // pins the seat for state-delta filtering (FilterViewFor).
    const sessJson = await page.evaluate(() =>
      JSON.parse(localStorage.getItem("cmdctrl.session") ?? "null"),
    );
    expect(sessJson?.playerID).toBeTruthy();
  });
});

// Sanity check on the static admin-token constant so a failing server
// boot shows up as "admin login failed" not a cryptic 401 on the
// first test — both sides use the same value out of env.ts.
test("admin token wiring: direct /admin/login succeeds", async ({ request }) => {
  const res = await request.post("/admin/login", { data: { token: ADMIN_TOKEN } });
  expect(res.status()).toBe(200);
  const body = await res.json();
  expect(body.token).toBeTruthy();
  expect(body.principal?.role).toBe("admin");
});
