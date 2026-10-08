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

// ADR 0130 PR 2 (#2048): exerting Oketra's Avenger from the client.
//
// Owner decision 1 is "every path asks". The path this walks is the
// widest one: "Attack with all" with an exert creature among the
// attackers opens Choose attackers instead of sending, the Avenger's
// row has an Exert toggle that is off, and turning it on sends the
// declaration with the Avenger exerted. The server then pays the exert
// as the declaration locks in: the Avenger won't untap during its
// controller's next untap step (no_untap.next), and the Bear, which
// was never asked about, is not exerted.

const AVENGER = "Oketra's Avenger";
const BEAR = "Grizzly Bears";

function makeDeck(): string {
  const lines = ["Commander:", `1 ${COMMANDER_NAME}`, "", "Mainboard:"];
  lines.push(`1 ${AVENGER}`, `1 ${BEAR}`, "97 Plains");
  return lines.join("\n") + "\n";
}

type NoUntapCard = { no_untap?: { next?: string[] }; attacking_target?: string };

function noUntapNext(v: SnapshotView, name: string): string[] {
  const c = findCardOnBattlefield(v, name) as (NoUntapCard & object) | null;
  return c?.no_untap?.next ?? [];
}

function attacking(v: SnapshotView, name: string): string | undefined {
  return (findCardOnBattlefield(v, name) as NoUntapCard | null)?.attacking_target;
}

test.describe("ADR 0130 exert", () => {
  test("Attack with all opens the picker, and its Exert toggle exerts the Avenger", async ({
    browser,
    request,
  }) => {
    test.slow();

    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Exert 2048 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");

      first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One");
      second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two");
      await uploadDeckAs(request, adminToken, game.id, first.playerID, makeDeck());
      await uploadDeckAs(request, adminToken, game.id, second.playerID, makeDeck());
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
      const defender = attacker === first ? second : first;

      for (const name of [AVENGER, BEAR]) {
        await adminMoveByName(admin, attacker.playerID, name, "library", "battlefield");
      }
      await admin.waitFor(
        (v) => [AVENGER, BEAR].every((n) => findCardOnBattlefield(v, n) !== null),
        "the Avenger and the Bear on the battlefield",
      );
      // Clears summoning sickness the way the untap step does (#318's
      // spec says why).
      await admin.sendActionAsPlayer(attacker.playerID, "untap_all", {});

      await admin.sendAction("advance_step", {});
      await admin.waitFor(
        (v) => v.turn?.step === "declare_attackers",
        "cursor on declare_attackers",
        20_000,
      );

      const dock = attacker.page.getByRole("region", { name: L.actions, exact: true });
      const cluster = dock.getByRole("group", { name: L.declareAttackers });
      await expect(cluster).toBeVisible({ timeout: 15_000 });
      const attackAll = cluster.getByRole("button", {
        name: new RegExp(`Attack ${defender.name} with all 2`),
      });
      await expect(attackAll).toBeVisible();
      await attackAll.click();

      // Not sent: the picker asks first.
      const exert = attacker.page.getByRole("switch", { name: L.exertAttacker(AVENGER) });
      await expect(exert).toBeVisible({ timeout: 10_000 });
      await expect(exert).toHaveAttribute("aria-checked", "false");
      expect(attacking(admin.snapshot(), AVENGER)).toBeUndefined();

      await exert.click();
      await expect(exert).toHaveAttribute("aria-checked", "true");
      await dock.getByRole("button", { name: /^Attack with 2/ }).click();

      await admin.waitFor(
        (v) => attacking(v, AVENGER) === defender.playerID && attacking(v, BEAR) === defender.playerID,
        "both declared at the defender",
        15_000,
      );
      await admin.waitFor(
        (v) => noUntapNext(v, AVENGER).length > 0,
        "the Avenger is exerted: it won't untap during the next untap step",
        15_000,
      );
      expect(noUntapNext(admin.snapshot(), BEAR), "the Bear was not exerted").toEqual([]);
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});
