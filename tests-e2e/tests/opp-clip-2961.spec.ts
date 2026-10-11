import { expect, test } from "@playwright/test";
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

// #2961: at 1600x900 the opponent's creatures lost their power/toughness
// behind the line between the two boards, and their commander and hand
// fan sat partly under the page header. Checked at 1280x720 and
// 1600x900: every opponent creature's P/T badge lies inside the
// opponent panel's rect, is what a point at its centre paints, and is
// not cut by a clipping ancestor; and the opponent panel starts at or
// below the page header.

const CREATURES = ["World Shaper", "Burnished Hart", "Llanowar Elves"];
// A back row with lands and an artifact leaves the creature row less of
// the panel, which is when its card used to be cut.
const BACK_ROW = ["Sol Ring", "Mind Stone", "Plains", "Plains", "Plains"];

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      ...CREATURES.map((n) => `1 ${n}`),
      "1 Sol Ring",
      "1 Mind Stone",
      "94 Plains",
    ].join("\n") + "\n"
  );
}

test.describe("#2961 opponent board is not clipped", () => {
  for (const size of [
    { width: 1280, height: 720 },
    { width: 1600, height: 900 },
  ]) {
    test(`${size.width}x${size.height}: opponent P/T badges are inside their panel`, async ({
      browser,
      request,
    }, testInfo) => {
      test.slow();
      const holder: {
        first?: JoinedPlayer;
        second?: JoinedPlayer;
        admin?: AdminClient;
      } = {};
      try {
        const adminToken = await adminLogin(request);
        const game = await createGame(
          request,
          adminToken,
          `Opp clip 2961 ${Date.now()}`,
        );
        if (!game.invite_token)
          throw new Error("invite token missing on fresh game");
        const first = await joinAsPlayer(
          browser,
          game.id,
          game.invite_token,
          "Seat One",
        );
        holder.first = first;
        const second = await joinAsPlayer(
          browser,
          game.id,
          game.invite_token,
          "Seat Two",
        );
        holder.second = second;
        await uploadDeckAs(
          request,
          adminToken,
          game.id,
          first.playerID,
          makeDeck(),
        );
        await uploadDeckAs(
          request,
          adminToken,
          game.id,
          second.playerID,
          makeDeck(),
        );
        await startGameAs(request, adminToken, game.id);

        const admin = await openAdminClient(
          adminToken,
          game.id,
          first.playerID,
          second.playerID,
        );
        holder.admin = admin;
        await keepAllHands(admin);
        await admin.waitFor((v) => v.state === "active", "game state active");

        for (const name of [...CREATURES, ...BACK_ROW]) {
          await adminMoveByName(
            admin,
            second.playerID,
            name,
            "library",
            "battlefield",
          );
        }
        await admin.waitFor(
          (v) => CREATURES.every((n) => findCardOnBattlefield(v, n)),
          "the opponent's creatures are out",
        );

        const page = first.page;
        await page.setViewportSize(size);
        const panel = page.locator(".panel.opponent").first();
        await expect(panel).toBeVisible();
        const pts = panel.locator(".grid-creatures .card .badge.pt");
        await expect(pts).toHaveCount(CREATURES.length);
        await testInfo.attach(`board ${size.width}x${size.height}`, {
          body: await page.screenshot(),
          contentType: "image/png",
        });

        const p = (await panel.boundingBox())!;
        for (let i = 0; i < CREATURES.length; i++) {
          const b = (await pts.nth(i).boundingBox())!;
          expect(b.x, `badge ${i} left`).toBeGreaterThanOrEqual(p.x);
          expect(b.y, `badge ${i} top`).toBeGreaterThanOrEqual(p.y);
          expect(b.x + b.width, `badge ${i} right`).toBeLessThanOrEqual(
            p.x + p.width + 0.5,
          );
          expect(b.y + b.height, `badge ${i} bottom`).toBeLessThanOrEqual(
            p.y + p.height + 0.5,
          );
          // It is also what is painted there: not covered by the
          // neighbouring board.
          const painted = await pts.nth(i).evaluate((el) => {
            const r = el.getBoundingClientRect();
            const hit = document.elementFromPoint(
              r.x + r.width / 2,
              r.y + r.height / 2,
            );
            return !!hit && !!el.closest(".card")?.contains(hit);
          });
          expect(painted, `badge ${i} is painted`).toBe(true);
          // And no clipping ancestor between the badge and the panel
          // cuts it.
          const clipped = await pts.nth(i).evaluate((el) => {
            const r = el.getBoundingClientRect();
            for (
              let a = el.closest(".card")?.parentElement ?? null;
              a;
              a = a.parentElement
            ) {
              const cs = getComputedStyle(a);
              if (cs.overflowX !== "visible" || cs.overflowY !== "visible") {
                const ar = a.getBoundingClientRect();
                if (
                  r.top < ar.top - 0.5 ||
                  r.bottom > ar.bottom + 0.5 ||
                  r.left < ar.left - 0.5 ||
                  r.right > ar.right + 0.5
                ) {
                  return String(a.className);
                }
              }
              if (a.classList.contains("panel")) break;
            }
            return "";
          });
          expect(clipped, `badge ${i} is clipped by ${clipped}`).toBe("");
        }

        // The panel (hand fan and commander tucked at its top edge)
        // starts below the page header.
        const headerBottom = await page.evaluate(() => {
          const h = document.querySelector("header");
          return h ? h.getBoundingClientRect().bottom : 0;
        });
        expect(p.y).toBeGreaterThanOrEqual(headerBottom - 0.5);
      } finally {
        holder.admin?.close();
        await closeAll(holder.first ?? null, holder.second ?? null);
      }
    });
  }
});
