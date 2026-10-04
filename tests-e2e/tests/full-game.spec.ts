import { expect, test, type Page } from "@playwright/test";
import { adminLogin, createGame, uploadDeckAs, startGameAs, getGameAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { joinAsPlayer } from "./players";

// End-to-end happy path: two players join, upload decks, start the
// game, roll for the first turn, keep their opening hands, and each take
// a turn. This exercises every layer — lobby HTTP, invite flow, deck
// import against the Scryfall index, game Start, the opening roll (ADR
// 0121: both press Roll, the winner gives the first turn away and
// confirms), WebSocket snapshot propagation, mulligan window, and the
// turn cursor.
//
// Prerequisites:
//   - Server running with the dev-default admin token (make server-dev).
//   - Scryfall bulk dump present at <repo>/data/scryfall/default-cards.json
//     so the deck fixture's cards (Kenrith, Plains) resolve.

test.describe("full game", () => {
  test("2 players: join, upload decks, start, roll, keep hands, play 2 turns", async ({
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

    // ---- 4. Start the game. The opening roll is played through the
    //         UI below, so the helper leaves it open. ----
    const started = await startGameAs(request, adminToken, game.id, { finishRoll: false });
    expect(started.state).toBe("active");

    // ---- 4b. The opening roll (ADR 0121 §6). Each page gets its own
    //          `roll for the first turn` request and the strip's
    //          `opening roll` banner; nothing is dealt yet. ----
    const rollDialog = (p: { page: Page }) =>
      p.page.getByRole("dialog", { name: "roll for the first turn", exact: true });
    const banner = (p: { page: Page }) =>
      p.page.getByRole("group", { name: "opening roll", exact: true });
    const chooseSheet = (p: { page: Page }) =>
      p.page.getByRole("dialog", { name: "choose who takes the first turn", exact: true });
    for (const p of [alice, bob]) {
      await expect(rollDialog(p)).toBeVisible({ timeout: 10_000 });
      await expect(rollDialog(p)).toContainText(
        "Roll a d20. The highest roll chooses who goes first.",
      );
      await expect(banner(p)).toBeVisible();
      await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toHaveCount(0);
    }

    // Both press Roll. A tie (1 in 20) opens a reroll for both with the
    // same button, so keep pressing until one page has the chooser's
    // sheet.
    const rollUntilAWinner = async (): Promise<typeof alice> => {
      const deadline = Date.now() + 30_000;
      while (Date.now() < deadline) {
        for (const p of [alice, bob]) {
          const roll = rollDialog(p).getByRole("button", { name: "Roll", exact: true });
          if ((await roll.isVisible()) && (await roll.isEnabled())) await roll.click();
        }
        for (const p of [alice, bob]) {
          if (await chooseSheet(p).isVisible()) return p;
        }
        await alice.page.waitForTimeout(200);
      }
      throw new Error("the opening roll never found a winner");
    };
    const winner = await rollUntilAWinner();
    const loser = winner === alice ? bob : alice;

    // Each page sees both results in the banner, and the loser reads
    // who is choosing in the dock's status line.
    for (const p of [alice, bob]) {
      await expect(banner(p)).toContainText(/Alice\s*\d+/);
      await expect(banner(p)).toContainText(/Bob\s*\d+/);
      await expect(banner(p)).not.toContainText("rolling");
    }
    await expect(
      loser.page.getByRole("region", { name: "actions", exact: true }),
    ).toContainText(`${winner.name} is choosing who goes first`);

    // The winner gives the first turn away, and has to confirm it
    // (owner decision 6): the choice has no undo.
    await expect(chooseSheet(winner).getByRole("button", { name: "I go first" })).toBeVisible();
    await chooseSheet(winner)
      .getByRole("button", { name: `${loser.name} goes first`, exact: true })
      .click();
    const confirm = winner.page.getByRole("dialog", {
      name: `Let ${loser.name} take the first turn?`,
      exact: true,
    });
    await expect(confirm).toBeVisible();
    await confirm.getByRole("button", { name: "Confirm", exact: true }).click();

    // ---- 5. Both players' game routes should flip into the
    //         mulligan window. The WS push happens automatically
    //         because both pages were already on /#/games/<id>.
    for (const p of [alice, bob]) {
      await expect(banner(p)).toHaveCount(0, { timeout: 10_000 });
      await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toBeVisible({
        timeout: 10_000,
      });
      // Each player sees 7 cards in their opening hand. The count
      // also surfaces in the dialog header, so assert on both.
      await expect(p.page.getByText(/Hand size: 7/)).toBeVisible();
      // The roll call's pill reads the winner's choice.
      await expect(
        p.page
          .getByLabel("opening hand decisions")
          .getByText(`chose ${loser.name} to go first`, { exact: false }),
      ).toBeVisible();
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

    // ---- 7. Play round 1: the starting seat passes its turn. ----
    // The winner of the opening roll handed the first turn to the
    // other seat, so the loser starts. pass_turn jumps to the next
    // seat's untap step. Only the active seat's button is enabled; the
    // other browser sees it disabled.
    // ADR 0111 PR 2: Pass turn is in the action dock's action bar, and
    // is rendered disabled (not hidden) for the seat that is not active.
    const passOf = (p: { page: Page }) =>
      p.page
        .getByRole("region", { name: "actions", exact: true })
        .getByRole("button", { name: "pass turn" });
    const first = loser;
    const second = winner;
    const aliceStarts = first === alice;
    await expect(passOf(first)).toBeEnabled({ timeout: 10_000 });
    await expect(passOf(second)).toBeDisabled();

    await passOf(first).click();

    // Snapshot ordering: server bumps seq on every accepted action.
    // Both pages should now show the same new active seat.
    await expect(passOf(second)).toBeEnabled({ timeout: 10_000 });
    await expect(passOf(first)).toBeDisabled();

    // ---- 8. Play round 2: the second seat passes its turn. ----
    await passOf(second).click();
    await expect(passOf(first)).toBeEnabled({ timeout: 10_000 });

    // ---- 9. Round 3: the first seat passes again to prove the
    //         cursor really walks through turns, not just seats. ----
    await passOf(first).click();
    await expect(passOf(second)).toBeEnabled({ timeout: 10_000 });

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

    // After 3 pass_turns in a 2-player game the cursor is on the seat
    // that did not start (Alice is seat 0, Bob seat 1), and the turn
    // number has wrapped at least once.
    expect(snapshotTurn.number).toBeGreaterThanOrEqual(2);
    expect(snapshotTurn.active_seat).toBe(aliceStarts ? 1 : 0);

    await alice.context.close();
    await bob.context.close();
  });
});
