import { expect, test } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  keepAllHands,
  openAdminClient,
  seedHandWithCard,
  type SnapshotView,
} from "./s19-helpers";
import { L } from "../../client/src/lib/labels";

// #4043: Lightning Storm under a response. In the default pile the lower
// card's strip carries a compact bolt button that activates the storm's
// "Discard a land card" ability (CR 113.6), the same as the top card
// would; the caster's name keeps its own row. Checked at 1024x768 (tablet,
// the minimum screen) and a 390px phone.
//
// SHOT_DIR writes screenshots (the before/after pair for the PR).

const STORM = "Lightning Storm";
const BOLT = "Lightning Bolt";
const SIZES = [
  { width: 1024, height: 768 },
  { width: 390, height: 780 },
];

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${STORM}`,
      `1 ${BOLT}`,
      "97 Mountain",
    ].join("\n") + "\n"
  );
}

const stackNames = (v: SnapshotView): string[] =>
  (v.stack_items ?? []).map((s) => s.source_card_id ?? "");

test("#4043 a storm under a response offers its ability on the lower card's strip", async ({
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
      `Stack ability 4043 ${Date.now()}`,
    );
    if (!game.invite_token) throw new Error("invite token missing");
    players.push(
      await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Archmage Thaddeus",
      ),
    );
    players.push(
      await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two"),
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

      const storm = await seedHandWithCard(admin, caster.playerID, STORM);
      const bolt = await seedHandWithCard(admin, caster.playerID, BOLT);
      await admin.sendActionAsPlayer(caster.playerID, "cast_spell", {
        instance_id: storm.instance_id,
        from_zone: "hand",
        targets: [{ kind: "player", id: other.playerID }],
      });
      await admin.waitFor(
        (v) => stackNames(v).includes(storm.instance_id),
        "the storm on the stack",
        10_000,
      );
      await admin.sendActionAsPlayer(caster.playerID, "cast_spell", {
        instance_id: bolt.instance_id,
        from_zone: "hand",
        targets: [{ kind: "player", id: other.playerID }],
      });
      await admin.waitFor(
        (v) => stackNames(v).includes(bolt.instance_id),
        "the response on top",
        10_000,
      );
      const before = (admin.snapshot().stack_items ?? []).length;

      // Any player may activate it, and `other` holds priority once the
      // response is cast.
      const page = other.page;
      await admin.waitFor(
        (v) =>
          v.turn?.priority_holder ===
          v.seats.findIndex((q) => q.id === other.playerID),
        "the other seat to hold priority",
        10_000,
      );
      const shotDir = process.env.SHOT_DIR;
      if (shotDir) mkdirSync(shotDir, { recursive: true });
      const name = L.stackAbility(STORM, "Discard a land card");
      const btn = page.getByRole("button", { name, exact: true });

      for (const size of SIZES) {
        await page.setViewportSize(size);
        await page.waitForTimeout(1500);
        if (shotDir) {
          writeFileSync(
            path.join(shotDir, `stack-${size.width}x${size.height}.png`),
            await page.screenshot(),
          );
        }
        if (process.env.SHOT_ONLY === "1") continue;
        // One bolt button, on the storm's strip, inside the viewport.
        await expect(btn).toHaveCount(1);
        await expect(btn).toBeVisible();
        const box = (await btn.boundingBox())!;
        expect(box.x).toBeGreaterThanOrEqual(0);
        expect(box.x + box.width).toBeLessThanOrEqual(size.width);
        // Nothing covers it: a tap would reach the button.
        const top = await btn.evaluate((e) => {
          const r = e.getBoundingClientRect();
          return (
            document
              .elementFromPoint(r.x + r.width / 2, r.y + r.height / 2)
              ?.closest("button") === e
          );
        });
        expect(top).toBe(true);
        // #4150: and a real tap lands on it (a phone used to leave the
        // strip zero-wide, the button in the DOM under the life chip).
        const hit = await btn.evaluate((e) => {
          const r = e.getBoundingClientRect();
          const at = document.elementFromPoint(
            r.x + r.width / 2,
            r.y + r.height / 2,
          );
          return !!at && (at === e || e.contains(at));
        });
        expect(hit).toBe(true);
        // The caster's name on the top card is not cut to an initial.
        const who = page.locator(".who-row .caster").first();
        if (await who.count()) {
          const t = (await who.textContent())?.trim() ?? "";
          expect(t.length).toBeGreaterThan(2);
        }
      }
      if (process.env.SHOT_ONLY === "1") return;

      await page.setViewportSize(SIZES[0]);
      await btn.click();
      const pick = page.locator(".prompt-opt").first();
      if (await pick.isVisible({ timeout: 3_000 }).catch(() => false)) {
        await pick.click();
        await page
          .getByRole("button", { name: "Discard", exact: true })
          .click();
      }
      await admin.waitFor(
        (v) => (v.stack_items ?? []).length === before + 1,
        "the storm's ability on the stack",
        10_000,
      );
    } finally {
      admin.close();
    }
  } finally {
    await closeAll(...players);
  }
});
