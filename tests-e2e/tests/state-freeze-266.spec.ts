import { expect, test, type Page } from "@playwright/test";
import { adminLogin, createGame, uploadDeckAs, startGameAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { joinAsPlayer } from "./players";

// Guard for the #266 "[in-app] State freeze" class: the board stops
// updating while the socket stays green and the server keeps moving.
//
// full-game.spec.ts only ever uses `pass turn`, which jumps whole
// turns at a time. Nothing in the suite walked the cursor priority
// window by priority window with the `next` button — which is the
// path the reporter was on when the board froze, and the path that
// exercises every step transition, the cleanup discard prompt, and
// the combat steps' render.
//
// The invariants asserted here are the ones a freeze breaks:
//   1. both clients converge on the same step after every window,
//   2. the game actually progresses,
//   3. no uncaught error / console.error reaches either page — one
//      throw is all it takes to wedge the store graph for good.

test.describe("#266 state freeze", () => {
  test("walking priority window-by-window keeps both boards live", async ({ browser, request }) => {
    test.slow();

    const pageErrors: string[] = [];
    const watch = (page: Page, who: string) => {
      page.on("pageerror", (e) => pageErrors.push(`${who} pageerror: ${e.message}`));
      page.on("console", (m) => {
        if (m.type() === "error") pageErrors.push(`${who} console.error: ${m.text()}`);
      });
    };

    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Freeze 266 ${Date.now()}`);
    const alice = await joinAsPlayer(browser, game.id, game.invite_token!, "Alice");
    const bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");
    watch(alice.page, "alice");
    watch(bob.page, "bob");

    const deck = makeCommanderDeck();
    await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
    await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
    await startGameAs(request, adminToken, game.id);

    for (const p of [alice, bob]) {
      await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toBeVisible({
        timeout: 10_000,
      });
      await p.page.getByRole("button", { name: "Keep hand" }).click();
    }
    for (const p of [alice, bob]) {
      await expect(p.page.getByRole("dialog", { name: /keep or mulligan/i })).toHaveCount(0, {
        timeout: 10_000,
      });
    }

    const players = [alice, bob];
    // PhaseDisplay's step row and turn counter — both read straight
    // off the snapshot, so they are exactly "the board updated".
    const stepOf = async (page: Page): Promise<string> =>
      ((await page.locator(".step-label").first().textContent()) ?? "").trim();
    const turnOf = async (page: Page): Promise<string> =>
      ((await page.locator(".turn-no").first().textContent()) ?? "").trim();

    // Cleanup parks the cursor with PriorityHolder=NoPriority until
    // every owed discard is submitted, so a hand over max size
    // legitimately blocks the walk. Answer any open prompt.
    const answerDiscards = async (): Promise<void> => {
      for (const p of players) {
        const dialog = p.page.getByRole("dialog", { name: /discard \d+ card/i });
        if (!(await dialog.isVisible().catch(() => false))) continue;
        const countText = (await dialog.locator(".prompt-count").textContent()) ?? "0 / 0";
        const owed = Number(/(\d+)\s*\/\s*(\d+)/.exec(countText)?.[2] ?? 1);
        for (let n = 0; n < Math.max(1, owed); n++) {
          await dialog.locator("button.card-pick:not([disabled])").first().click();
        }
        await dialog.getByRole("button", { name: "Discard", exact: true }).click();
        await expect(dialog).toHaveCount(0, { timeout: 10_000 });
      }
    };

    const startTurn = await turnOf(alice.page);

    // 12 windows is enough to cross a turn boundary and walk both
    // players through every combat step; more only costs wall clock,
    // and this spec shares a single-worker runner with the rest.
    for (let i = 0; i < 12; i++) {
      // The default settings auto-pass most windows, so the cursor
      // keeps moving on its own; wait for it to land on a seat that
      // actually holds priority before clicking.
      await expect
        .poll(
          async () => {
            await answerDiscards();
            for (const p of players) {
              if (await p.page.getByRole("button", { name: "next", exact: true }).isEnabled()) {
                return true;
              }
            }
            return false;
          },
          {
            timeout: 20_000,
            message: `no seat held priority at iteration ${i}`,
          },
        )
        .toBe(true);

      for (const p of players) {
        const next = p.page.getByRole("button", { name: "next", exact: true });
        if (await next.isEnabled()) {
          await next.click().catch(() => {});
          break;
        }
      }

      // Both clients must land on the same step. A frozen client keeps
      // rendering whatever it last saw while the other moves on.
      await expect
        .poll(
          async () => {
            const [a, b] = await Promise.all(players.map((p) => stepOf(p.page)));
            return a === b;
          },
          { timeout: 15_000, message: `boards diverged at iteration ${i}` },
        )
        .toBe(true);
    }

    // The walk must have actually moved the game on, not just spun.
    expect(await turnOf(alice.page), `turn counter never left ${startTurn}`).not.toBe(startTurn);

    expect(pageErrors, `client errors:\n${pageErrors.join("\n")}`).toEqual([]);
  });
});
