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

// #2614 / ADR 0134: combat you can watch.
//
// When combat damage lands, the attacker's art-only copy lunges at what
// it hit and snaps back, while its live tile hides for the flight. This
// guards the two things only a real browser shows:
//
//   1. with combat motion on, a strike copy (`[data-strike-copy]`) is
//      drawn during combat damage, and within 2 s it is gone and the
//      attacker's tile is visible again;
//   2. with `animations.combat` off, no copy is ever drawn, and the life
//      total still changes.
//
// The setting is seeded into each browser's settings blob rather than
// toggled through the Settings row, so no label is selected on.
//
// The board is staged through the admin WS (the #318 spec's machinery):
// one vanilla bear on the attacker's side, attacking unblocked.

const BEAR = "Grizzly Bears";
const SETTINGS_VERSION = 21;

function makeBearDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${BEAR}`,
      "98 Forest",
    ].join("\n") + "\n"
  );
}

function lifeOf(v: SnapshotView, playerID: string): number {
  const seat = v.seats.find((s) => s.id === playerID) as unknown as
    { life?: number } | undefined;
  return seat?.life ?? NaN;
}

// Count every strike copy the page ever mounts, so a copy that lives
// half a second is not missed between polls.
async function watchStrikes(page: Page): Promise<void> {
  await page.evaluate(() => {
    const w = window as unknown as { __strikes: number };
    w.__strikes = 0;
    new MutationObserver((records) => {
      for (const r of records) {
        for (const n of r.addedNodes) {
          if (!(n instanceof Element)) continue;
          if (n.matches("[data-strike-copy]")) w.__strikes++;
          else w.__strikes += n.querySelectorAll("[data-strike-copy]").length;
        }
      }
    }).observe(document.body, { childList: true, subtree: true });
  });
}

async function strikesSeen(page: Page): Promise<number> {
  return page.evaluate(
    () => (window as unknown as { __strikes: number }).__strikes,
  );
}

test.describe("#2614 combat you can watch", () => {
  test("an unblocked attacker lunges; with combat motion off, nothing moves", async ({
    browser,
    request,
  }, testInfo) => {
    test.slow();

    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(
        request,
        adminToken,
        `Combat strike 2614 ${Date.now()}`,
      );
      if (!game.invite_token)
        throw new Error("invite token missing on fresh game");

      // Seat One watches at half speed (a 1 s strike, easy to catch);
      // Seat Two has combat motion off. Which of them attacks is the
      // opening roll's call, so the assertions below follow the roles.
      first = await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat One",
        {
          __version: SETTINGS_VERSION,
          animations: { speed: 2 },
        },
      );
      second = await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat Two",
        {
          __version: SETTINGS_VERSION,
          animations: { combat: false },
        },
      );
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        first.playerID,
        makeBearDeck(),
      );
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        second.playerID,
        makeBearDeck(),
      );
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(
        adminToken,
        game.id,
        first.playerID,
        second.playerID,
      );
      await keepAllHands(admin);
      await admin.waitFor((v) => v.state === "active", "game state active");
      await admin.waitFor(
        (v) =>
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the attacker's precombat main",
        15_000,
      );

      const snap = admin.snapshot();
      const activeSeat = snap.turn?.active_seat ?? 0;
      const attackerID = snap.seats[activeSeat]?.id;
      const attacker = attackerID === first.playerID ? first : second;
      const defender = attacker === first ? second : first;
      const defenderSeat = snap.seats.findIndex(
        (s) => s.id === defender.playerID,
      );

      await adminMoveByName(
        admin,
        attacker.playerID,
        BEAR,
        "library",
        "battlefield",
      );
      await admin.waitFor(
        (v) => findCardOnBattlefield(v, BEAR) !== null,
        "the bear is out",
      );
      // Clear summoning sickness, as the #318 spec does (CR 302.6).
      await admin.sendActionAsPlayer(attacker.playerID, "untap_all", {});
      const bear = findCardOnBattlefield(admin.snapshot(), BEAR);
      if (!bear) throw new Error(`${BEAR} missing from the battlefield`);

      await admin.sendAction("advance_step", {});
      await admin.waitFor(
        (v) => v.turn?.step === "declare_attackers",
        "cursor on declare_attackers",
        20_000,
      );
      await admin.sendActionAsPlayer(attacker.playerID, "declare_attacker", {
        attacker: bear.instance_id,
        target: defender.playerID,
      });
      await admin.waitFor(
        (v) =>
          v.battlefield.cards.some(
            (c) =>
              c.instance_id === bear.instance_id &&
              c.attacking_target === defender.playerID,
          ),
        "the bear attacks the defender",
      );

      await watchStrikes(first.page);
      await watchStrikes(second.page);
      const strikeCopy = first.page.locator("[data-strike-copy]");
      // Started before the damage, so the copy cannot come and go unseen.
      const copyMounted = first.page
        .waitForSelector("[data-strike-copy]", {
          state: "attached",
          timeout: 60_000,
        })
        .then(
          () => true,
          () => false,
        );

      // Walk combat to its damage: the defender declares no blocks, and
      // whoever holds priority passes. A browser's own auto-pass may get
      // there first, so a refused pass is fine.
      const lifeBefore = lifeOf(admin.snapshot(), defender.playerID);
      const deadline = Date.now() + 60_000;
      while (lifeOf(admin.snapshot(), defender.playerID) === lifeBefore) {
        if (Date.now() > deadline)
          throw new Error("combat damage never landed");
        const v = admin.snapshot();
        const t = v.turn;
        try {
          if (
            t?.step === "declare_blockers" &&
            (t.block_pending_seats ?? []).includes(defenderSeat)
          ) {
            await admin.sendActionAsPlayer(
              defender.playerID,
              "finish_blocks",
              {},
            );
          } else if (t && t.priority_holder >= 0) {
            const holder = v.seats[t.priority_holder]?.id;
            if (holder)
              await admin.sendActionAsPlayer(holder, "pass_priority", {});
          }
        } catch {
          // Raced a browser's auto-pass; look again.
        }
        await admin
          .waitFor(
            (n) =>
              lifeOf(n, defender.playerID) !== lifeBefore ||
              n.turn?.step !== t?.step ||
              n.turn?.priority_holder !== t?.priority_holder,
            "combat moved on",
            3_000,
          )
          .catch(() => undefined);
      }
      expect(lifeOf(admin.snapshot(), defender.playerID)).toBe(lifeBefore - 2);

      // --- 1: Seat One (motion on) sees the strike ------------------
      expect(
        await copyMounted,
        "a strike copy is drawn during combat damage",
      ).toBe(true);
      await testInfo.attach("strike in flight (Seat One)", {
        body: await first.page.screenshot(),
        contentType: "image/png",
      });
      await expect(strikeCopy, "the copy is gone within 2 s").toHaveCount(0, {
        timeout: 2_500,
      });
      expect(await strikesSeen(first.page)).toBeGreaterThan(0);
      await expect(
        first.page.locator(`[data-instance-id="${bear.instance_id}"]`).first(),
        "the attacker's tile is visible again",
      ).toBeVisible();

      // --- 2: Seat Two (combat off) sees no copy, and the life change -
      await expect(
        second.page.locator(`[data-seat-id="${defender.playerID}"]`).first(),
      ).toHaveAttribute("aria-label", new RegExp(`${lifeBefore - 2} life`), {
        timeout: 10_000,
      });
      await second.page.waitForTimeout(1_500);
      expect(
        await strikesSeen(second.page),
        "no strike copy with combat motion off",
      ).toBe(0);
      await testInfo.attach("after combat (Seat Two, motion off)", {
        body: await second.page.screenshot(),
        contentType: "image/png",
      });
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });

  // ADR 0134 PR 2: first strike, a double block, trample and deaths in
  // one combat. The attacker swings with a first striker (unblocked) and
  // a 6/6 trampler; the defender double-blocks the trampler with two
  // 2/2s. So:
  //   - the first-strike beat lunges the knight before the regular beat
  //     lunges the trampler (CR 510.4);
  //   - the trampler lunges ONCE, at its blockers' centroid (CR 510.1c);
  //   - its 2 excess to the player draws a streak (CR 702.19b);
  //   - both blockers die, so each is drawn as a dying copy that
  //     crumbles into shards (CR 704.3).
  test("first strike, a double block, trample and deaths", async ({
    browser,
    request,
  }, testInfo) => {
    test.slow();

    let first: JoinedPlayer | null = null;
    let second: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;

    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(
        request,
        adminToken,
        `Combat strike 2614 b ${Date.now()}`,
      );
      if (!game.invite_token)
        throw new Error("invite token missing on fresh game");

      // Both seats watch at half speed, so every phase is twice as long
      // and easier to catch in a screenshot.
      const watch = { __version: SETTINGS_VERSION, animations: { speed: 2 } };
      first = await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat One",
        watch,
      );
      second = await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat Two",
        watch,
      );
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        first.playerID,
        makeMeleeDeck(),
      );
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        second.playerID,
        makeMeleeDeck(),
      );
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(
        adminToken,
        game.id,
        first.playerID,
        second.playerID,
      );
      await keepAllHands(admin);
      await admin.waitFor((v) => v.state === "active", "game state active");
      await admin.waitFor(
        (v) =>
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the attacker's precombat main",
        15_000,
      );

      const snap = admin.snapshot();
      const activeSeat = snap.turn?.active_seat ?? 0;
      const attackerID = snap.seats[activeSeat]?.id;
      const attacker = attackerID === first.playerID ? first : second;
      const defender = attacker === first ? second : first;
      const defenderSeat = snap.seats.findIndex(
        (s) => s.id === defender.playerID,
      );

      await adminMoveByName(
        admin,
        attacker.playerID,
        TRAMPLER,
        "library",
        "battlefield",
      );
      await adminMoveByName(
        admin,
        attacker.playerID,
        KNIGHT,
        "library",
        "battlefield",
      );
      await adminMoveByName(
        admin,
        defender.playerID,
        BEAR,
        "library",
        "battlefield",
      );
      await adminMoveByName(
        admin,
        defender.playerID,
        OTHER_BEAR,
        "library",
        "battlefield",
      );
      await admin.waitFor(
        (v) =>
          [TRAMPLER, KNIGHT, BEAR, OTHER_BEAR].every((n) =>
            findCardOnBattlefield(v, n),
          ),
        "all four creatures are out",
      );
      // Clear summoning sickness, as the #318 spec does (CR 302.6).
      await admin.sendActionAsPlayer(attacker.playerID, "untap_all", {});
      const v0 = admin.snapshot();
      const wurm = findCardOnBattlefield(v0, TRAMPLER)!;
      const knight = findCardOnBattlefield(v0, KNIGHT)!;
      const bear = findCardOnBattlefield(v0, BEAR)!;
      const otherBear = findCardOnBattlefield(v0, OTHER_BEAR)!;

      await admin.sendAction("advance_step", {});
      await admin.waitFor(
        (v) => v.turn?.step === "declare_attackers",
        "cursor on declare_attackers",
        20_000,
      );
      for (const c of [wurm, knight]) {
        await admin.sendActionAsPlayer(attacker.playerID, "declare_attacker", {
          attacker: c.instance_id,
          target: defender.playerID,
        });
      }
      await admin.waitFor(
        (v) =>
          [wurm, knight].every((c) =>
            v.battlefield.cards.some(
              (x) =>
                x.instance_id === c.instance_id &&
                x.attacking_target === defender.playerID,
            ),
          ),
        "both attack the defender",
      );

      for (const p of [first, second]) await recordStrikes(p.page);

      // Walk combat through: the defender double-blocks the trampler,
      // the attacker divides its damage (2, 2, and 2 to the player), and
      // whoever holds priority passes. A browser's own auto-pass may get
      // there first, so a refused action is fine.
      const lifeBefore = lifeOf(admin.snapshot(), defender.playerID);
      const want = lifeBefore - 4;
      let blocked = false;
      const deadline = Date.now() + 90_000;
      while (lifeOf(admin.snapshot(), defender.playerID) !== want) {
        if (Date.now() > deadline)
          throw new Error("combat damage never fully landed");
        const v = admin.snapshot();
        const t = v.turn;
        const division = (v.pending_choices ?? []).find(
          (c) => c.kind === "damage_assignment",
        );
        try {
          if (
            t?.step === "declare_blockers" &&
            (t.block_pending_seats ?? []).includes(defenderSeat)
          ) {
            if (!blocked) {
              for (const b of [bear, otherBear]) {
                await admin.sendActionAsPlayer(
                  defender.playerID,
                  "declare_blocker",
                  {
                    blocker: b.instance_id,
                    attacker: wurm.instance_id,
                  },
                );
              }
              blocked = true;
            }
            await admin.sendActionAsPlayer(
              defender.playerID,
              "finish_blocks",
              {},
            );
          } else if (division) {
            await admin.sendActionAsPlayer(
              attacker.playerID,
              "resolve_choice",
              {
                choice_id: division.id,
                assignments: [
                  { blocker_id: bear.instance_id, amount: 2 },
                  { blocker_id: otherBear.instance_id, amount: 2 },
                ],
                trample_to_player: 2,
              },
            );
          } else if (t && t.priority_holder >= 0) {
            const holder = v.seats[t.priority_holder]?.id;
            if (holder)
              await admin.sendActionAsPlayer(holder, "pass_priority", {});
          }
        } catch {
          // Raced a browser's auto-pass or prompt; look again.
        }
        await admin
          .waitFor(
            (n) =>
              lifeOf(n, defender.playerID) === want ||
              n.turn?.step !== t?.step ||
              n.turn?.priority_holder !== t?.priority_holder ||
              (n.pending_choices ?? []).length !==
                (v.pending_choices ?? []).length,
            "combat moved on",
            3_000,
          )
          .catch(() => undefined);
      }

      // The regular beat's frame has landed: catch it mid-flight.
      const page = first.page;
      const streakSeen = page
        .waitForSelector("[data-strike-streak]", {
          state: "attached",
          timeout: 3_000,
        })
        .then(
          () => true,
          () => false,
        );
      if (await streakSeen) {
        await testInfo.attach("trample streak (Seat One)", {
          body: await page.screenshot(),
          contentType: "image/png",
        });
      }
      const shardsSeen = await page
        .waitForSelector("[data-strike-shard]", {
          state: "attached",
          timeout: 3_000,
        })
        .then(
          () => true,
          () => false,
        );
      if (shardsSeen) {
        await testInfo.attach("blockers crumbling (Seat One)", {
          body: await page.screenshot(),
          contentType: "image/png",
        });
      }

      // Everything is gone well inside the slowest beat at speed 2.
      await expect(page.locator("[data-strike-copy]")).toHaveCount(0, {
        timeout: 6_000,
      });
      await expect(page.locator("[data-strike-streak]")).toHaveCount(0, {
        timeout: 6_000,
      });

      const after = admin.snapshot();
      expect(
        findCardOnBattlefield(after, BEAR),
        "the first blocker died",
      ).toBeNull();
      expect(
        findCardOnBattlefield(after, OTHER_BEAR),
        "the second blocker died",
      ).toBeNull();
      expect(
        findCardOnBattlefield(after, TRAMPLER),
        "the trampler lived",
      ).not.toBeNull();

      for (const p of [first, second]) {
        const rec = await strikeRecord(p.page);
        const who = p === first ? "Seat One" : "Seat Two";
        const lunge = (id: string) =>
          rec.copies
            .filter((c) => c.id === id && c.mode === "lunge")
            .map((c) => c.t);
        // One lunge each: the trampler does not lunge once per blocker.
        expect(
          lunge(knight.instance_id),
          `${who}: the knight lunges once`,
        ).toHaveLength(1);
        expect(
          lunge(wurm.instance_id),
          `${who}: the trampler lunges once`,
        ).toHaveLength(1);
        // First strike first (CR 510.4).
        expect(
          lunge(knight.instance_id)[0],
          `${who}: first strike plays first`,
        ).toBeLessThan(lunge(wurm.instance_id)[0]);
        // Both blockers die after the hit, drawn from the cache.
        const dying = rec.copies
          .filter((c) => c.mode === "die")
          .map((c) => c.id);
        expect(dying.sort(), `${who}: both blockers die`).toEqual(
          [bear.instance_id, otherBear.instance_id].sort(),
        );
        expect(rec.streaks, `${who}: trample's excess draws a streak`).toBe(1);
        expect(rec.shards, `${who}: the dead crumble`).toBeGreaterThan(0);
      }
      await expect(
        page.locator(`[data-instance-id="${wurm.instance_id}"]`).first(),
        "the trampler's tile is visible again",
      ).toBeVisible();
      await testInfo.attach("after combat (Seat One)", {
        body: await page.screenshot(),
        contentType: "image/png",
      });
    } finally {
      admin?.close();
      await closeAll(first, second);
    }
  });
});

const TRAMPLER = "Colossal Dreadmaw";
const KNIGHT = "Youthful Knight";
const OTHER_BEAR = "Balduvian Bears";

function makeMeleeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${TRAMPLER}`,
      `1 ${KNIGHT}`,
      `1 ${BEAR}`,
      `1 ${OTHER_BEAR}`,
      "95 Forest",
    ].join("\n") + "\n"
  );
}

interface StrikeRecord {
  copies: { id: string; mode: string; t: number }[];
  streaks: number;
  shards: number;
}

// Record every strike copy (with its mode and when it mounted), every
// streak and every crumble shard the page mounts, so nothing that lives
// half a second is missed between polls.
async function recordStrikes(page: Page): Promise<void> {
  await page.evaluate(() => {
    const rec = { copies: [], streaks: 0, shards: 0 } as StrikeRecord;
    (window as unknown as { __strikeRecord: StrikeRecord }).__strikeRecord =
      rec;
    const see = (el: Element) => {
      if (el.matches("[data-strike-copy]")) {
        rec.copies.push({
          id: el.getAttribute("data-strike-copy") ?? "",
          mode: el.getAttribute("data-strike-mode") ?? "",
          t: performance.now(),
        });
      }
      if (el.matches("[data-strike-streak]")) rec.streaks++;
      if (el.matches("[data-strike-shard]")) rec.shards++;
    };
    new MutationObserver((records) => {
      for (const r of records) {
        for (const n of r.addedNodes) {
          if (!(n instanceof Element)) continue;
          see(n);
          for (const el of n.querySelectorAll(
            "[data-strike-copy], [data-strike-streak], [data-strike-shard]",
          )) {
            see(el);
          }
        }
      }
    }).observe(document.body, { childList: true, subtree: true });
  });
}

async function strikeRecord(page: Page): Promise<StrikeRecord> {
  return page.evaluate(
    () =>
      (window as unknown as { __strikeRecord: StrikeRecord }).__strikeRecord,
  );
}
