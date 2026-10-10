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
} from "./s19-helpers";

// #2960: a land with a mana ability and another ability (Desolate
// Lighthouse) opens its menu where the player can see and click it, and
// seven lands in the back row each keep a visible, clickable tile.
//
// Before the fix the menu was absolute inside the card, below it: the
// lands row grew a scrollbar and the menu sat behind the hand, and at
// 1998x716 it could not be seen at all. Now it is a fixed box in the
// board's popover host, above the card (or flipped below), inside the
// viewport. Checked at 1280x720 and 1998x716.

const LANDS = [
  "Steam Vents",
  "Mountain",
  "Mutavault",
  "Shivan Reef",
  "Desolate Lighthouse",
  "Castle Vantress",
  "Island",
];
const LIGHTHOUSE = "Desolate Lighthouse";

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      ...LANDS.filter((n) => n !== "Mountain" && n !== "Island").map((n) => `1 ${n}`),
      "10 Mountain",
      "10 Island",
      "74 Forest",
    ].join("\n") + "\n"
  );
}

// topCardAt names the instance whose card is painted at a point, or
// null when something else (the hand, a scrollbar) is.
async function topCardAt(page: Page, x: number, y: number): Promise<string | null> {
  return page.evaluate(
    ([px, py]) =>
      document.elementFromPoint(px, py)?.closest("[data-instance-id]")?.getAttribute("data-instance-id") ??
      null,
    [x, y],
  );
}

test.describe("#2960 land ability menu and the lands row", () => {
  for (const size of [
    { width: 1280, height: 720 },
    { width: 1998, height: 716 },
  ]) {
    test(`${size.width}x${size.height}: the menu's rows are visible and clickable`, async ({
      browser,
      request,
    }, testInfo) => {
      test.slow();
      const holder: { first?: JoinedPlayer; second?: JoinedPlayer; admin?: AdminClient } = {};
      try {
        const adminToken = await adminLogin(request);
        const game = await createGame(request, adminToken, `Land menu 2960 ${Date.now()}`);
        if (!game.invite_token) throw new Error("invite token missing on fresh game");
        const first = await joinAsPlayer(browser, game.id, game.invite_token, "Seat One");
        holder.first = first;
        const second = await joinAsPlayer(browser, game.id, game.invite_token, "Seat Two");
        holder.second = second;
        await uploadDeckAs(request, adminToken, game.id, first.playerID, makeDeck());
        await uploadDeckAs(request, adminToken, game.id, second.playerID, makeDeck());
        await startGameAs(request, adminToken, game.id);

        const admin = await openAdminClient(adminToken, game.id, first.playerID, second.playerID);
        holder.admin = admin;
        await keepAllHands(admin);
        await admin.waitFor((v) => v.state === "active", "game state active");

        for (const name of LANDS) {
          await adminMoveByName(admin, first.playerID, name, "library", "battlefield");
        }
        await admin.waitFor(
          (v) => LANDS.every((n) => findCardOnBattlefield(v, n)),
          "all seven lands are out",
        );
        const lighthouse = findCardOnBattlefield(admin.snapshot(), LIGHTHOUSE)!;

        const page = first.page;
        await page.setViewportSize(size);
        const tile = page.locator(`.card[data-instance-id="${lighthouse.instance_id}"]`).first();
        await expect(tile).toBeVisible();

        // The lands row: no card is buried under its neighbour. A point
        // inside the tile, near its left edge and at its centre, paints
        // this very card.
        const box = (await tile.boundingBox())!;
        for (const dx of [10, box.width / 2]) {
          expect(await topCardAt(page, box.x + dx, box.y + box.height / 2)).toBe(
            lighthouse.instance_id,
          );
        }
        await testInfo.attach(`lands row ${size.width}x${size.height}`, {
          body: await page.screenshot(),
          contentType: "image/png",
        });

        await tile.click();
        const menu = page.locator("[data-popover-host] .mana-menu");
        await expect(menu).toBeVisible();
        const rows = menu.locator(".menu-item");
        expect(await rows.count()).toBeGreaterThanOrEqual(2);

        // Every row is inside the viewport and is what a click there hits.
        const n = await rows.count();
        for (let i = 0; i < n; i++) {
          const b = (await rows.nth(i).boundingBox())!;
          expect(b.x).toBeGreaterThanOrEqual(0);
          expect(b.y).toBeGreaterThanOrEqual(0);
          expect(b.x + b.width).toBeLessThanOrEqual(size.width);
          expect(b.y + b.height).toBeLessThanOrEqual(size.height);
          const hit = await page.evaluate(
            ([x, y]) => !!document.elementFromPoint(x, y)?.closest(".mana-menu"),
            [b.x + b.width / 2, b.y + b.height / 2],
          );
          expect(hit, `row ${i} is covered`).toBe(true);
          await rows.nth(i).click({ trial: true });
        }

        // Opening it did not make the lands row scroll.
        const scrolls = await page.evaluate(() =>
          [...document.querySelectorAll<HTMLElement>(".row.strip")].map(
            (r) => r.scrollHeight - r.clientHeight,
          ),
        );
        for (const extra of scrolls) expect(extra).toBeLessThanOrEqual(1);

        await testInfo.attach(`menu open ${size.width}x${size.height}`, {
          body: await page.screenshot(),
          contentType: "image/png",
        });

        // Escape still closes it.
        await page.keyboard.press("Escape");
        await expect(menu).toHaveCount(0);
      } finally {
        holder.admin?.close();
        await closeAll(holder.first ?? null, holder.second ?? null);
      }
    });
  }
});
