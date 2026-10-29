import { expect, test } from "@playwright/test";
import { writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  openAdminClient,
  seedHandWithCard,
} from "./s19-helpers";

// #4435: while a player is being chosen as a target, every legal player
// must be reachable however the table is drawn. The action dock and its
// sheets float over the board and the hand fans across its foot, so a
// player's avatar can sit under one of them (your own, under the hand at
// 1366x768 and 1920x1080). The targeting request lists every legal player
// as a button, so the board is never the only way to pick.
//
// Per table layout and size, with Lightning Bolt's prompt open at a
// four-seat table: every legal player has a button in the prompt, whether
// or not their avatar is clear on the board. Pressing one picks that
// player and casts the Bolt.
//
// SHOT_DIR writes the screenshots (the before / after pair for the PR);
// SHOT_ONLY=1 takes them without asserting (the "before" run).

const BOLT = "Lightning Bolt";
const SIZES = [
  { width: 1024, height: 768 },
  { width: 1366, height: 768 },
  { width: 1920, height: 1080 },
];
const NAMES = ["First", "Second", "Third", "Fourth"].map(
  (n) => `Archmage Thaddeus the ${n}`,
);

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${BOLT}`,
      "98 Mountain",
    ].join("\n") + "\n"
  );
}

for (const layout of ["quadrant", "row", "focus"]) {
  test(`#4435 ${layout}: every legal player target is reachable while a target is chosen`, async ({
    browser,
    request,
  }, testInfo) => {
    test.setTimeout(240_000);
    const players: JoinedPlayer[] = [];
    const admins: Awaited<ReturnType<typeof openAdminClient>>[] = [];
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(
        request,
        adminToken,
        `Target reach 4435 ${Date.now()}`,
      );
      if (!game.invite_token) throw new Error("invite token missing");
      for (const name of NAMES) {
        players.push(
          await joinAsPlayer(browser, game.id, game.invite_token, name, {
            display: { tableLayout: layout },
          }),
        );
      }
      for (const p of players) {
        await uploadDeckAs(
          request,
          adminToken,
          game.id,
          p.playerID,
          makeDeck(),
        );
      }
      await startGameAs(request, adminToken, game.id);
      for (const p of players) {
        admins.push(
          await openAdminClient(adminToken, game.id, p.playerID, p.playerID),
        );
      }
      const admin = admins[0];
      for (let i = 0; i < 20; i++) {
        const v = await admin.waitFor(
          (s) =>
            s.mulligans_open !== true ||
            s.seats.some((q) => q.mulligan_turn === true),
          "a seat to decide",
          10_000,
        );
        if (v.mulligans_open !== true) break;
        const who = v.seats.find((q) => q.mulligan_turn === true)!;
        await admins[
          players.findIndex((p) => p.playerID === who.id)
        ].sendActionAsPlayer(who.id, "keep_hand", {});
        await admin.waitFor(
          (s) =>
            s.mulligans_open !== true ||
            s.seats.find((q) => q.mulligan_turn === true)?.id !== who.id,
          "next seat",
          10_000,
        );
      }
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
      const cAdmin = admins[players.indexOf(caster)];
      // A Mountain to pay with, so the Bolt is castable and a click starts the cast.
      await adminMoveByName(
        cAdmin,
        caster.playerID,
        "Mountain",
        "library",
        "battlefield",
      );
      await cAdmin.sendActionAsPlayer(caster.playerID, "untap_all", {});
      const boltID = (await seedHandWithCard(cAdmin, caster.playerID, BOLT))
        .instance_id;

      const page = caster.page;
      await page
        .getByRole("button", { name: "Hide tips" })
        .click({ timeout: 5_000 })
        .catch(() => undefined);
      const dialog = page.getByRole("dialog", { name: /Select target for/ });

      const sizes = SIZES;
      for (const size of sizes) {
        await page.setViewportSize(size);
        await page.waitForTimeout(600);
        // Enter on the focused card casts it, as a click does; a click can land on
        // the dock where the quadrant table's hand runs under it.
        const bolt = page.locator(`[data-instance-id="${boltID}"]`).first();
        await bolt.focus();
        await page.keyboard.press("Enter");
        await expect(dialog).toBeVisible({ timeout: 10_000 });
        await page.waitForTimeout(600);
        const tag = `target-${layout}-${size.width}x${size.height}`;
        const shot = await page.screenshot();
        await testInfo.attach(tag, { body: shot, contentType: "image/png" });
        if (process.env.SHOT_DIR)
          writeFileSync(path.join(process.env.SHOT_DIR, `${tag}.png`), shot);

        // Each legal avatar on the board: is it the top element at its centre?
        const covered = () =>
          page.evaluate(
            () =>
              [
                ...document.querySelectorAll<HTMLElement>(
                  ".identity.cast-targetable",
                ),
              ]
                .filter((id) => id.dataset.seatIdentity)
                .map((id) => {
                  const wrap = id.querySelector<HTMLElement>(".avatar-wrap")!;
                  const r = wrap.getBoundingClientRect();
                  const top = document.elementFromPoint(
                    r.x + r.width / 2,
                    r.y + r.height / 2,
                  );
                  return !!top && wrap.contains(top);
                })
                .filter((clear) => !clear).length,
          );
        const lead = dialog.getByText("Players \u2192", { exact: true });
        const names = NAMES.map((n) =>
          dialog.getByRole("button", { name: `Target ${n}` }),
        );
        const natural = await covered();
        console.log(
          `${tag}: ${natural} legal avatars covered as the table is drawn`,
        );
        if (!process.env.SHOT_ONLY) {
          if (natural === 0) {
            // Nothing is hidden, so the prompt stays as short as it was: the
            // dock's height is what lifts your own avatar at the quadrant table.
            await expect(
              lead,
              `${tag}: no row while every avatar is clear`,
            ).toHaveCount(0);
          } else {
            // Something on the table already hides a legal player.
            for (const b of names)
              await expect(
                b,
                `${tag}: a covered player's button`,
              ).toBeVisible();
          }
        }

        // A dialog sits over one opponent's avatar, as it did for Ian: that
        // player is still one press away, and so is everyone else.
        const victim = players.find((p) => p !== caster)!;
        await page.evaluate((id) => {
          const wrap = document.querySelector<HTMLElement>(
            `.identity.cast-targetable .avatar-wrap[data-seat-id="${id}"]`,
          )!;
          const r = wrap.getBoundingClientRect();
          const d = document.createElement("div");
          d.id = "cover-4435";
          d.setAttribute("role", "dialog");
          d.setAttribute("aria-label", "something in the way");
          d.style.cssText = `position:fixed;z-index:5000;left:${r.left - 12}px;top:${r.top - 12}px;width:${r.width + 24}px;height:${r.height + 24}px;background:#223;border:1px solid #88a;`;
          document.querySelector(".board")!.appendChild(d);
        }, victim.playerID);
        if (process.env.SHOT_DIR) {
          writeFileSync(
            path.join(process.env.SHOT_DIR, `${tag}-covered.png`),
            await page.screenshot(),
          );
        }
        if (!process.env.SHOT_ONLY) {
          expect(
            await covered(),
            `${tag}: the dialog covers one avatar`,
          ).toBeGreaterThanOrEqual(1);
          for (const b of names)
            await expect(b, `${tag}: one press away`).toBeVisible();
          await expect(lead).toBeVisible();
        }
        await page.evaluate(() =>
          document.getElementById("cover-4435")?.remove(),
        );

        const last = size === sizes[sizes.length - 1];
        if (!last || process.env.SHOT_ONLY) {
          await page.keyboard.press("Escape");
          await expect(dialog).toBeHidden();
          continue;
        }
        // The button picks: the Bolt is cast, and the prompt is done.
        await dialog
          .getByRole("button", { name: `Target ${victim.name}` })
          .click();
        await expect(dialog).toBeHidden();
        await admin.waitFor(
          (v) => (v.stack_items ?? []).some((s) => s.source_card_id === boltID),
          "the Bolt on the stack",
          10_000,
        );
      }
    } finally {
      admins.forEach((a) => a.close());
      await closeAll(...players);
    }
  });
}
