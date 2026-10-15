import { expect, test, type Locator, type Page } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  keepAllHands,
  openAdminClient,
  type AdminClient,
  type SnapshotCard,
  type SnapshotView,
} from "./s19-helpers";

// #3898: declaring blockers on the board. The defender clicks their
// creature, then the attacker, with no picker in between:
//
//   1. With Grizzly Bears selected, Wind Drake (flying) is greyed and
//      says why on hover, and a click on it declares nothing.
//   2. The pile of two Persistent Petitioners fans out on a click, and
//      the second copy is blocked by clicking it.
//   3. The block is taken back by selecting the bear and clicking its
//      attacker again ("Remove block" is offered meanwhile).
//   4. Boggart Brute (menace) takes two: the bear, the brute, then
//      Balduvian Bears, and both blocks land as one declaration, each
//      drawn as a block arrow.
//   5. "Show as list" opens the same declaration as a list, and Done
//      blocking finishes it.
//
// The board is staged through the admin WS, as the #2956 spec does.

const DRAKE = "Wind Drake";
const BRUTE = "Boggart Brute";
const PETITIONERS = "Persistent Petitioners";
const BEAR = "Grizzly Bears";
const OTHER_BEAR = "Balduvian Bears";

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${DRAKE}`,
      `1 ${BRUTE}`,
      // "A deck can have any number of cards named Persistent
      // Petitioners": the two identical attackers that pile (#3046).
      `2 ${PETITIONERS}`,
      `1 ${BEAR}`,
      `1 ${OTHER_BEAR}`,
      "93 Plains",
    ].join("\n") + "\n"
  );
}

type Card = SnapshotCard & { blocking_target?: string };

const cards = (v: SnapshotView): Card[] => v.battlefield.cards as Card[];
const ofName = (v: SnapshotView, owner: string, name: string): Card[] =>
  cards(v).filter((c) => c.name === name && c.controller === owner);
const blockOf = (v: SnapshotView, id: string): string | undefined =>
  cards(v).find((c) => c.instance_id === id)?.blocking_target;

const cardOn = (page: Page, id: string): Locator =>
  page.locator(`.card[data-instance-id="${id}"]`).first();

test.describe("#3898 blocking on the board", () => {
  test("click your creature, then the attacker", async ({ browser, request }) => {
    test.slow();
    let first: JoinedPlayer | undefined;
    let second: JoinedPlayer | undefined;
    let admin: AdminClient | undefined;
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Block on board 3898 ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing on fresh game");
      first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One");
      second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two");
      await uploadDeckAs(request, adminToken, game.id, first.playerID, makeDeck());
      await uploadDeckAs(request, adminToken, game.id, second.playerID, makeDeck());
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
      const a = admin;
      await keepAllHands(a);
      await a.waitFor((v) => v.state === "active", "game state active");
      await a.waitFor(
        (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the attacker's precombat main",
        15_000,
      );

      const snap = a.snapshot();
      const activeSeat = snap.turn?.active_seat ?? 0;
      const attacker = snap.seats[activeSeat]?.id === first.playerID ? first : second;
      const defender = attacker === first ? second : first;
      const defenderSeat = snap.seats.findIndex((s) => s.id === defender.playerID);

      for (const name of [DRAKE, BRUTE, PETITIONERS, PETITIONERS]) {
        await adminMoveByName(a, attacker.playerID, name, "library", "battlefield");
      }
      for (const name of [BEAR, OTHER_BEAR]) {
        await adminMoveByName(a, defender.playerID, name, "library", "battlefield");
      }
      await a.waitFor(
        (v) =>
          ofName(v, attacker.playerID, PETITIONERS).length === 2 &&
          [DRAKE, BRUTE].every((n) => ofName(v, attacker.playerID, n).length === 1) &&
          [BEAR, OTHER_BEAR].every((n) => ofName(v, defender.playerID, n).length === 1),
        "every creature is out",
      );
      // Clear summoning sickness, as the #318 and #2956 specs do.
      await a.sendActionAsPlayer(attacker.playerID, "untap_all", {});
      const v0 = a.snapshot();
      const drake = ofName(v0, attacker.playerID, DRAKE)[0]!.instance_id;
      const brute = ofName(v0, attacker.playerID, BRUTE)[0]!.instance_id;
      const pets = ofName(v0, attacker.playerID, PETITIONERS).map((c) => c.instance_id);
      const bear = ofName(v0, defender.playerID, BEAR)[0]!.instance_id;
      const otherBear = ofName(v0, defender.playerID, OTHER_BEAR)[0]!.instance_id;

      await a.sendAction("advance_step", {});
      await a.waitFor(
        (v) => v.turn?.step === "declare_attackers",
        "cursor on declare_attackers",
        20_000,
      );
      for (const id of [drake, brute, ...pets]) {
        await a.sendActionAsPlayer(attacker.playerID, "declare_attacker", {
          attacker: id,
          target: defender.playerID,
        });
      }
      await a.waitFor(
        (v) =>
          [drake, brute, ...pets].every(
            (id) =>
              cards(v).find((c) => c.instance_id === id)?.attacking_target === defender.playerID,
          ),
        "all four attack",
      );

      // Walk to the defender's declaration: whoever holds priority
      // passes. A refused pass raced a browser's own; look again.
      const deadline = Date.now() + 60_000;
      for (;;) {
        const v = a.snapshot();
        const t = v.turn;
        if (
          t?.step === "declare_blockers" &&
          (t.block_pending_seats ?? []).includes(defenderSeat)
        ) {
          break;
        }
        if (Date.now() > deadline) throw new Error("never reached the defender's declaration");
        try {
          if (t && t.priority_holder >= 0) {
            const holder = v.seats[t.priority_holder]?.id;
            if (holder) await a.sendActionAsPlayer(holder, "pass_priority", {});
          }
        } catch {
          // Raced a browser; look again.
        }
        await a
          .waitFor(
            (n) => n.turn?.step !== t?.step || n.turn?.priority_holder !== t?.priority_holder,
            "combat moved on",
            3_000,
          )
          .catch(() => undefined);
      }

      const page = defender.page;
      const dock = page.getByRole("region", { name: "actions", exact: true });
      const blockGroup = dock.getByRole("group", { name: L.declareBlockers });
      await expect(blockGroup).toContainText("Choose blockers, or declare none", {
        timeout: 20_000,
      });

      // 1. Select the bear: the flyer is greyed with the reason, and a
      // click on it declares nothing.
      await cardOn(page, bear).click();
      await expect(dock).toContainText(`Blocking with ${BEAR}`);
      const drakeCard = cardOn(page, drake);
      await expect(drakeCard).toHaveClass(/block-refused/);
      await expect(drakeCard).toHaveAttribute("title", /can't block: .*flying/);
      // The click declares nothing; the dock says why (the touch path).
      await drakeCard.click({ force: true });
      await expect(dock).toContainText("can't block that one");
      await page.waitForTimeout(1_000);
      expect(blockOf(a.snapshot(), bear), "a refused attacker took the click").toBeUndefined();

      // 2. The pile fans out on a click, and the second copy is blocked.
      const drawnPet = (await cardOn(page, pets[0]!).count()) > 0 ? pets[0]! : pets[1]!;
      const hiddenPet = drawnPet === pets[0] ? pets[1]! : pets[0]!;
      await expect(page.locator(`.card[data-instance-id="${hiddenPet}"]`)).toHaveCount(0);
      await cardOn(page, drawnPet).click();
      await expect(cardOn(page, hiddenPet)).toBeVisible();
      await cardOn(page, hiddenPet).click();
      await a.waitFor((v) => blockOf(v, bear) === hiddenPet, "the bear blocks the copy clicked");
      await expect(page.locator("svg.arrows path.block")).toHaveCount(1);

      // 3. Take it back: select the bear, then its attacker again.
      await cardOn(page, bear).click();
      await expect(dock.getByRole("button", { name: L.removeBlock, exact: true })).toBeVisible();
      await cardOn(page, hiddenPet).click();
      await a.waitFor((v) => blockOf(v, bear) === undefined, "the bear's block is taken back");

      // 4. Menace: the bear, the brute, then the other bear.
      await cardOn(page, bear).click();
      await cardOn(page, brute).click();
      await expect(dock).toContainText("needs 2 blockers");
      await expect(page.locator("svg.arrows path.block-pending")).toHaveCount(1);
      expect(
        blockOf(a.snapshot(), bear),
        "nothing is sent before the group is complete",
      ).toBeUndefined();
      await cardOn(page, otherBear).click();
      await a.waitFor(
        (v) => blockOf(v, bear) === brute && blockOf(v, otherBear) === brute,
        "both bears block the brute",
      );
      await expect(page.locator("svg.arrows path.block")).toHaveCount(2);

      // 5. The list fallback, then Done blocking.
      await dock.getByRole("button", { name: L.showAsList, exact: true }).click();
      const list = dock.getByRole("dialog", { name: L.chooseBlockers });
      await expect(list).toBeVisible();
      await expect(list.getByRole("group", { name: `${BEAR} 2/2 blocks` })).toBeVisible();
      await expect(
        list.getByRole("group", { name: `${BEAR} 2/2 blocks` }).getByRole("button", {
          name: /Wind Drake/,
        }),
      ).toBeDisabled();
      await list.getByRole("button", { name: "Back to board" }).click();
      await expect(list).toHaveCount(0);

      await dock.getByRole("button", { name: "Done blocking", exact: true }).click();
      await a.waitFor(
        (v) =>
          v.turn?.step !== "declare_blockers" ||
          (v.turn?.blocks_declared_seats ?? []).includes(defenderSeat),
        "the declaration finished",
        20_000,
      );
      const end = a.snapshot();
      expect(blockOf(end, bear)).toBe(brute);
      expect(blockOf(end, otherBear)).toBe(brute);
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});
