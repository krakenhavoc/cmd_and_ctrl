import { expect, test, type BrowserContext, type Page } from "@playwright/test";
import { adminLogin, createGame, uploadDeckAs, startGameAs, getGameAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";

// End-to-end happy path: two players join, upload decks, start the
// game, keep their opening hands, and each take a turn. This exercises
// every layer — lobby HTTP, invite flow, deck import against the
// Scryfall index, game Start, WebSocket snapshot propagation, mulligan
// window, and the turn cursor.
//
// Prerequisites:
//   - Server running with the dev-default admin token (make server-dev).
//   - Scryfall bulk dump present at <repo>/data/scryfall/default-cards.json
//     so the deck fixture's cards (Kenrith, Plains) resolve.

interface JoinedPlayer {
  context: BrowserContext;
  page: Page;
  name: string;
  token: string;
  playerID: string;
}

// joinAsPlayer walks the public invite-link flow in a fresh browser
// context, entering `name` and capturing the issued session. The
// resulting token + playerID lets the test drive that player's seat
// either via the UI (this page) or the server API (admin helpers).
async function joinAsPlayer(
  browser: import("@playwright/test").Browser,
  gameID: string,
  inviteToken: string,
  name: string,
): Promise<JoinedPlayer> {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto(`/#/games/${gameID}/join?t=${encodeURIComponent(inviteToken)}`);
  await page.getByPlaceholder("your name").fill(name);
  await page.getByRole("button", { name: "join" }).click();
  await expect(page).toHaveURL(new RegExp(`#/games/${gameID}$`), { timeout: 10_000 });

  const session = await page.evaluate(() =>
    JSON.parse(localStorage.getItem("cmdctrl.session") ?? "null"),
  );
  if (!session?.token || !session?.playerID) {
    throw new Error(`${name}: session missing after join`);
  }
  return { context, page, name, token: session.token, playerID: session.playerID };
}

test.describe("full game", () => {
  test("2 players: join, upload decks, start, keep hands, play 2 turns", async ({
    browser,
    request,
  }) => {
    test.slow(); // deck import + WS dance needs more than the 30s default.

    // ---- 1. Admin creates the game and grabs the invite token. ----
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Full Game ${Date.now()}`);
    expect(game.invite_token).toBeTruthy();

    // ---- 2. Two players join via the public invite flow. ----
    const alice = await joinAsPlayer(browser, game.id, game.invite_token!, "Alice");
    const bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");

    // Sanity: admin-side view shows both seats, neither deck uploaded.
    const beforeUpload = await getGameAs(request, adminToken, game.id);
    expect(beforeUpload.players).toHaveLength(2);
    expect(beforeUpload.players.every((p) => !p.deck_uploaded)).toBe(true);
    expect(beforeUpload.state).toBe("lobby");

    // ---- 3. Admin uploads the same deck for each seat. ----
    // Admin role can upload for any player; skipping the lobby UI
    // here keeps the test focused on gameplay (the lobby upload UI
    // has its own coverage in lobby.spec.ts).
    const deck = makeCommanderDeck();
    const up1 = await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
    const up2 = await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
    expect(up1.card_count).toBe(100);
    expect(up2.card_count).toBe(100);
    expect(up1.commanders).toContain("Kenrith, the Returned King");

    const afterUpload = await getGameAs(request, adminToken, game.id);
    expect(afterUpload.players.every((p) => p.deck_uploaded)).toBe(true);

    // ---- 4. Start the game. ----
    const started = await startGameAs(request, adminToken, game.id);
    expect(started.state).toBe("active");

    // ---- 5. Both players' game routes should flip into the
    //         mulligan window. The WS push happens automatically
    //         because both pages were already on /#/games/<id>.
    for (const p of [alice, bob]) {
      await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toBeVisible({
        timeout: 10_000,
      });
      // Each player sees 7 cards in their opening hand. The count
      // also surfaces in the dialog header, so assert on both.
      await expect(p.page.getByText(/Hand size: 7/)).toBeVisible();
    }

    // ---- 6. Both players keep their opening hand. ----
    for (const p of [alice, bob]) {
      await p.page.getByRole("button", { name: "Keep hand" }).click();
    }

    // Once both have kept, the mulligan dialog tears down on every
    // page and the primary toolbar appears.
    for (const p of [alice, bob]) {
      await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toHaveCount(0, {
        timeout: 10_000,
      });
    }

    // ---- 7. Play round 1: Alice (seat 0) passes her turn. ----
    // pass_turn jumps to the next seat's untap step. Only the active
    // seat's button is enabled; the other browser sees it disabled.
    const aliceIsActive = alice.page.getByRole("button", { name: "pass turn" });
    await expect(aliceIsActive).toBeEnabled({ timeout: 10_000 });
    await expect(bob.page.getByRole("button", { name: "pass turn" })).toBeDisabled();

    await aliceIsActive.click();

    // Snapshot ordering: server bumps seq on every accepted action.
    // Both pages should now show the same new active seat (Bob).
    await expect(bob.page.getByRole("button", { name: "pass turn" })).toBeEnabled({
      timeout: 10_000,
    });
    await expect(alice.page.getByRole("button", { name: "pass turn" })).toBeDisabled();

    // ---- 8. Play round 2: Bob passes his turn. Turn number should
    //         increment from 1 to 2 when we wrap back to seat 0. ----
    await bob.page.getByRole("button", { name: "pass turn" }).click();
    await expect(alice.page.getByRole("button", { name: "pass turn" })).toBeEnabled({
      timeout: 10_000,
    });

    // ---- 9. Round 3: Alice passes again to prove the cursor really
    //         walks through turns, not just seats. ----
    await alice.page.getByRole("button", { name: "pass turn" }).click();
    await expect(bob.page.getByRole("button", { name: "pass turn" })).toBeEnabled({
      timeout: 10_000,
    });

    // ---- 10. Final assertion on authoritative state via WS snapshot
    //          snooping. The server doesn't expose an HTTP "get turn"
    //          endpoint; instead, read the store the UI is bound to
    //          via page.evaluate against each page's session.
    const snapshotTurn = await alice.page.evaluate(async () => {
      const { token, playerID, gameID } = JSON.parse(
        localStorage.getItem("cmdctrl.session") ?? "{}",
      ) as { token?: string; playerID?: string; gameID?: string };
      if (!token || !playerID || !gameID) throw new Error("session missing");
      const proto = location.protocol === "https:" ? "wss:" : "ws:";
      const url = `${proto}//${location.host}/ws?game=${gameID}&player=${playerID}&token=${token}`;
      return await new Promise<{
        number: number;
        active_seat: number;
      }>((resolve, reject) => {
        const ws = new WebSocket(url);
        const timer = setTimeout(() => {
          ws.close();
          reject(new Error("timed out waiting for snapshot"));
        }, 5000);
        ws.onmessage = (ev) => {
          const frame = JSON.parse(String(ev.data)) as {
            kind: string;
            payload?: { game?: { turn?: { number: number; active_seat: number } } };
          };
          if (frame.kind === "snapshot" && frame.payload?.game?.turn) {
            clearTimeout(timer);
            ws.close();
            resolve(frame.payload.game.turn);
          }
        };
        ws.onerror = () => {
          clearTimeout(timer);
          reject(new Error("ws error"));
        };
      });
    });

    // After 3 pass_turns in a 2-player game:
    //   start:            turn=1 seat=0 (Alice)
    //   after pass #1 :   turn=1 seat=1 (Bob)
    //   after pass #2 :   turn=2 seat=0 (Alice, number++ on wrap)
    //   after pass #3 :   turn=2 seat=1 (Bob)
    expect(snapshotTurn.number).toBe(2);
    expect(snapshotTurn.active_seat).toBe(1);

    await alice.context.close();
    await bob.context.close();
  });
});
