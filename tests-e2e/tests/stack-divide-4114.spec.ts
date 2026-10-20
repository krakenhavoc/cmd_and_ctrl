import { expect, test } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  keepAllHands,
  openAdminClient,
  seedHandWithCard,
} from "./s19-helpers";

// #4114: a divided effect says what each target gets. Arc Lightning
// deals 3 damage divided among up to three targets; with two players
// targeted, 2 and 1, the stack's row for it names each with its share
// ("Seat Two (2 damage)"), and each player's avatar carries a badge with
// the same share, so the table reads who takes how much without opening
// anything. The server has always sent the division as the item's
// `distribution`; the client used to drop it.
//
// SHOT_DIR writes screenshots (the before/after pair for the PR).
// SHOT_ONLY=1 skips the assertions, for taking the "before" pair.

const SPELL = "Arc Lightning";
// Careful pass mode: nobody passes for the seat, so the item stays on the
// stack for the screenshot and the checks (the stack hold has its own
// spec).
const CAREFUL = { gameplay: { passMode: "careful" } };
const SIZES = [
  { width: 1280, height: 720 },
  { width: 1024, height: 768 },
];

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${SPELL}`,
      "98 Mountain",
    ].join("\n") + "\n"
  );
}

test("#4114 the stack says how much of a divided effect each target gets", async ({
  browser,
  request,
}) => {
  test.setTimeout(240_000);
  const players: JoinedPlayer[] = [];
  try {
    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Stack divide 4114 ${Date.now()}`,
    );
    if (!game.invite_token) throw new Error("invite token missing");
    players.push(
      await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Archmage Thaddeus",
        CAREFUL,
      ),
    );
    players.push(
      await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat Two",
        CAREFUL,
      ),
    );
    for (const p of players)
      await uploadDeckAs(request, adminToken, game.id, p.playerID, makeDeck());
    await startGameAs(request, adminToken, game.id);
    const admin = await openAdminClient(
      adminToken,
      game.id,
      players[0].playerID,
      players[1].playerID,
    );
    try {
      await keepAllHands(admin);
      await admin.waitFor(
        (v) =>
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === v.turn?.active_seat,
        "the caster's precombat main",
        20_000,
      );
      const snap = admin.snapshot();
      const casterID = snap.seats[snap.turn?.active_seat ?? 0]?.id;
      const caster = players.find((p) => p.playerID === casterID)!;
      const other = players.find((p) => p !== caster)!;

      for (let i = 0; i < 3; i++) {
        await adminMoveByName(
          admin,
          caster.playerID,
          "Mountain",
          "library",
          "battlefield",
        );
      }
      const spell = await seedHandWithCard(admin, caster.playerID, SPELL);
      await admin.sendActionAsPlayer(caster.playerID, "cast_spell", {
        instance_id: spell.instance_id,
        from_zone: "hand",
        targets: [
          { kind: "player", id: other.playerID },
          { kind: "player", id: caster.playerID },
        ],
        distribution: { [other.playerID]: 2, [caster.playerID]: 1 },
        hold_priority: true,
        strict: true,
        auto_tap: true,
      });
      await admin.waitFor(
        (v) =>
          (v.stack_items ?? []).some(
            (s) => s.source_card_id === spell.instance_id,
          ),
        "Arc Lightning on the stack",
        10_000,
      );

      const shotDir = process.env.SHOT_DIR;
      if (shotDir) mkdirSync(shotDir, { recursive: true });
      const page = other.page;
      for (const size of SIZES) {
        await page.setViewportSize(size);
        await page.waitForTimeout(1500);
        if (shotDir) {
          writeFileSync(
            path.join(shotDir, `stack-divide-${size.width}x${size.height}.png`),
            await page.screenshot(),
          );
        }
        if (process.env.SHOT_ONLY === "1") continue;

        // The stack's row names each target with its share.
        const lane = page.locator(
          '[data-stack-item-id], [aria-label*="stack" i]',
        );
        await expect(lane.first()).toBeVisible();
        const text = (await page.locator("body").innerText()).replace(
          /\s+/g,
          " ",
        );
        expect(text).toContain("(2 damage)");
        expect(text).toContain("(1 damage)");

        // And each player's avatar carries its share as a badge, whole
        // inside the screen.
        const badges = page.locator("[data-stack-share]");
        await expect(badges).toHaveCount(2);
        await expect(page.locator('[data-stack-share="2 dmg"]')).toBeVisible();
        await expect(page.locator('[data-stack-share="1 dmg"]')).toBeVisible();
        for (const b of await badges.all()) {
          const box = (await b.boundingBox())!;
          expect(box.x).toBeGreaterThanOrEqual(0);
          expect(box.y).toBeGreaterThanOrEqual(0);
          expect(box.x + box.width).toBeLessThanOrEqual(size.width);
          expect(box.y + box.height).toBeLessThanOrEqual(size.height);
        }
      }
    } finally {
      admin.close();
    }
  } finally {
    await closeAll(...players);
  }
});
