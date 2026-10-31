import { expect, test } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  keepAllHands,
  openAdminClient,
  playerByID,
  resolveStack,
  type S19Setup,
} from "./s19-helpers";

// #4497: a spell whose first target is "up to one" can be cast with that
// target skipped. Meager Meal, the Adventure half of Gollum, Silent
// Slinker, reads "Put a +1/+1 counter on up to one target creature.
// Target player gains 2 life." Choosing no creature is a legal
// announcement (CR 601.2c), and then the one target answers the player
// clause, so the cast must file it there: slot 1.
//
// Through the UI, at two seats: the caster casts the Adventure half, says
// Done to the creature clause with nothing picked, picks the opponent for
// the player clause, and the spell goes on the stack with the opponent
// filed under slot 1. It resolves, and the opponent gains 2 life.
//
// Tips are off from the start, as in target-player-reach-4435: a first-use
// hint card can open over an avatar once targeting starts.

const GOLLUM = "Gollum, Silent Slinker";
const MEAL = "Meager Meal";
const NAMES = ["Player One", "Player Two"];

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${GOLLUM}`,
      "98 Swamp",
    ].join("\n") + "\n"
  );
}

test("#4497: Meager Meal is cast with its 'up to one' creature skipped", async ({
  browser,
  request,
}) => {
  test.setTimeout(180_000);
  const players: JoinedPlayer[] = [];
  let admin: Awaited<ReturnType<typeof openAdminClient>> | null = null;
  try {
    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Skipped target 4497 ${Date.now()}`,
    );
    if (!game.invite_token) throw new Error("invite token missing");
    for (const name of NAMES) {
      players.push(
        await joinAsPlayer(browser, game.id, game.invite_token, name, {
          help: { tipsOff: true },
          // Hold on every stack item, so the spell is read on the stack
          // before anyone passes it.
          gameplay: { passMode: "careful" },
        }),
      );
    }
    for (const p of players) {
      await uploadDeckAs(request, adminToken, game.id, p.playerID, makeDeck());
    }
    await startGameAs(request, adminToken, game.id);
    admin = await openAdminClient(
      adminToken,
      game.id,
      players[0].playerID,
      players[1].playerID,
    );
    await keepAllHands(admin);
    await admin.waitFor(
      (v) =>
        v.turn?.step === "precombat_main" &&
        v.turn?.priority_holder === v.turn?.active_seat,
      "the active seat's precombat main",
      20_000,
    );
    const snap = admin.snapshot();
    const activeID = snap.seats[snap.turn?.active_seat ?? 0]?.id;
    const caster = players.find((p) => p.playerID === activeID)!;
    const opponent = players.find((p) => p !== caster)!;

    // One Swamp to pay {B}, and the card in hand.
    await adminMoveByName(
      admin,
      caster.playerID,
      "Swamp",
      "library",
      "battlefield",
    );
    await admin.sendActionAsPlayer(caster.playerID, "untap_all", {});
    let gollumID = "";
    for (const c of playerByID(admin.snapshot(), caster.playerID).library
      .cards) {
      if ((await admin.cardName(c.instance_id))?.startsWith(GOLLUM)) {
        gollumID = c.instance_id;
        break;
      }
    }
    if (!gollumID) {
      gollumID =
        playerByID(admin.snapshot(), caster.playerID).hand.cards.find((c) =>
          c.name?.startsWith(GOLLUM),
        )?.instance_id ?? "";
    } else {
      await admin.sendActionAsPlayer(caster.playerID, "move_card", {
        src: { kind: "library", owner: caster.playerID },
        dst: { kind: "hand", owner: caster.playerID },
        instance_id: gollumID,
      });
    }
    expect(gollumID, "Gollum is in the deck").not.toBe("");
    await admin.waitFor(
      (v) =>
        playerByID(v, caster.playerID).hand.cards.some(
          (c) => c.instance_id === gollumID,
        ),
      "Gollum in the caster's hand",
    );

    const page = caster.page;
    await page.setViewportSize({ width: 1366, height: 768 });
    // Enter on the focused card casts it, as a click does.
    const card = page.locator(`[data-instance-id="${gollumID}"]`).first();
    await expect(card).toBeVisible({ timeout: 15_000 });
    await card.focus();
    await page.keyboard.press("Enter");

    // CR 715.3: which half. The Adventure.
    const half = page.locator("button.face-opt", { hasText: MEAL });
    await expect(half).toBeVisible({ timeout: 10_000 });
    await half.click();
    await page.getByRole("button", { name: `Cast ${MEAL}` }).click();

    // The creature clause, left empty: Done with nothing picked.
    const prompt = page.getByRole("dialog", { name: L.selectTarget(MEAL) });
    await expect(prompt).toBeVisible({ timeout: 10_000 });
    await expect(prompt).toContainText("up to one target creature");
    const done = prompt.getByRole("button", { name: "Done", exact: true });
    await expect(done).toBeEnabled();
    await done.click();

    // The player clause: the opponent, picked on the board.
    await expect(prompt).toContainText("target player");
    const avatar = page
      .locator(
        `.identity.cast-targetable .avatar-wrap[data-seat-id="${opponent.playerID}"]`,
      )
      .first();
    await expect(avatar).toBeVisible({ timeout: 10_000 });
    await avatar.focus();
    await page.keyboard.press("Enter");
    await expect(prompt).toBeHidden({ timeout: 10_000 });

    // On the stack, the opponent filed under the player clause.
    const onStack = await admin.waitFor(
      (v) => (v.stack_items ?? []).some((s) => s.source_card_id === gollumID),
      "Meager Meal on the stack",
      15_000,
    );
    const item = (onStack.stack_items ?? []).find(
      (s) => s.source_card_id === gollumID,
    )!;
    expect(item.targets).toEqual([
      { kind: "player", id: opponent.playerID, slot: 1 },
    ]);

    const lifeOf = (v: typeof onStack, id: string) =>
      (playerByID(v, id) as unknown as { life: number }).life;
    const before = lifeOf(onStack, opponent.playerID);
    const setup = { admin, caster, opponent } as unknown as S19Setup;
    const after = await resolveStack(setup);
    expect(lifeOf(after, opponent.playerID)).toBe(before + 2);
  } finally {
    admin?.close();
    await closeAll(...players);
  }
});
