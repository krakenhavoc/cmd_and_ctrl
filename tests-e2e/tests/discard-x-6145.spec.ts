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
} from "./s19-helpers";

// #6145: "As an additional cost to cast this spell, discard X cards."
// Channeled Force prints no {X} in its mana cost, so X is announced first
// (CR 107.3a) and the discard picker that follows asks for exactly X cards
// (CR 601.2f). Through the UI, at two seats: the caster announces X = 2,
// picks two cards to discard, targets the opponent (no creature is on the
// table, so the "up to one" clause has nothing to pick), and the spell goes
// on the stack with X = 2 and the two cards in the caster's graveyard.

const FORCE = "Channeled Force";
const NAMES = ["Player One", "Player Two"];

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${FORCE}`,
      "49 Island",
      "49 Mountain",
    ].join("\n") + "\n"
  );
}

test("#6145: Channeled Force announces X, then discards exactly X cards", async ({
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
      `Discard X 6145 ${Date.now()}`,
    );
    if (!game.invite_token) throw new Error("invite token missing");
    for (const name of NAMES) {
      players.push(
        await joinAsPlayer(browser, game.id, game.invite_token, name, {
          help: { tipsOff: true },
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

    // {2}{U}{R}: two Islands and two Mountains.
    for (const land of ["Island", "Island", "Mountain", "Mountain"]) {
      await adminMoveByName(
        admin,
        caster.playerID,
        land,
        "library",
        "battlefield",
      );
    }
    await admin.sendActionAsPlayer(caster.playerID, "untap_all", {});
    await adminMoveByName(admin, caster.playerID, FORCE, "library", "hand");
    const inHand = await admin.waitFor(
      (v) =>
        playerByID(v, caster.playerID).hand.cards.some((c) => c.name === FORCE),
      "Channeled Force in the caster's hand",
    );
    const hand = playerByID(inHand, caster.playerID).hand.cards;
    const forceID = hand.find((c) => c.name === FORCE)!.instance_id;
    expect(
      hand.length,
      "cards to discard besides the spell",
    ).toBeGreaterThanOrEqual(3);

    const page = caster.page;
    await page.setViewportSize({ width: 1366, height: 768 });
    const card = page.locator(`[data-instance-id="${forceID}"]`).first();
    await expect(card).toBeVisible({ timeout: 15_000 });
    await card.focus();
    await page.keyboard.press("Enter");

    // X first: the mana cost has no {X}, the additional cost does.
    const xField = page.getByLabel("X value");
    await expect(xField).toBeVisible({ timeout: 10_000 });
    await xField.fill("2");
    await page.getByRole("button", { name: /with X = 2/ }).click();

    // Then the discard picker, sized to X.
    await expect(page.getByText("Discard 2 cards")).toBeVisible({
      timeout: 10_000,
    });
    const options = page.locator("button.prompt-opt");
    await options.nth(0).click();
    await options.nth(1).click();
    await page.getByRole("button", { name: "Discard", exact: true }).click();

    // The player clause: the opponent, picked on the board.
    const prompt = page.getByRole("dialog", { name: L.selectTarget(FORCE) });
    await expect(prompt).toBeVisible({ timeout: 10_000 });
    const avatar = page
      .locator(
        `.identity.cast-targetable .avatar-wrap[data-seat-id="${opponent.playerID}"]`,
      )
      .first();
    await expect(avatar).toBeVisible({ timeout: 10_000 });
    await avatar.focus();
    await page.keyboard.press("Enter");

    const onStack = await admin.waitFor(
      (v) => (v.stack_items ?? []).some((s) => s.source_card_id === forceID),
      "Channeled Force on the stack",
      15_000,
    );
    const item = (onStack.stack_items ?? []).find(
      (s) => s.source_card_id === forceID,
    )! as unknown as {
      x_value?: number;
    };
    expect(item.x_value).toBe(2);
    expect(playerByID(onStack, caster.playerID).graveyard.cards.length).toBe(2);
  } finally {
    admin?.close();
    await closeAll(...players);
  }
});
