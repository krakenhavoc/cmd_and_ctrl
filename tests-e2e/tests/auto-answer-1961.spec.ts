import { expect, test, type Page } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { CARDS } from "./s19-deck-fixture";
import {
  adminMoveByName,
  findCardOnBattlefield,
  resolveStack,
  setupS19Game,
  triggerOnStack,
  type S19Setup,
  type SnapshotView,
} from "./s19-helpers";

// ADR 0127 (#1961): answering a repeated prompt once. The opponent's
// Smothering Tithe taxes the caster's draws — the same pay-unless shape
// as Rhystic Study's tax, and already in the S19 decks. The caster ticks
// "Remember this answer" and presses "Don't pay". On the next draw the
// SERVER answers for them: no prompt, a Treasure for the Tithe's
// controller, an "(automatic)" log line every seat can read, and a
// notice in the caster's dock. "Ask me next time" there brings the
// prompt back on the draw after that.

const dockOf = (page: Page) =>
  page.getByRole("region", { name: "actions", exact: true });

type LogLine = { kind: string; text: string };
const logOf = (v: SnapshotView): LogLine[] =>
  (v as unknown as { log?: LogLine[] }).log ?? [];

function treasures(v: SnapshotView, controller: string): number {
  const cards =
    (
      v as unknown as {
        battlefield?: { cards?: { name: string; controller: string }[] };
      }
    ).battlefield?.cards ?? [];
  return cards.filter(
    (c) => c.name === CARDS.Treasure && c.controller === controller,
  ).length;
}

test.describe("ADR 0127 automatic answers", () => {
  let setup: S19Setup | null = null;

  test.afterEach(async () => {
    if (setup) {
      await setup.shutdown();
      setup = null;
    }
  });

  test("Remember this answer, the automatic answer, and Ask me next time", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request);
    const { admin, caster, opponent } = setup;

    await adminMoveByName(
      admin,
      opponent.playerID,
      CARDS.SmotheringTithe,
      "library",
      "battlefield",
    );
    await admin.waitFor(
      (v) => findCardOnBattlefield(v, CARDS.SmotheringTithe) !== null,
      "Smothering Tithe on battlefield",
    );

    const drawIntoTheTax = async (n: number) => {
      await admin.sendActionAsPlayer(caster.playerID, "draw_card", {});
      await admin.waitFor(
        (v) => triggerOnStack(v, CARDS.SmotheringTithe) !== null,
        `Tithe trigger on the stack (draw ${n})`,
      );
      await resolveStack(setup!);
    };

    // 1. Asked: tick Remember this answer, then Don't pay.
    await drawIntoTheTax(1);
    const dock = dockOf(caster.page);
    const remember = dock.getByRole("button", { name: L.rememberThisAnswer });
    await expect(remember).toBeVisible({ timeout: 20_000 });
    await expect(remember).toHaveAttribute("aria-pressed", "false");
    await remember.click();
    await expect(remember).toHaveAttribute("aria-pressed", "true");
    await dock.getByRole("button", { name: /^Don't pay$/ }).click();
    await admin.waitFor(
      (v) => treasures(v, opponent.playerID) === 1,
      "first Treasure",
    );
    // The setting reaches the server on the caster's next frame
    // (set_auto_answers, the reconcile). Give that round trip a beat
    // before the next draw, which comes from another socket.
    await caster.page.waitForTimeout(1500);

    // 2. Answered for them: no prompt, a second Treasure, the log line.
    await drawIntoTheTax(2);
    const after = await admin.waitFor(
      (v) => treasures(v, opponent.playerID) === 2,
      "second Treasure, with nobody clicking",
    );
    expect(
      (after.pending_choices ?? []).filter((c) => c.kind === "pay_unless"),
    ).toHaveLength(0);
    const line = logOf(after).find((e) => e.kind === "auto_answer");
    expect(line?.text).toBe(
      `Caster didn't pay for ${CARDS.SmotheringTithe} (automatic)`,
    );

    // The caster's dock says so, with Undo and Ask me next time.
    const notice = dock.getByRole("dialog", { name: L.automaticAnswer });
    await expect(notice).toBeVisible({ timeout: 20_000 });
    await expect(notice).toContainText(
      `${CARDS.SmotheringTithe}: didn't pay for you`,
    );
    await expect(
      notice.getByRole("button", { name: L.undoAutomaticAnswer }),
    ).toBeVisible();

    // 3. Ask me next time: the rule is gone, and the prompt comes back.
    await notice.getByRole("button", { name: L.askMeNextTime }).click();
    await expect(notice).toHaveCount(0);
    await caster.page.waitForTimeout(1500);
    await drawIntoTheTax(3);
    await admin.waitFor(
      (v) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "pay_unless" && c.chooser === caster.playerID,
        ),
      "the tax is asked again",
    );
    await expect(dock.getByRole("button", { name: /^Don't pay$/ })).toBeVisible(
      { timeout: 20_000 },
    );
  });
});
