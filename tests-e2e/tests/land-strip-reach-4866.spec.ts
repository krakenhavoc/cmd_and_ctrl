import { expect, test } from "@playwright/test";
import { writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  findCardOnBattlefield,
  openAdminClient,
} from "./s19-helpers";

// #4866: a full land strip on an across-table (flipped) opponent ran
// past its row's left edge, where its scroll range does not reach, so
// the lands there could not be seen at all. Every seat gets the same
// crowded board (a group per permanent type, as in #4087, and fifteen
// lands, twelve of them with names of their own so each is a pile), at
// 2, 3 and 4 seats and at 1280x720, 1366x768 and 1920x1080, and every
// land on every seat's strip, yours and an upright opponent's included,
// must be on screen once the rows a player can scroll (the strip and
// the back row) are scrolled to it.
//
// SHOT_DIR writes the screenshots; SHOT_ONLY=1 takes them without
// asserting.

const OTHERS = [
  "Sorin, Vengeful Bloodlord",
  "Sorin, Imperious Bloodlord",
  "Invasion of Ikoria",
  "Sol Ring",
  "Mind Stone",
  "Rhystic Study",
  "Propaganda",
  "Llanowar Elves",
];
const LANDS = [
  "Forest",
  "Island",
  "Swamp",
  "Mountain",
  "Command Tower",
  "Evolving Wilds",
  "Reliquary Tower",
  "Exotic Orchard",
  "Path of Ancestry",
  "Myriad Landscape",
  "Terramorphic Expanse",
  "Temple of the False God",
];
// Three Plains on top of the twelve: one pile of three.
const PLAINS = 3;
const LAND_COUNT = LANDS.length + PLAINS;

function makeDeck(): string {
  const named = [...OTHERS, ...LANDS];
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

for (const seats of [2, 3, 4]) {
  test(`#4866 ${seats} seats: every land on every strip can be seen`, async ({
    browser,
    request,
  }, testInfo) => {
    test.setTimeout(300_000);
    const players: JoinedPlayer[] = [];
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(
        request,
        adminToken,
        `Land strip 4866 ${seats} ${Date.now()}`,
      );
      if (!game.invite_token) throw new Error("invite token missing");
      for (let i = 0; i < seats; i++) {
        players.push(
          await joinAsPlayer(
            browser,
            game.id,
            game.invite_token,
            `Seat ${i + 1}`,
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
      const admins = await Promise.all(
        players.map((p) =>
          openAdminClient(adminToken, game.id, p.playerID, p.playerID),
        ),
      );
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
        for (let i = 0; i < players.length; i++) {
          for (const name of [...OTHERS, ...LANDS]) {
            await adminMoveByName(
              admins[i],
              players[i].playerID,
              name,
              "library",
              "battlefield",
            );
          }
          for (let k = 0; k < PLAINS; k++) {
            await adminMoveByName(
              admins[i],
              players[i].playerID,
              "Plains",
              "library",
              "battlefield",
            );
          }
          await admins[i].waitFor(
            (v) =>
              [...OTHERS, ...LANDS, "Plains"].every((n) =>
                findCardOnBattlefield(v, n),
              ),
            `seat ${i + 1}'s board is out`,
          );
        }
      } finally {
        admins.forEach((a) => a.close());
      }

      const page = players[0].page;
      for (const size of SIZES) {
        await page.setViewportSize(size);
        // Every seat's lands have reached the page. The count only
        // climbs (#634); the page works through a burst of frames.
        await expect(
          page.locator(
            '.seat-panel .row[data-zone="lands"] .card[data-instance-id]',
          ),
        ).toHaveCount(seats * LAND_COUNT, { timeout: 60_000 });
        await page.waitForTimeout(700);
        const shot = await page.screenshot();
        const tag = `${seats}-seats-${size.width}x${size.height}`;
        await testInfo.attach(tag, { body: shot, contentType: "image/png" });
        if (process.env.SHOT_DIR)
          writeFileSync(path.join(process.env.SHOT_DIR, `${tag}.png`), shot);
        if (process.env.SHOT_ONLY) continue;

        const panels = await page.evaluate(() => {
          // The boxes between a card and its seat's panel that clip it,
          // and of those, the ones a player can scroll.
          const clippers = (
            el: HTMLElement,
            stop: HTMLElement,
          ): HTMLElement[] => {
            const out: HTMLElement[] = [];
            for (let e = el.parentElement; e; e = e.parentElement) {
              const cs = getComputedStyle(e);
              if (cs.overflowX !== "visible" || cs.overflowY !== "visible")
                out.push(e);
              if (e === stop) break;
            }
            return out;
          };
          const scrolls = (e: HTMLElement): boolean => {
            const cs = getComputedStyle(e);
            return (
              /auto|scroll/.test(cs.overflowX) ||
              /auto|scroll/.test(cs.overflowY)
            );
          };
          // Scroll each scrollable box, innermost first, by as much as
          // brings the card into it. The browser clamps a scroll to its
          // range, so a card outside the range stays out of view.
          const bringIntoView = (
            card: HTMLElement,
            boxes: HTMLElement[],
          ): void => {
            for (const s of boxes.filter(scrolls)) {
              const c = card.getBoundingClientRect();
              const k = s.getBoundingClientRect();
              const dx =
                c.left < k.left
                  ? c.left - k.left
                  : c.right > k.right
                    ? c.right - k.right
                    : 0;
              const dy =
                c.top < k.top
                  ? c.top - k.top
                  : c.bottom > k.bottom
                    ? c.bottom - k.bottom
                    : 0;
              s.scrollLeft += dx;
              s.scrollTop += dy;
            }
          };
          const shown = (card: HTMLElement, boxes: HTMLElement[]): number => {
            const c = card.getBoundingClientRect();
            let x0 = Math.max(c.left, 0);
            let y0 = Math.max(c.top, 0);
            let x1 = Math.min(c.right, innerWidth);
            let y1 = Math.min(c.bottom, innerHeight);
            for (const b of boxes) {
              const k = b.getBoundingClientRect();
              x0 = Math.max(x0, k.left);
              y0 = Math.max(y0, k.top);
              x1 = Math.min(x1, k.right);
              y1 = Math.min(y1, k.bottom);
            }
            const area = c.width * c.height;
            return area > 0
              ? (Math.max(0, x1 - x0) * Math.max(0, y1 - y0)) / area
              : 0;
          };
          return [
            ...document.querySelectorAll<HTMLElement>(".panel.seat-panel"),
          ].map((p) => {
            const strip = p.querySelector<HTMLElement>(
              '.row[data-zone="lands"]',
            );
            const cards = strip
              ? [
                  ...strip.querySelectorAll<HTMLElement>(
                    ".card[data-instance-id]",
                  ),
                ]
              : [];
            const result = {
              flipped: p.classList.contains("flipped"),
              self: p.classList.contains("self"),
              width: Math.round(p.getBoundingClientRect().width),
              stripHidden: strip ? strip.scrollWidth - strip.clientWidth : 0,
              lands: cards.map((c) => {
                const boxes = clippers(c, p);
                bringIntoView(c, boxes);
                return {
                  id: c.dataset.instanceId ?? "",
                  shown: shown(c, boxes),
                };
              }),
            };
            // Leave the rows where a player first sees them.
            for (const s of p.querySelectorAll<HTMLElement>(
              ".row, .grid-middle",
            )) {
              s.scrollLeft = 0;
              s.scrollTop = 0;
            }
            return result;
          });
        });

        expect(panels.length, `${tag}: seat panels`).toBe(seats);
        expect(
          panels.some((p) => p.flipped),
          `${tag}: an across-table seat`,
        ).toBe(true);
        for (const [pi, panel] of panels.entries()) {
          const where = `${tag} panel ${pi} (${panel.self ? "you" : panel.flipped ? "across" : "upright"}, ${panel.width}px, strip scrolls ${panel.stripHidden}px)`;
          expect(panel.lands.length, `${where}: lands`).toBe(LAND_COUNT);
          for (const land of panel.lands) {
            // The whole card, give or take a pixel's rounding at its edges.
            expect
              .soft(
                land.shown,
                `${where}: land ${land.id} can be scrolled into view`,
              )
              .toBeGreaterThanOrEqual(0.95);
          }
        }
      }
    } finally {
      await closeAll(...players);
    }
  });
}
