import { expect, test, type Page } from "@playwright/test";
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

// #2956 / ADR 0147: combat damage without the typing.
//
// A 6/6 trampler (Colossal Dreadmaw) is double-blocked by two 2/2s. Its
// damage covers lethal for both, so the split is fixed: 2, 2 and 2 over.
//
//   1. With "Auto-assign combat damage" off, the sheet opens on that
//      split (not 2 / 2 / 0), each blocker carries a − / + stepper on
//      the board, a tick there moves the sheet's numbers, and Deal
//      damage sends it.
//   2. With the setting at its default (on), nothing is asked: the
//      split is sent for the attacker and the dock says so.
//
// The board is staged through the admin WS, as the #2614 spec does.

const TRAMPLER = "Colossal Dreadmaw";
const BEAR = "Grizzly Bears";
const OTHER_BEAR = "Balduvian Bears";
const SETTINGS_VERSION = 27;

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${TRAMPLER}`,
      `1 ${BEAR}`,
      `1 ${OTHER_BEAR}`,
      "96 Forest",
    ].join("\n") + "\n"
  );
}

function lifeOf(v: SnapshotView, playerID: string): number {
  const seat = v.seats.find((s) => s.id === playerID) as unknown as
    | { life?: number }
    | undefined;
  return seat?.life ?? NaN;
}

interface Staged {
  first: JoinedPlayer;
  second: JoinedPlayer;
  admin: AdminClient;
  attacker: JoinedPlayer;
  defender: JoinedPlayer;
  bearID: string;
  lifeBefore: number;
}

// stage seats two players with `gameplay`, puts the trampler on the
// attacker's side and both bears on the defender's, attacks, and
// double-blocks. It returns once combat has reached its damage step
// (or the damage has already landed, when the browser answered).
async function stage(
  browser: import("@playwright/test").Browser,
  request: import("@playwright/test").APIRequestContext,
  label: string,
  gameplay: Record<string, unknown>,
  holder: { first?: JoinedPlayer; second?: JoinedPlayer; admin?: AdminClient },
): Promise<Staged> {
  const adminToken = await adminLogin(request);
  const game = await createGame(request, adminToken, `${label} ${Date.now()}`);
  if (!game.invite_token) throw new Error("invite token missing on fresh game");
  const seed = { __version: SETTINGS_VERSION, gameplay };
  const first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One", seed);
  holder.first = first;
  const second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two", seed);
  holder.second = second;
  for (const p of [first, second]) await watchForSheet(p.page);
  await uploadDeckAs(request, adminToken, game.id, first.playerID, makeDeck());
  await uploadDeckAs(request, adminToken, game.id, second.playerID, makeDeck());
  await startGameAs(request, adminToken, game.id);

  const admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
  holder.admin = admin;
  await keepAllHands(admin);
  await admin.waitFor((v) => v.state === "active", "game state active");
  await admin.waitFor(
    (v) =>
      v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
    "cursor settled on the attacker's precombat main",
    15_000,
  );

  const snap = admin.snapshot();
  const activeSeat = snap.turn?.active_seat ?? 0;
  const attacker = snap.seats[activeSeat]?.id === first.playerID ? first : second;
  const defender = attacker === first ? second : first;
  const defenderSeat = snap.seats.findIndex((s) => s.id === defender.playerID);

  await adminMoveByName(admin, attacker.playerID, TRAMPLER, "library", "battlefield");
  await adminMoveByName(admin, defender.playerID, BEAR, "library", "battlefield");
  await adminMoveByName(admin, defender.playerID, OTHER_BEAR, "library", "battlefield");
  await admin.waitFor(
    (v) => [TRAMPLER, BEAR, OTHER_BEAR].every((n) => findCardOnBattlefield(v, n)),
    "all three creatures are out",
  );
  // Clear summoning sickness, as the #318 spec does (CR 302.6).
  await admin.sendActionAsPlayer(attacker.playerID, "untap_all", {});
  const v0 = admin.snapshot();
  const wurm = findCardOnBattlefield(v0, TRAMPLER)!;
  const bear = findCardOnBattlefield(v0, BEAR)!;
  const otherBear = findCardOnBattlefield(v0, OTHER_BEAR)!;

  await admin.sendAction("advance_step", {});
  await admin.waitFor(
    (v) => v.turn?.step === "declare_attackers",
    "cursor on declare_attackers",
    20_000,
  );
  await admin.sendActionAsPlayer(attacker.playerID, "declare_attacker", {
    attacker: wurm.instance_id,
    target: defender.playerID,
  });
  await admin.waitFor(
    (v) =>
      v.battlefield.cards.some(
        (c) => c.instance_id === wurm.instance_id && c.attacking_target === defender.playerID,
      ),
    "the trampler attacks",
  );

  // Walk combat to the damage prompt: the defender double-blocks, and
  // whoever holds priority passes. The damage prompt itself is left to
  // the attacker's browser. A refused action raced a browser's own
  // auto-pass, so look again.
  const lifeBefore = lifeOf(admin.snapshot(), defender.playerID);
  let blocked = false;
  const deadline = Date.now() + 90_000;
  for (;;) {
    const v = admin.snapshot();
    const t = v.turn;
    if ((v.pending_choices ?? []).some((c) => c.kind === "damage_assignment")) break;
    if (lifeOf(v, defender.playerID) !== lifeBefore) break;
    if (Date.now() > deadline) throw new Error("combat never reached its damage");
    try {
      if (
        t?.step === "declare_blockers" &&
        (t.block_pending_seats ?? []).includes(defenderSeat)
      ) {
        if (!blocked) {
          for (const b of [bear, otherBear]) {
            await admin.sendActionAsPlayer(defender.playerID, "declare_blocker", {
              blocker: b.instance_id,
              attacker: wurm.instance_id,
            });
          }
          blocked = true;
        }
        await admin.sendActionAsPlayer(defender.playerID, "finish_blocks", {});
      } else if (t && t.priority_holder >= 0) {
        const holder = v.seats[t.priority_holder]?.id;
        if (holder) await admin.sendActionAsPlayer(holder, "pass_priority", {});
      }
    } catch {
      // Raced a browser; look again.
    }
    await admin
      .waitFor(
        (n) =>
          lifeOf(n, defender.playerID) !== lifeBefore ||
          n.turn?.step !== t?.step ||
          n.turn?.priority_holder !== t?.priority_holder ||
          (n.pending_choices ?? []).length !== (v.pending_choices ?? []).length,
        "combat moved on",
        3_000,
      )
      .catch(() => undefined);
  }
  return {
    first,
    second,
    admin,
    attacker,
    defender,
    bearID: bear.instance_id,
    lifeBefore,
  };
}

// watchForSheet counts every damage sheet the page ever opens, so one
// that flashes up and is answered between polls is not missed.
async function watchForSheet(page: Page): Promise<void> {
  await page.evaluate(() => {
    const w = window as unknown as { __damageSheets: number };
    w.__damageSheets = 0;
    const sel = '[aria-label="Assign combat damage"]';
    new MutationObserver((records) => {
      for (const r of records) {
        for (const n of r.addedNodes) {
          if (!(n instanceof Element)) continue;
          if (n.matches(sel) || n.querySelector(sel)) w.__damageSheets++;
        }
        if (r.type === "attributes" && r.target instanceof Element && r.target.matches(sel)) {
          w.__damageSheets++;
        }
      }
    }).observe(document.body, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ["aria-label"],
    });
  });
}

async function sheetsSeen(page: Page): Promise<number> {
  return page.evaluate(() => (window as unknown as { __damageSheets: number }).__damageSheets);
}

async function sheetNumbers(page: Page): Promise<number[]> {
  const inputs = page.locator('.dock-sheet input[type="number"]');
  const n = await inputs.count();
  const out: number[] = [];
  for (let i = 0; i < n; i++) out.push(Number(await inputs.nth(i).inputValue()));
  return out;
}

test.describe("#2956 combat damage without the typing", () => {
  test("auto-assign off: the sheet opens on the split, and the blockers tick", async ({
    browser,
    request,
  }, testInfo) => {
    test.slow();
    const holder: { first?: JoinedPlayer; second?: JoinedPlayer; admin?: AdminClient } = {};
    try {
      const s = await stage(
        browser,
        request,
        "Combat damage 2956 sheet",
        { autoAssignCombatDamage: false },
        holder,
      );
      const page = s.attacker.page;
      const sheet = page.getByRole("dialog", { name: "Assign combat damage" });
      await expect(sheet).toBeVisible({ timeout: 15_000 });
      await expect.poll(() => sheetNumbers(page)).toEqual([2, 2, 2]);
      await expect(page.locator('[data-testid="damage-stepper"]')).toHaveCount(2);
      await testInfo.attach("damage sheet and board steppers", {
        body: await page.screenshot(),
        contentType: "image/png",
      });

      // + on the bear on the board takes a point from the trample.
      const onBear = page.locator(`[data-instance-id="${s.bearID}"] [data-testid="damage-stepper"]`);
      await onBear.getByRole("button", { name: `Add 1 damage to ${BEAR}` }).click();
      await expect.poll(() => sheetNumbers(page)).toEqual([3, 2, 1]);
      await expect(onBear).toContainText("3");
      // And back, in the sheet: the board follows it.
      const sheetBody = page.locator(".dock-sheet");
      await sheetBody.getByRole("button", { name: `Remove 1 damage from ${BEAR}` }).click();
      await expect(onBear).toContainText("2");
      await sheetBody.getByRole("button", { name: "Add 1 damage to the defending player" }).click();
      await expect.poll(() => sheetNumbers(page)).toEqual([2, 2, 2]);

      await page.getByRole("button", { name: /^Deal damage/ }).click();
      await s.admin.waitFor(
        (v) => lifeOf(v, s.defender.playerID) === s.lifeBefore - 2,
        "2 trample damage reached the defender",
        15_000,
      );
    } finally {
      holder.admin?.close();
      await closeAll(holder.first ?? null, holder.second ?? null);
    }
  });

  test("auto-assign on (the default): nothing is asked", async ({ browser, request }, testInfo) => {
    test.slow();
    const holder: { first?: JoinedPlayer; second?: JoinedPlayer; admin?: AdminClient } = {};
    try {
      const s = await stage(browser, request, "Combat damage 2956 auto", {}, holder);
      await s.admin.waitFor(
        (v) => lifeOf(v, s.defender.playerID) === s.lifeBefore - 2,
        "the split was sent for the attacker",
        20_000,
      );
      const page = s.attacker.page;
      expect(await sheetsSeen(page), "the damage sheet never opened").toBe(0);
      await testInfo.attach("after the automatic assignment", {
        body: await page.screenshot(),
        contentType: "image/png",
      });
    } finally {
      holder.admin?.close();
      await closeAll(holder.first ?? null, holder.second ?? null);
    }
  });
});
