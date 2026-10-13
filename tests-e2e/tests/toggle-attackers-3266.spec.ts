import { expect, test } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  findCardOnBattlefield,
  keepAllHands,
  openAdminClient,
  type AdminClient,
  type SnapshotView,
} from "./s19-helpers";

// #3266: declaring attackers by toggling creatures. In a duel a click
// on a creature that can attack stages it at the one opponent (nothing
// is sent), a second click takes it back out, and the dock's primary
// counts what is staged: "No attacks" with nothing, "Attack with N"
// otherwise. Pressing it sends the staged set as one declaration and
// passes, so the game moves past declare attackers.
//
// The board is staged through the admin WS, as in attack-all-318.

const BEARS = ["Grizzly Bears", "Runeclaw Bear"] as const;

function makeBearDeck(): string {
  const lines = ["Commander:", `1 ${COMMANDER_NAME}`, "", "Mainboard:"];
  for (const bear of BEARS) lines.push(`1 ${bear}`);
  lines.push(`${99 - BEARS.length} Forest`);
  return lines.join("\n") + "\n";
}

function attackerNames(v: SnapshotView): string[] {
  return v.battlefield.cards
    .filter((c) => !!(c as { attacking_target?: string }).attacking_target)
    .map((c) => c.name ?? "?")
    .sort();
}

test.describe("#3266 toggle attackers", () => {
  test("clicks stage and unstage creatures; Attack with N declares the staged set", async ({
    browser,
    request,
  }) => {
    test.slow();

    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Toggle attackers 3266 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");

      first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One");
      second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two");
      await uploadDeckAs(request, adminToken, game.id, first.playerID, makeBearDeck());
      await uploadDeckAs(request, adminToken, game.id, second.playerID, makeBearDeck());
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
      await keepAllHands(admin);
      await admin.waitFor((v) => v.state === "active", "game state active");
      await admin.waitFor(
        (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the attacker's precombat main",
        15_000,
      );

      const snap = admin.snapshot();
      const activeID = snap.seats[snap.turn?.active_seat ?? 0]?.id;
      const attacker = activeID === first.playerID ? first : second;

      for (const bear of BEARS) {
        await adminMoveByName(admin, attacker.playerID, bear, "library", "battlefield");
      }
      await admin.waitFor(
        (v) => BEARS.every((b) => findCardOnBattlefield(v, b) !== null),
        "both bears on the battlefield",
      );
      await admin.sendActionAsPlayer(attacker.playerID, "untap_all", {});

      await admin.sendAction("advance_step", {});
      await admin.waitFor(
        (v) => v.turn?.step === "declare_attackers",
        "cursor on declare_attackers",
        20_000,
      );

      const page = attacker.page;
      const dock = page.getByRole("region", { name: L.actions, exact: true });
      const primary = (name: string) => dock.getByRole("button", { name, exact: true });
      const board = page.getByRole("region", { name: L.yourBoard, exact: true });
      const bear = (name: string) =>
        board.getByRole("button", { name: new RegExp(`^${name}`) }).first();

      await expect(primary(L.noAttacks)).toBeVisible({ timeout: 15_000 });

      await bear(BEARS[0]).click();
      await expect(primary(L.attackWith(1))).toBeVisible();
      await bear(BEARS[1]).click();
      await expect(primary(L.attackWith(2))).toBeVisible();
      // The second click takes the first bear back out.
      await bear(BEARS[0]).click();
      await expect(primary(L.attackWith(1))).toBeVisible();

      // Nothing was sent while staging.
      expect(attackerNames(admin.snapshot())).toEqual([]);

      await primary(L.attackWith(1)).click();
      await admin.waitFor(
        (v) => attackerNames(v).length === 1,
        "the staged bear declared",
        15_000,
      );
      expect(attackerNames(admin.snapshot())).toEqual([BEARS[1]]);
      // The press passed as well: the game left declare attackers.
      await admin.waitFor(
        (v) => v.turn?.step !== "declare_attackers",
        "the game moved past declare attackers",
        20_000,
      );
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});
