import { expect, test } from "@playwright/test";
import { adminLogin, createGame } from "./lobby-api";
import { ADMIN_TOKEN } from "./env";

// These tests exercise the game route (Game.svelte) at the transport
// layer: the WebSocket handshake, the session wiring, and the URL
// router. Rendering is covered by board-layout / mulligan /
// zone-browser, which assert on the HTML board's roles and labels.

test.describe("game route", () => {
  test("admin opens a game → WS connects, snapshot arrives", async ({ page, request }) => {
    const token = await adminLogin(request);
    const game = await createGame(request, token, `Admin View ${Date.now()}`);

    // Seed the admin session in localStorage so the Game route can
    // authenticate without us going through the login form. The
    // shape matches lib/session.ts:Session. addInitScript runs
    // before any page script on every navigation — seeding after the
    // app boots (goto + evaluate) is too late: the session store
    // hydrates from localStorage at module init, and a hash-only
    // navigation never re-runs it, so the client would connect
    // anonymously and the server would reject the upgrade.
    await page.context().addInitScript((t) => {
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
    }, token);

    // An authenticated upgrade is proven by a snapshot frame actually
    // arriving — an anonymous socket is rejected before any frame.
    // The listener must attach inside the websocket event handler:
    // the server pre-stages the snapshot at upgrade time, so the
    // frame can land in the same protocol batch as socket creation —
    // a waitForEvent("framereceived") attached after an await would
    // miss it forever.
    const firstFrame = new Promise<string>((resolve) => {
      page.on("websocket", (ws) => {
        if (!ws.url().includes("/ws")) return;
        ws.on("framereceived", () => resolve(ws.url()));
      });
    });
    await page.goto(`/#/games/${game.id}`);
    const wsURL = await firstFrame;
    expect(wsURL).toContain(`game=${game.id}`);
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

    // Players land in the lobby first (deck import) — s085 (#43);
    // a seated session can then open the game route directly.
    await expect(page).toHaveURL(/#\/lobby$/);
    await page.goto(`/#/games/${game.id}`);
    await expect(page).toHaveURL(new RegExp(`#/games/${game.id}$`));

    // The seat is pinned by the session's player_id, which the WS
    // authorizer resolves for state-delta filtering (FilterViewFor).
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
