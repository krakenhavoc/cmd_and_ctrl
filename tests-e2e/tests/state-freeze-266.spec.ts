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
  test("walking priority window-by-window keeps both boards live", async ({
    browser,
    request,
  }) => {
    test.slow();

    const pageErrors: string[] = [];
    const watch = (page: Page, who: string) => {
      page.on("pageerror", (e) =>
        pageErrors.push(`${who} pageerror: ${e.message}`),
      );
      page.on("console", (m) => {
        if (m.type() === "error")
          pageErrors.push(`${who} console.error: ${m.text()}`);
      });
    };

    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Freeze 266 ${Date.now()}`,
    );
    const alice = await joinAsPlayer(
      browser,
      game.id,
      game.invite_token!,
      "Alice",
    );
    const bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");
    watch(alice.page, "alice");
    watch(bob.page, "bob");

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
    }
    for (const p of [alice, bob]) {
      await expect(
        p.page.getByRole("dialog", { name: /keep or mulligan/i }),
      ).toHaveCount(0, {
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
        const dialog = p.page.getByRole("dialog", {
          name: /discard \d+ card/i,
        });
        if (!(await dialog.isVisible().catch(() => false))) continue;
        const countText =
          (await dialog.locator(".prompt-count").textContent()) ?? "0 / 0";
        const owed = Number(/(\d+)\s*\/\s*(\d+)/.exec(countText)?.[2] ?? 1);
        for (let n = 0; n < Math.max(1, owed); n++) {
          await dialog
            .locator("button.card-pick:not([disabled])")
            .first()
            .click();
        }
        await dialog
          .getByRole("button", { name: "Discard", exact: true })
          .click();
        await expect(dialog).toHaveCount(0, { timeout: 10_000 });
      }
    };

    const startTurn = await turnOf(alice.page);

    // #602: an explicit budget rather than test.slow()'s 3x. The walk
    // below waits on a cursor it does not control, and the failure
    // this spec exists to catch (a board that stops updating) is not
    // the one a 90s cap produces.
    test.setTimeout(180_000);

    // 8 windows still crosses a turn boundary and walks both players
    // through the combat steps; each one costs real wall clock on a
    // single-worker runner shared with the rest of the suite.
    let held = 0;
    for (let i = 0; i < 8; i++) {
      // Whether any seat HOLDS priority is not this spec's business.
      // Since #429 the client asks the server's move list before
      // auto-passing, so most windows pass on their own and the
      // cursor keeps moving without anybody clicking. That is the
      // engine working, not a freeze — so a window nobody holds is
      // skipped, not failed. What must stay true is that the boards
      // agree and the game advances, which is asserted either way.
      const someoneHolds = await expect
        .poll(
          async () => {
            await answerDiscards();
            for (const p of players) {
              if (
                await p.page
                  .getByRole("button", { name: "next", exact: true })
                  .isEnabled()
              ) {
                return true;
              }
            }
            return false;
          },
          { timeout: 8_000, message: `polling for priority at iteration ${i}` },
        )
        .toBe(true)
        .then(
          () => true,
          () => false,
        );

      if (someoneHolds) {
        held++;
        for (const p of players) {
          const next = p.page.getByRole("button", {
            name: "next",
            exact: true,
          });
          if (await next.isEnabled()) {
            await next.click().catch(() => {});
            break;
          }
        }
      }

      // Both clients must land on the same step AND the same turn. A
      // frozen client keeps rendering whatever it last saw while the
      // other moves on; comparing the step alone can agree by
      // coincidence across a turn boundary.
      await expect
        .poll(
          async () => {
            const [as, bs, at, bt] = await Promise.all([
              stepOf(alice.page),
              stepOf(bob.page),
              turnOf(alice.page),
              turnOf(bob.page),
            ]);
            return as === bs && at === bt;
          },
          { timeout: 15_000, message: `boards diverged at iteration ${i}` },
        )
        .toBe(true);
    }

    // A walk in which no seat ever held priority is not a walk. The
    // cursor auto-passing every window is legitimate; eight of them
    // with nobody ever able to act is the stuck-cursor case this
    // spec would otherwise report as a pass.
    expect(
      held,
      "no seat held priority in any of the 8 windows",
    ).toBeGreaterThan(0);

    // The walk must have actually moved the game on, not just spun.
    expect(
      await turnOf(alice.page),
      `turn counter never left ${startTurn}`,
    ).not.toBe(startTurn);

    expect(pageErrors, `client errors:\n${pageErrors.join("\n")}`).toEqual([]);
  });
});
