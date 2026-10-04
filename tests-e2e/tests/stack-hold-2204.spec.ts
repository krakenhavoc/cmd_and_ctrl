import { expect, test } from "@playwright/test";
import { CARDS } from "./s19-deck-fixture";
import {
  seedHandWithCard,
  setupS19Game,
  triggerOnStack,
  type S19Setup,
} from "./s19-helpers";

// ADR 0119 §2 (#2204): the stack hold. Smart auto-pass used to pass an
// opponent's item the viewer could not answer in one round trip, so it
// was often on screen for less than a second. Now an automatic pass on
// someone else's top item waits until that item has been on this
// client's screen for gameplay.stackHoldMs (2 s by default).
//
// The opponent here is on the shipped defaults: smart auto-pass on, no
// "always stop", the 2 s hold. Its deck has no instants and it has no
// lands out, so it cannot answer anything and smart auto-pass would
// pass at once without the hold. The caster keeps the S19 seed
// (stackHoldMs 0), and its own item passes at once under
// autoPassOwnStack, as before. So the time the item spends on the stack
// is the opponent's hold.
//
// The item is Mulldrifter's enters trigger, put there by an admin move,
// the same way the S19 suite stages it: no mana, no casting UI. It is
// controlled by the caster, so to the opponent it is "an opponent's
// item", which is what the hold is keyed on.

const LABEL = /Mulldrifter — draw two cards/i;
const HOLD_MS = 2000;

test.describe("ADR 0119 stack hold", () => {
  let setup: S19Setup | null = null;

  test.afterEach(async () => {
    if (setup) {
      await setup.shutdown();
      setup = null;
    }
  });

  test("an opponent's item the viewer cannot answer stays up for the hold, then resolves", async ({
    browser,
    request,
  }) => {
    test.slow();
    setup = await setupS19Game(browser, request, { opponentGameplay: {} });
    const { admin, caster, opponent } = setup;

    const mulldrifter = await seedHandWithCard(admin, caster.playerID, CARDS.Mulldrifter);

    await admin.sendActionAsPlayer(caster.playerID, "move_card", {
      src: { kind: "hand", owner: caster.playerID },
      dst: { kind: "battlefield" },
      instance_id: mulldrifter.instance_id,
    });
    await admin.waitFor(
      (v) => triggerOnStack(v, CARDS.Mulldrifter) !== null,
      "Mulldrifter's trigger on the stack",
    );
    // The admin socket hears the broadcast no earlier than the server
    // sent it, and the opponent's client cannot have seen the item
    // before then either, so its hold cannot end before
    // appeared + HOLD_MS minus the admin socket's own delay.
    const appeared = Date.now();

    // When each page shows the item. Watched from now, in parallel,
    // because the item is only up for about the hold.
    const seenAt = (page: typeof caster.page) =>
      expect(page.getByText(LABEL).first())
        .toBeVisible({ timeout: 20_000 })
        .then(() => Date.now());
    const casterSeen = seenAt(caster.page);
    const opponentSeen = seenAt(opponent.page);

    // 1.5 s after it appeared it is still on the stack: the opponent's
    // automatic pass is waiting out the hold.
    const untilCheck = appeared + 1500 - Date.now();
    if (untilCheck > 0) await new Promise((r) => setTimeout(r, untilCheck));
    expect(
      triggerOnStack(admin.snapshot(), CARDS.Mulldrifter),
      "the trigger resolved before the opponent's 2 s stack hold was over",
    ).not.toBeNull();

    // And it is gone within the hold plus 2 s, measured from the later
    // of the moment it appeared and the moments both pages showed it:
    // a page that is a beat behind under load starts its hold late,
    // and that is load, not the hold.
    const shown = Math.max(appeared, await casterSeen, await opponentSeen);
    const deadline = shown + HOLD_MS + 2000;
    await admin.waitFor(
      (v) => triggerOnStack(v, CARDS.Mulldrifter) === null,
      "the trigger resolves once the hold is over",
      Math.max(1, deadline - Date.now()),
    );
  });
});
