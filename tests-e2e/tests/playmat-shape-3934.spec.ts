import { expect, test, type Page } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { openAdminClient } from "./s19-helpers";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";

// #3934: a seat's playmat covers THAT seat's battlefield box, not the
// screen. The server mints a real mat only for a signed-in account, so
// the spec stands one in: the game socket's frames get a playmat_url on
// every seat, and the route is answered with a labelled SVG.

const MAT = "/playmats/11111111-1111-4111-8111-111111111111";
const MAT_W = 2400;
const MAT_H = 1400;
const SVG = `<svg xmlns="http://www.w3.org/2000/svg" width="${MAT_W}" height="${MAT_H}" viewBox="0 0 ${MAT_W} ${MAT_H}">
<rect width="${MAT_W}" height="${MAT_H}" fill="#2d6a4f"/>
<rect x="20" y="20" width="${MAT_W - 40}" height="${MAT_H - 40}" fill="none" stroke="#fefae0" stroke-width="40"/>
<circle cx="1200" cy="700" r="600" fill="none" stroke="#e9c46a" stroke-width="40"/>
<path d="M0 0L2400 1400M2400 0L0 1400" stroke="#e76f51" stroke-width="20"/>
</svg>`;

type Box = { x: number; y: number; width: number; height: number };

async function standInMats(page: Page): Promise<void> {
  await page.route(`**${MAT}*`, (r) =>
    r.fulfill({ status: 200, contentType: "image/svg+xml", body: SVG }),
  );
  await page.routeWebSocket(/\/ws/, (ws) => {
    const server = ws.connectToServer();
    ws.onMessage((m) => server.send(m));
    server.onMessage((m) => {
      if (typeof m !== "string") return ws.send(m);
      try {
        const v = JSON.parse(m);
        const walk = (o: unknown): void => {
          if (Array.isArray(o)) o.forEach(walk);
          else if (o && typeof o === "object") {
            const r = o as Record<string, unknown>;
            if (typeof r.seat === "number" && typeof r.life === "number") r.playmat_url = MAT;
            Object.values(r).forEach(walk);
          }
        };
        walk(v);
        ws.send(JSON.stringify(v));
      } catch {
        ws.send(m);
      }
    });
  });
}

// What the mat shows of its image: the cover scale, and the fraction
// of the image's width and height that lands inside the box.
function coverShown(box: Box) {
  const s = Math.max(box.width / MAT_W, box.height / MAT_H);
  return {
    w: Math.min(1, box.width / (MAT_W * s)),
    h: Math.min(1, box.height / (MAT_H * s)),
  };
}

for (const seats of [2, 3, 4]) {
  test(`#3934 ${seats} seats: each mat covers its own battlefield box`, async ({
    browser,
    request,
  }, testInfo) => {
    test.setTimeout(90_000);
    const players: JoinedPlayer[] = [];
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(request, adminToken, `Playmat 3934 ${seats} ${Date.now()}`);
      if (!game.invite_token) throw new Error("invite token missing");
      for (let i = 0; i < seats; i++) {
        players.push(await joinAsPlayer(browser, game.id, game.invite_token, `Seat ${i + 1}`));
      }
      for (const p of players) {
        await uploadDeckAs(request, adminToken, game.id, p.playerID, makeCommanderDeck());
      }
      await startGameAs(request, adminToken, game.id);
      const page = players[0].page;
      // One admin connection per seat: an admin acts for the seat it is bound to.
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
      } finally {
        admins.forEach((a) => a.close());
      }
      await standInMats(page);
      await page.reload();
      await expect(page.locator(".seat-panel .playmat img").first()).toBeVisible();

      for (const size of [
        { width: 1920, height: 1080 },
        { width: 1366, height: 768 },
        { width: 390, height: 844 },
      ]) {
        await page.setViewportSize(size);
        await page.waitForTimeout(500);
        const rows = await page.evaluate(() =>
          [...document.querySelectorAll<HTMLElement>(".panel.seat-panel")].map((p) => {
            const m = p.querySelector<HTMLElement>(".playmat img");
            const r = (e: Element) => {
              const b = e.getBoundingClientRect();
              return { x: b.x, y: b.y, width: b.width, height: b.height };
            };
            return {
              self: p.classList.contains("seat-self") || !p.classList.contains("opponent"),
              panel: r(p),
              mat: m ? r(m) : null,
              fit: m ? getComputedStyle(m).objectFit : null,
            };
          }),
        );
        const lines = rows.map((r) => {
          const c = coverShown(r.panel);
          return `${r.self ? "self" : "opp "} ${Math.round(r.panel.width)}x${Math.round(r.panel.height)} ratio ${(r.panel.width / r.panel.height).toFixed(2)} shows ${(c.w * 100).toFixed(0)}% w x ${(c.h * 100).toFixed(0)}% h; mat ${r.mat ? `${Math.round(r.mat.width)}x${Math.round(r.mat.height)}` : "none"}`;
        });
        console.log(`${seats} seats @ ${size.width}x${size.height}\n  ${lines.join("\n  ")}`);
        await testInfo.attach(`${seats}-seats-${size.width}x${size.height}`, {
          body: await page.screenshot(),
          contentType: "image/png",
        });
        for (const r of rows) {
          if (!r.mat) continue;
          // The mat IS the panel's box: not the screen's.
          expect(Math.abs(r.mat.width - r.panel.width)).toBeLessThan(4);
          expect(Math.abs(r.mat.height - r.panel.height)).toBeLessThan(4);
        }
      }
    } finally {
      await closeAll(...players);
    }
  });
}
