import { expect, test, type Page } from "@playwright/test";
import { writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import { adminMoveByName, findCardOnBattlefield, openAdminClient } from "./s19-helpers";

// #4725: a player on Windows saw the middle top-row opponent's board
// flash. Windows draws classic scrollbars, which take room; headless
// Chromium hides them (--hide-scrollbars), which is why no run of this
// suite ever saw it, so this file launches with them drawn.
//
// The loop: a hovered land pile spreads its copies with a translate
// (#4422), which on a strip only as wide as its lands ran past the
// strip's end. The strip scrolled on any overflow, so it grew a sideways
// scrollbar; on a top-row panel the strip's zones sit against the row
// below, so the bar's height lifted the strip and the pile moved out from
// under a pointer resting near its lower edge. The hover ended, the bar
// went, the pile came back under the pointer, and so on, every few
// frames. A top-row pile's copies also reached 3px below the strip and
// gave it a vertical bar, which is the one in the report's screenshot.
//
// Every seat gets the reporter's kind of board (a pile of four Plains,
// Mind Stone and four creatures, a flier among them) at 4 seats, three of
// them across the top. At each size the page is sampled every frame,
// first with the pointer off the board and then resting at the centre and
// the top and bottom edges of every land pile on every panel. Each sample
// is every battlefield row's box (size, scrollbars, position, the fitted
// card size and overlap) and which pile is hovered. Once the hover has
// settled it may change at most twice in forty frames; the loop changed
// it on two frames in every three. A strip may draw a scrollbar only
// while its piles do not fit it.
//
// SHOT_DIR writes a screenshot per size.

test.use({ launchOptions: { ignoreDefaultArgs: ["--hide-scrollbars"] } });

const CREATURES = ["Llanowar Elves", "Grizzly Bears", "Savannah Lions", "Serra Angel"];
const OTHERS = ["Mind Stone", ...CREATURES];
const PLAINS = 4;
const SEATS = 4;

// The reporter's window (Chrome on Windows), a narrower desktop where the
// loop showed with the pointer anywhere on the pile's lower edge, and the
// most common laptop size.
const SIZES = [
  { width: 1920, height: 1089 },
  { width: 1440, height: 1089 },
  { width: 1366, height: 768 },
];

// How long the page settles after a resize or a pointer move (the hover
// spread eases open over 140ms), then how many frames it is watched for.
// Frames, not milliseconds: a loop changes the layout every frame or
// two, however few frames a busy runner draws, and a settled page none.
const SETTLE_MS = 500;
const WATCH_FRAMES = 40;
// A late hover or one re-fit may land in the window; a loop is a change
// every few frames.
const MAX_CHANGES = 2;

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      ...OTHERS.map((n) => `1 ${n}`),
      `${99 - OTHERS.length} Plains`,
    ].join("\n") + "\n"
  );
}

interface Watch {
  frames: number;
  // How many frames differed from the one before.
  changes: number;
  // The first few distinct states, for the failure message.
  states: string[];
}

// watch samples the board on each of `frames` animation frames.
async function watch(page: Page, frames: number): Promise<Watch> {
  return page.evaluate(async (frames) => {
    const state = (): string => {
      const rows = [...document.querySelectorAll<HTMLElement>(".panel.seat-panel .row[data-zone]")];
      const parts = rows.map((r, i) => {
        const list = r.querySelector<HTMLElement>(":scope > .row-cards");
        const box = r.getBoundingClientRect();
        return [
          `${i}:${r.dataset.zone}`,
          `box ${r.offsetWidth}x${r.offsetHeight}@${Math.round(box.left)},${Math.round(box.top)}`,
          // A scrollbar is the difference between the two (the rows have
          // no border).
          `client ${r.clientWidth}x${r.clientHeight}`,
          `k ${list?.style.getPropertyValue("--strip-k") ?? ""}`,
          `w ${list?.style.getPropertyValue("--card-w") ?? ""}`,
          `o ${list?.style.getPropertyValue("--fit-overlap") ?? ""}`,
        ].join(" ");
      });
      const hovered = document.querySelectorAll(".panel.seat-panel .row .pile:hover").length;
      return `${parts.join(" | ")} | hovered ${hovered}`;
    };
    let last = state();
    const out = { frames: 0, changes: 0, states: [last] };
    await new Promise<void>((done) => {
      const tick = (): void => {
        out.frames++;
        const s = state();
        if (s !== last) {
          out.changes++;
          last = s;
          if (out.states.length < 4) out.states.push(s);
        }
        if (out.frames < frames) requestAnimationFrame(tick);
        else done();
      };
      requestAnimationFrame(tick);
    });
    return out;
  }, frames);
}

function describe(w: Watch): string {
  return `${w.changes} changes in ${w.frames} frames; first states:\n${w.states.map((s) => s.replaceAll(" | ", "\n    ")).join("\n  ---\n")}`;
}

test(`#4725 ${SEATS} seats with classic scrollbars: no row flickers, at rest or hovered`, async ({
  browser,
  request,
}, testInfo) => {
  test.setTimeout(600_000);
  const players: JoinedPlayer[] = [];
  try {
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Board flicker 4725 ${Date.now()}`);
    if (!game.invite_token) throw new Error("invite token missing");
    for (let i = 0; i < SEATS; i++) {
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
          (s) =>
            s.mulligans_open !== true ||
            s.seats.find((q) => q.mulligan_turn === true)?.id !== who.id,
          "next seat",
          10_000,
        );
      }
      await admins[0].waitFor((v) => v.state === "active", "game state active");
      for (let i = 0; i < players.length; i++) {
        for (const name of OTHERS) {
          await adminMoveByName(admins[i], players[i].playerID, name, "library", "battlefield");
        }
        for (let k = 0; k < PLAINS; k++) {
          await adminMoveByName(admins[i], players[i].playerID, "Plains", "library", "battlefield");
        }
        await admins[i].waitFor(
          (v) => OTHERS.every((n) => findCardOnBattlefield(v, n)),
          `seat ${i + 1}'s board is out`,
        );
      }
    } finally {
      admins.forEach((a) => a.close());
    }

    const page = players[0].page;
    for (const size of SIZES) {
      const tag = `${size.width}x${size.height}`;
      await page.mouse.move(1, 1);
      await page.setViewportSize(size);
      // Every seat's lands have reached the page (a count that only
      // climbs, #634).
      await expect(
        page.locator('.seat-panel .row[data-zone="lands"] .card[data-instance-id]'),
      ).toHaveCount(SEATS * PLAINS, { timeout: 60_000 });
      await expect(page.locator(".panel.seat-panel.flipped")).toHaveCount(SEATS - 1);
      await page.waitForTimeout(SETTLE_MS);
      const shot = await page.screenshot();
      await testInfo.attach(tag, { body: shot, contentType: "image/png" });
      if (process.env.SHOT_DIR) writeFileSync(path.join(process.env.SHOT_DIR, `${tag}.png`), shot);

      const rest = await watch(page, WATCH_FRAMES);
      expect
        .soft(rest.changes, `${tag}, pointer off the board: ${describe(rest)}`)
        .toBeLessThanOrEqual(MAX_CHANGES);

      // No strip draws a scrollbar while its piles fit it.
      const bars = await page.$$eval(".panel.seat-panel .row.strip", (rows) =>
        (rows as HTMLElement[])
          .filter((r) => !r.hasAttribute("data-scrolls"))
          .filter((r) => r.offsetWidth !== r.clientWidth || r.offsetHeight !== r.clientHeight)
          .map((r) => `${r.offsetWidth}x${r.offsetHeight} drawn ${r.clientWidth}x${r.clientHeight}`),
      );
      expect.soft(bars, `${tag}: strips that fit but draw a scrollbar`).toEqual([]);

      const piles = page.locator(".panel.seat-panel .row.strip .pile.multi");
      const n = await piles.count();
      expect(n, `${tag}: a land pile on every seat`).toBe(SEATS);
      for (let i = 0; i < n; i++) {
        const b = await piles.nth(i).boundingBox();
        if (!b) throw new Error(`${tag}: pile ${i} has no box`);
        const flipped = await piles
          .nth(i)
          .evaluate((el) => !!el.closest(".panel.seat-panel.flipped"));
        const x = b.x + b.width / 2;
        for (const [where, y] of [
          ["centre", b.y + b.height / 2],
          ["top edge", b.y + 3],
          ["bottom edge", b.y + b.height - 3],
        ] as const) {
          await page.mouse.move(x, y);
          await page.waitForTimeout(SETTLE_MS);
          const w = await watch(page, WATCH_FRAMES);
          expect
            .soft(
              w.changes,
              `${tag}, pointer at the ${where} of pile ${i} (${flipped ? "across the table" : "upright"}): ${describe(w)}`,
            )
            .toBeLessThanOrEqual(MAX_CHANGES);
        }
      }
      await page.mouse.move(1, 1);
    }
  } finally {
    await closeAll(...players);
  }
});
