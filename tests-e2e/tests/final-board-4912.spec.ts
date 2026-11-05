import { expect, test, type Page } from "@playwright/test";
import { writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  findCardOnBattlefield,
  openAdminClient,
  playerByID,
} from "./s19-helpers";

// #4912: an eliminated player's final board. The last seat puts three
// permanents onto the battlefield and a card into the graveyard, then
// concedes. On the first seat's page, at 2, 3 and 4 seats and at
// 1280x720, 1366x768 and 1920x1080, the tombstone on the departed seat's
// panel carries a Final board button that is on screen and not covered;
// it opens the read-only viewer with the recorded battlefield, and the
// hand tab holds the hand the player left with. A reload opens it the
// same. At two seats the concession ends the game, and the stone and the
// viewer are there on the ended table too.
//
// SHOT_DIR writes the screenshots.

const ON_BOARD = ["Sol Ring", "Llanowar Elves", "Mind Stone"];
const IN_GRAVE = "Rhystic Study";

function makeDeck(): string {
  const named = [...ON_BOARD, IN_GRAVE];
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      ...named.map((n) => `1 ${n}`),
      `${99 - named.length} Plains`,
    ].join("\n") + "\n"
  );
}

const SIZES = [
  { width: 1280, height: 720 },
  { width: 1366, height: 768 },
  { width: 1920, height: 1080 },
];

async function shoot(page: Page, tag: string): Promise<void> {
  const shot = await page.screenshot();
  if (process.env.SHOT_DIR)
    writeFileSync(path.join(process.env.SHOT_DIR, `${tag}.png`), shot);
}

for (const seats of [2, 3, 4]) {
  test(`#4912 ${seats} seats: the departed seat's final board opens from its tombstone`, async ({
    browser,
    request,
  }) => {
    test.setTimeout(300_000);
    const players: JoinedPlayer[] = [];
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(
        request,
        adminToken,
        `Final board 4912 ${seats} ${Date.now()}`,
      );
      if (!game.invite_token) throw new Error("invite token missing");
      for (let i = 0; i < seats; i++) {
        players.push(
          await joinAsPlayer(
            browser,
            game.id,
            game.invite_token,
            `Seat ${i + 1}`,
            // Tips off: a first-use tip can sit over the stone (hints.spec
            // is their subject, not this one).
            { help: { tipsOff: true } },
          ),
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
      const gone = players[seats - 1];
      const admins = await Promise.all(
        players.map((p) =>
          openAdminClient(adminToken, game.id, p.playerID, p.playerID),
        ),
      );
      let handLeft = 0;
      try {
        for (let i = 0; i < 20; i++) {
          const v = await admins[0].waitFor(
            (s) =>
              s.mulligans_open !== true ||
              s.seats.some((q) => q.mulligan_turn === true),
            "a seat to decide",
            10_000,
          );
          if (v.mulligans_open !== true) break;
          const who = v.seats.find((q) => q.mulligan_turn === true)!;
          const idx = players.findIndex((p) => p.playerID === who.id);
          await admins[idx].sendActionAsPlayer(who.id, "keep_hand", {});
          await admins[0].waitFor(
            (s) =>
              s.mulligans_open !== true ||
              s.seats.find((q) => q.mulligan_turn === true)?.id !== who.id,
            "next seat",
            10_000,
          );
        }
        await admins[0].waitFor(
          (v) => v.state === "active",
          "game state active",
        );
        const admin = admins[seats - 1];
        for (const name of ON_BOARD) {
          await adminMoveByName(
            admin,
            gone.playerID,
            name,
            "library",
            "battlefield",
          );
        }
        await adminMoveByName(
          admin,
          gone.playerID,
          IN_GRAVE,
          "library",
          "graveyard",
        );
        const before = await admin.waitFor(
          (v) => ON_BOARD.every((n) => findCardOnBattlefield(v, n)),
          "the board is out",
        );
        handLeft = playerByID(before, gone.playerID).hand.count;
        await admin.sendActionAsPlayer(gone.playerID, "concede", {});
        await admins[0].waitFor(
          (v) =>
            v.seats.some(
              (q) =>
                q.id === gone.playerID &&
                (q as { eliminated?: boolean }).eliminated === true,
            ),
          "the last seat has left",
        );
      } finally {
        admins.forEach((a) => a.close());
      }

      const page = players[0].page;
      const openButton = page.getByRole("button", {
        name: `show ${gone.name}'s final board`,
      });
      const dialog = page.getByRole("dialog", {
        name: `final board of ${gone.name}`,
      });
      // At two seats the concession ended the game: the winner's fanfare
      // is up (again after a reload), and the player puts it away to look
      // at the board.
      const keepLooking = async (): Promise<void> => {
        if (seats !== 2) return;
        const look = page.getByRole("button", {
          name: "Keep looking at the board",
        });
        await expect(look).toBeVisible({ timeout: 10_000 });
        await look.click();
        await expect(look).toHaveCount(0);
      };
      await keepLooking();
      for (const size of SIZES) {
        await page.setViewportSize(size);
        const tag = `${seats}-seats-${size.width}x${size.height}`;
        await expect(openButton).toBeVisible({ timeout: 30_000 });
        // Let the stone's drop and any cue settle.
        await page.waitForTimeout(3_000);
        // On screen, and the button itself is what a click there reaches.
        const reach = await openButton.evaluate((el) => {
          const r = el.getBoundingClientRect();
          const x = r.left + r.width / 2;
          const y = r.top + r.height / 2;
          const hit = document.elementFromPoint(x, y);
          return {
            inside:
              r.left >= 0 &&
              r.top >= 0 &&
              r.right <= innerWidth &&
              r.bottom <= innerHeight,
            hit: hit === el || el.contains(hit),
            box: { left: r.left, top: r.top, width: r.width, height: r.height },
            hitDesc: hit ? `${hit.tagName}.${hit.className}` : "none",
          };
        });
        expect(
          reach.inside,
          `${tag}: button on screen ${JSON.stringify(reach.box)}`,
        ).toBe(true);
        expect(reach.hit, `${tag}: button covered by ${reach.hitDesc}`).toBe(
          true,
        );
        await shoot(page, `${tag}-tombstone`);

        await openButton.click();
        await expect(dialog).toBeVisible();
        await expect(dialog.locator("[data-final-card]")).toHaveCount(
          ON_BOARD.length,
        );
        await shoot(page, `${tag}-battlefield`);
        await dialog.locator('[data-final-zone="hand"]').click();
        await expect(dialog.locator("[data-final-card]")).toHaveCount(handLeft);
        await dialog.locator('[data-final-zone="graveyard"]').click();
        await expect(dialog.locator("[data-final-card]")).toHaveCount(1);
        await dialog.locator('[data-final-zone="command"]').click();
        await expect(dialog.locator("[data-final-card]")).toHaveCount(1);
        await page.keyboard.press("Escape");
        await expect(dialog).toHaveCount(0);
      }

      // A reload has no frame from before the departure, and the viewer
      // still has the record.
      await page.reload();
      await keepLooking();
      await expect(openButton).toBeVisible({ timeout: 30_000 });
      await openButton.click();
      await expect(dialog).toBeVisible();
      await expect(dialog.locator("[data-final-card]")).toHaveCount(
        ON_BOARD.length,
      );
    } finally {
      await closeAll(...players);
    }
  });
}
