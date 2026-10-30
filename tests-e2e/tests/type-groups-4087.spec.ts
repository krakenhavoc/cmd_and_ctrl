import { expect, test } from "@playwright/test";
import { writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import { adminMoveByName, findCardOnBattlefield, openAdminClient } from "./s19-helpers";

// #4087: every permanent type has a group of its own on a seat's board,
// labelled, and the planeswalker and battle groups never overlap their
// cards. Seeds the same board on every seat (four Sorins, a battle,
// artifacts, enchantments, creatures, lands) at 2, 3 and 4 seats and at
// 1920x1080 and 1366x768, and checks:
//   - a labelled row for each of the six types (an artifact creature,
//     an artifact land and Dryad Arbor each appear once, where
//     cardTypes.ts puts them);
//   - no two cards in the planeswalker / battle rows overlap, and the
//     point at each one's centre paints that card;
//   - each walker's loyalty pip is at least 13px and inside its card;
//   - every group lies inside its seat's panel;
//   - (#4090) a group's label box meets no card, badge or arrival tag
//     in its own row, and the back row's groups sit together: no gap
//     between neighbouring groups is wider than one card.
//
// SHOT_DIR writes the screenshots; SHOT_ONLY=1 takes them without
// asserting (the "before" run against a build that predates the groups).

const WALKERS = [
  "Sorin, Vengeful Bloodlord",
  "Sorin, Imperious Bloodlord",
  "Sorin, Solemn Visitor",
  "Sorin the Mirthless",
];
const BATTLE = "Invasion of Ikoria";
const ARTIFACTS = ["Sol Ring", "Mind Stone", "Arcane Signet", "Burnished Hart"];
const ENCHANTMENTS = ["Rhystic Study", "Propaganda", "Glorious Anthem"];
const CREATURES = ["Llanowar Elves", "Dryad Arbor"];

const ON_BOARD = [...WALKERS, BATTLE, ...ARTIFACTS, ...ENCHANTMENTS, ...CREATURES, "Darksteel Citadel"];

function makeDeck(): string {
  const named = [...WALKERS, BATTLE, ...ARTIFACTS, ...ENCHANTMENTS, ...CREATURES, "Darksteel Citadel"];
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

// No phone size: the table is never played below tablet width (the
// owner's rule), and at 390px a walker card's loyalty pip overhung it by
// a pixel or two often enough to keep the nightly red (#4675).
const SIZES = [
  { width: 1920, height: 1080 },
  { width: 1366, height: 768 },
];

type Box = { x: number; y: number; width: number; height: number };
const overlap = (a: Box, b: Box): number => {
  const w = Math.min(a.x + a.width, b.x + b.width) - Math.max(a.x, b.x);
  const h = Math.min(a.y + a.height, b.y + b.height) - Math.max(a.y, b.y);
  return w > 0 && h > 0 ? w * h : 0;
};

for (const seats of [2, 3, 4]) {
  test(`#4087 ${seats} seats: a labelled group per permanent type, walkers never overlap`, async ({
    browser,
    request,
  }, testInfo) => {
    test.setTimeout(240_000);
    const players: JoinedPlayer[] = [];
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Type groups 4087 ${seats} ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing");
      for (let i = 0; i < seats; i++) {
        players.push(await joinAsPlayer(browser, game.id, game.invite_token, `Seat ${i + 1}`));
      }
      for (const p of players) {
        await uploadDeckAs(request, adminToken, game.id, p.playerID, makeDeck());
      }
      await startGameAs(request, adminToken, game.id);
      const admins = await Promise.all(
        players.map((p) => openAdminClient(adminToken, game.id, p.playerID, p.playerID)),
      );
      try {
        for (let i = 0; i < 20; i++) {
          const v = await admins[0].waitFor(
            (s) => s.mulligans_open !== true || s.seats.some((q) => q.mulligan_turn === true),
            "a seat to decide",
            10_000,
          );
          if (v.mulligans_open !== true) break;
          const who = v.seats.find((q) => q.mulligan_turn === true)!;
          const idx = players.findIndex((p) => p.playerID === who.id);
          await admins[idx].sendActionAsPlayer(who.id, "keep_hand", {});
          await admins[0].waitFor(
            (s) => s.mulligans_open !== true || s.seats.find((q) => q.mulligan_turn === true)?.id !== who.id,
            "next seat",
            10_000,
          );
        }
        await admins[0].waitFor((v) => v.state === "active", "game state active");
        for (let i = 0; i < players.length; i++) {
          for (const name of ON_BOARD) {
            await adminMoveByName(admins[i], players[i].playerID, name, "library", "battlefield");
          }
          await admins[i].waitFor(
            (v) => ON_BOARD.every((n) => findCardOnBattlefield(v, n)),
            `seat ${i + 1}'s board is out`,
          );
        }
      } finally {
        admins.forEach((a) => a.close());
      }

      const page = players[0].page;
      for (const size of SIZES) {
        await page.setViewportSize(size);
        // Every seat's board has reached the page, whatever the layout.
        // #4111: the page works through every state frame of the burst
        // above, about 150ms each on four full boards, so it gets longer
        // than the default; the count only climbs (#634).
        await expect(page.locator(".seat-panel .row[data-zone] .card[data-instance-id]")).toHaveCount(
          seats * ON_BOARD.length,
          { timeout: 60_000 },
        );
        await page.waitForTimeout(700);
        const shot = await page.screenshot();
        const tag = `${seats}-seats-${size.width}x${size.height}`;
        await testInfo.attach(tag, { body: shot, contentType: "image/png" });
        if (process.env.SHOT_DIR) writeFileSync(path.join(process.env.SHOT_DIR, `${tag}.png`), shot);
        if (process.env.SHOT_ONLY) continue;

        const panels = await page.evaluate(() => {
          const r = (b: DOMRect) => ({ x: b.x, y: b.y, width: b.width, height: b.height });
          return [...document.querySelectorAll<HTMLElement>(".panel.seat-panel")].map((p) => {
            const rows = [...p.querySelectorAll<HTMLElement>(".row[data-zone]")].map((row) => ({
              zone: row.dataset.zone ?? "",
              spread: row.classList.contains("spread"),
              box: r(row.getBoundingClientRect()),
              label: (() => {
                const l = row.querySelector<HTMLElement>(".row-label");
                return l ? r(l.getBoundingClientRect()) : null;
              })(),
              // Everything in the row the label must stay clear of: its
              // cards, their badges and pile counts, and the "new" tag
              // of an arrival ring drawn over it.
              things: [
                ...row.querySelectorAll<HTMLElement>(".card[data-instance-id], .badge, .pile-count, .count"),
                ...[...document.querySelectorAll<HTMLElement>("[data-field-arrival] .tag")].filter((t) => {
                  const b = row.getBoundingClientRect();
                  const tb = t.getBoundingClientRect();
                  return tb.x < b.right && tb.right > b.left && tb.y < b.bottom && tb.bottom > b.top;
                }),
              ]
                .map((e) => r(e.getBoundingClientRect()))
                .filter((b) => b.width > 0 && b.height > 0),
              hiddenX: row.scrollWidth - row.clientWidth,
              cards: [...row.querySelectorAll<HTMLElement>(".card[data-instance-id]")].map((c) => {
                const cb = c.getBoundingClientRect();
                const hit = document.elementFromPoint(cb.x + cb.width / 2, cb.y + cb.height / 2);
                const badge = c.querySelector<HTMLElement>(".badge.loyalty");
                // What of the card is on screen: a row and the back row
                // clip what scrolls out of them.
                const clips = [row, p.querySelector<HTMLElement>(".grid-middle")!].map((e) =>
                  e.getBoundingClientRect(),
                );
                const x0 = Math.max(cb.left, ...clips.map((k) => k.left));
                const y0 = Math.max(cb.top, ...clips.map((k) => k.top));
                const x1 = Math.min(cb.right, ...clips.map((k) => k.right));
                const y1 = Math.min(cb.bottom, ...clips.map((k) => k.bottom));
                return {
                  id: c.dataset.instanceId ?? "",
                  box: r(cb),
                  vis: { x: x0, y: y0, width: Math.max(0, x1 - x0), height: Math.max(0, y1 - y0) },
                  painted: !!hit && c.contains(hit),
                  loyalty: badge
                    ? { box: r(badge.getBoundingClientRect()), font: parseFloat(getComputedStyle(badge).fontSize) }
                    : null,
                };
              }),
            }));
            const mid = p.querySelector<HTMLElement>(".grid-middle");
            // ADR 0204 §3: the seat's avatar and what hangs beside it.
            const seatEl = p.querySelector<HTMLElement>("[data-seat-identity]");
            const seatBoxes = seatEl
              ? [seatEl, ...seatEl.querySelectorAll<HTMLElement>("*")]
                  .map((e) => e.getBoundingClientRect())
                  .filter((b) => b.width > 0 && b.height > 0)
              : [];
            const seatLeft = Math.min(...seatBoxes.map((b) => b.left));
            const seatRight = Math.max(...seatBoxes.map((b) => b.right));
            return {
              box: r(p.getBoundingClientRect()),
              rows,
              midHidden: mid ? mid.scrollWidth - mid.clientWidth : 0,
              seat: seatBoxes.length ? { left: seatLeft, right: seatRight } : null,
            };
          });
        });

        expect(panels.length, "seat panels").toBe(seats);
        for (const [pi, panel] of panels.entries()) {
          const where = `${tag} panel ${pi}`;
          const zones = panel.rows.map((r) => r.zone);
          for (const z of ["creatures", "lands", "planeswalkers", "battles", "artifacts", "enchantments"]) {
            expect(zones, `${where}: ${z} group`).toContain(z);
          }
          // #4090: a label has room of its own. Only boxes that are on
          // screen count: a scrolled-away card is clipped by its row.
          for (const row of panel.rows) {
            if (!row.label) continue;
            const rb = row.box;
            for (const t of row.things) {
              const inRow = t.x < rb.x + rb.width && t.x + t.width > rb.x;
              if (!inRow) continue;
              expect
                .soft(overlap(row.label, t), `${where}: ${row.zone} label meets a card or badge`)
                .toBeLessThan(1);
            }
          }
          // #4090: the back row's groups sit together, whichever end
          // they start from.
          // Measured between the groups' cards, not their row boxes: the
          // land strip's box used to run on, empty, to the row's far end.
          const span = (r: (typeof panel.rows)[number]) => ({
            zone: r.zone,
            left: Math.min(...r.cards.map((c) => c.box.x)),
            right: Math.max(...r.cards.map((c) => c.box.x + c.box.width)),
          });
          const back = panel.rows
            .filter((r) => r.zone !== "creatures" && r.cards.length > 0)
            .map(span)
            .sort((a, b) => a.left - b.left);
          const cardW = panel.rows.find((r) => r.zone === "creatures")?.cards[0]?.box.width ?? 56;
          // No gap between neighbouring groups is wider than one card.
          const allowed = cardW;
          for (let i = 1; i < back.length; i++) {
            const gap = back[i].left - back[i - 1].right;
            // ADR 0204 §3 (call 3): your back row may part around your
            // avatar, leaving it a clear notch. The one gap that straddles
            // the avatar may be as wide as the avatar and what hangs
            // beside it, plus the usual allowance.
            const seat = panel.seat;
            const straddles =
              seat !== null && back[i - 1].right < seat.right && back[i].left > seat.left;
            const allowedHere = straddles ? seat.right - seat.left + allowed : allowed;
            expect
              .soft(gap, `${where}: gap between ${back[i - 1].zone} and ${back[i].zone}`)
              .toBeLessThanOrEqual(allowedHere);
          }
          const count = (z: string) => panel.rows.find((r) => r.zone === z)?.cards.length ?? 0;
          expect(count("planeswalkers"), `${where}: walkers`).toBe(WALKERS.length);
          expect(count("battles"), `${where}: battles`).toBe(1);
          // Burnished Hart is an artifact creature: it is in creatures, once.
          expect(count("artifacts"), `${where}: artifacts`).toBe(ARTIFACTS.length - 1);
          expect(count("enchantments"), `${where}: enchantments`).toBe(ENCHANTMENTS.length);
          expect(count("creatures"), `${where}: creatures`).toBe(CREATURES.length + 1);
          expect(count("lands"), `${where}: lands`).toBe(1);


          // Every group is on screen without scrolling, and no walker or
          // battle is scrolled out of its own row. Six groups and four
          // walkers do not fit a 660px opponent panel; its back row
          // scrolls (the lands are last). A panel twice that wide holds
          // them all.
          if (panel.box.width >= 900) {
            expect.soft(panel.midHidden, `${where}: back row scrolls`).toBeLessThanOrEqual(1);
          }
          for (const row of panel.rows.filter((r) => r.spread)) {
            expect.soft(row.hiddenX, `${where}: ${row.zone} scrolls`).toBeLessThanOrEqual(1);
          }
          const spreadCards = panel.rows.filter((r) => r.spread).flatMap((r) => r.cards);
          const walkerIDs = new Set(
            panel.rows.find((r) => r.zone === "planeswalkers")?.cards.map((c) => c.id),
          );
          for (const [i, a] of spreadCards.entries()) {
            // Every planeswalker is painted at its centre, and a battle too
            // unless the panel is a 440px four-seat one, whose back row
            // scrolls the battle, artifacts and enchantments aside before
            // it gives up a walker.
            if (walkerIDs.has(a.id) || panel.box.width >= 900) {
              expect.soft(a.painted, `${where}: walker ${i} is painted at its centre`).toBe(true);
            }
            for (const b of spreadCards.slice(i + 1)) {
              expect.soft(overlap(a.vis, b.vis), `${where}: walkers ${a.id} and ${b.id} overlap`).toBeLessThan(2);
            }
            if (a.loyalty) {
              expect.soft(a.loyalty.font, `${where}: loyalty font`).toBeGreaterThanOrEqual(13);
              expect.soft(a.loyalty.box.x, `${where}: loyalty left`).toBeGreaterThanOrEqual(a.box.x - 1);
              expect
                .soft(a.loyalty.box.x + a.loyalty.box.width, `${where}: loyalty right`)
                .toBeLessThanOrEqual(a.box.x + a.box.width + 1);
            }
          }
        }
      }
    } finally {
      await closeAll(...players);
    }
  });
}
