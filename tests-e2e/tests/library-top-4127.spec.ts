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
  returnToLibrary,
  seedHandWithCard,
} from "./s19-helpers";

// #4127: playing from the top of your library (Bolas's Citadel). The
// library pile was a small tile whose "cast" pill hung over its bottom
// edge at 8.5px, under the dev dock's spawn and replay panel in the
// corner; a player could not see the card they were being offered or
// find the button. With a playable top card the pile now shows the card
// larger and the cast pill sits on it, clear of the dock.
//
// SHOT_DIR writes screenshots (the before/after pair for the PR).
// SHOT_ONLY=1 skips the assertions, for taking the "before" pair.

const CITADEL = "Bolas's Citadel";
const TOP = "Sol Ring";
const SIZES = [
  { width: 1280, height: 720 },
  { width: 1024, height: 768 },
  { width: 1920, height: 1080 },
];

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${CITADEL}`,
      `1 ${TOP}`,
      "97 Mountain",
    ].join("\n") + "\n"
  );
}

test("#4127 the top of the library is big enough to read and its cast button is clear", async ({
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
      `Library top 4127 ${Date.now()}`,
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

      // Citadel on the table, a Mountain to pay with, and a Sol Ring on
      // top of the library.
      await adminMoveByName(
        admin,
        caster.playerID,
        CITADEL,
        "library",
        "battlefield",
      );
      await adminMoveByName(
        admin,
        caster.playerID,
        "Mountain",
        "library",
        "battlefield",
      );
      await seedHandWithCard(admin, caster.playerID, TOP);
      await returnToLibrary(admin, caster.playerID, TOP, 1);

      const page = caster.page;
      const shotDir = process.env.SHOT_DIR;
      if (shotDir) mkdirSync(shotDir, { recursive: true });
      const castPill = page.getByRole("button", {
        name: /^cast Sol Ring from the top of the library/,
      });

      for (const size of SIZES) {
        await page.setViewportSize(size);
        await page.waitForTimeout(1500);
        if (shotDir) {
          writeFileSync(
            path.join(shotDir, `library-top-${size.width}x${size.height}.png`),
            await page.screenshot(),
          );
        }
        if (process.env.SHOT_ONLY === "1") continue;
        await expect(castPill).toHaveCount(1);
        await expect(castPill).toBeVisible();
        const box = (await castPill.boundingBox())!;
        // A tap target a thumb can hit, not an 8px sliver.
        expect(box.height).toBeGreaterThanOrEqual(24);
        expect(box.width).toBeGreaterThanOrEqual(44);
        expect(box.x).toBeGreaterThanOrEqual(0);
        expect(box.x + box.width).toBeLessThanOrEqual(size.width);
        expect(box.y + box.height).toBeLessThanOrEqual(size.height);
        // Nothing covers it: the dev dock's spawn and replay panel
        // used to.
        const reachable = await castPill.evaluate((e) => {
          const r = e.getBoundingClientRect();
          const at = document.elementFromPoint(
            r.x + r.width / 2,
            r.y + r.height / 2,
          );
          return !!at && (at === e || e.contains(at));
        });
        expect(reachable).toBe(true);
        // The card on offer is drawn big enough to read (the pile's own
        // thumb is about 28px wide), and sits whole inside the screen,
        // clear of the dock strip at the bottom.
        const peek = page.locator("[data-library-top-peek] .peek-card");
        await expect(peek).toBeVisible();
        const pb = (await peek.boundingBox())!;
        expect(pb.width).toBeGreaterThanOrEqual(90);
        expect(pb.y).toBeGreaterThanOrEqual(0);
        const peekFrame = (await page
          .locator("[data-library-top-peek]")
          .boundingBox())!;
        const pile = (await page
          .locator(
            `[data-pile="library"][data-pile-owner="${caster.playerID}"]`,
          )
          .boundingBox())!;
        expect(peekFrame.y + peekFrame.height).toBeLessThanOrEqual(pile.y);
      }
      if (process.env.SHOT_ONLY === "1") return;

      await page.setViewportSize(SIZES[0]);
      await castPill.click();
      // Citadel's own cost (life for the mana value) is asked first.
      await expect(
        page.getByRole("button", { name: /^Pay life equal to its mana/ }),
      ).toBeVisible();
      await page
        .getByRole("button", { name: /^Cast\b/ })
        .last()
        .click();
      await admin.waitFor(
        (v) => (v.stack_items ?? []).length >= 1,
        "the Sol Ring cast from the library",
        10_000,
      );
    } finally {
      admin.close();
    }
  } finally {
    await closeAll(...players);
  }
});
