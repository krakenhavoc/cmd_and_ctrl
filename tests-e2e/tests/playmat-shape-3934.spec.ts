import { expect, test, type Page } from "@playwright/test";
import { writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { openAdminClient } from "./s19-helpers";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";

// #3934: a seat's playmat is shown WHOLE (object-fit: contain) inside THAT
// seat's battlefield box, over a blurred copy of itself that fills the
// strips; the copy is the box (not the screen). The server mints a real
// mat only for a signed-in account, so the spec stands one in: the game socket's frames get a playmat_url on
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
            const mat = p.querySelector<HTMLImageElement>(".playmat img.mat");
            const back = p.querySelector<HTMLImageElement>(".playmat img.backdrop");
            const r = (b: DOMRect) => ({ x: b.x, y: b.y, width: b.width, height: b.height });
            // Where object-fit: contain really paints the image: the
            // largest rectangle of its own shape inside the element box.
            let drawn: { x: number; y: number; width: number; height: number } | null = null;
            if (mat) {
              const b = mat.getBoundingClientRect();
              const k = Math.min(b.width / mat.naturalWidth, b.height / mat.naturalHeight);
              const w = mat.naturalWidth * k;
              const h = mat.naturalHeight * k;
              drawn = { x: b.x + (b.width - w) / 2, y: b.y + (b.height - h) / 2, width: w, height: h };
            }
            return {
              self: !p.classList.contains("opponent"),
              panel: r(p.getBoundingClientRect()),
              // Layout size: the blur's scale-up is a transform and not part of it.
              back: back ? { width: back.offsetWidth, height: back.offsetHeight } : null,
              mat: mat ? r(mat.getBoundingClientRect()) : null,
              fit: mat ? getComputedStyle(mat).objectFit : null,
              backFit: back ? getComputedStyle(back).objectFit : null,
              drawn,
            };
          }),
        );
        const lines = rows.map(
          (r) =>
            `${r.self ? "self" : "opp "} panel ${Math.round(r.panel.width)}x${Math.round(r.panel.height)} (${(r.panel.width / r.panel.height).toFixed(2)}); mat drawn ${r.drawn ? `${Math.round(r.drawn.width)}x${Math.round(r.drawn.height)}` : "none"}`,
        );
        console.log(`${seats} seats @ ${size.width}x${size.height}\n  ${lines.join("\n  ")}`);
        const shot = await page.screenshot();
        await testInfo.attach(`${seats}-seats-${size.width}x${size.height}`, {
          body: shot,
          contentType: "image/png",
        });
        if (process.env.PLAYMAT_SHOT_DIR) {
          writeFileSync(
            path.join(process.env.PLAYMAT_SHOT_DIR, `${seats}-seats-${size.width}x${size.height}.png`),
            shot,
          );
        }
        for (const r of rows) {
          // The backdrop layer is the panel's box: not the screen's.
          if (!r.back || !r.drawn) {
            expect.soft(false, "both mat layers are drawn").toBe(true);
            continue;
          }
          const box = r.back;
          expect.soft(Math.abs(box.width - r.panel.width), "backdrop width").toBeLessThan(4);
          expect.soft(Math.abs(box.height - r.panel.height), "backdrop height").toBeLessThan(4);
          // The mat itself is shown whole: contained, nothing cropped.
          expect.soft(r.fit, "mat object-fit").toBe("contain");
          expect.soft(r.backFit, "backdrop object-fit").toBe("cover");
          const d = r.drawn;
          // It fits inside the panel...
          expect.soft(d.x, "inside left").toBeGreaterThanOrEqual(r.panel.x - 0.5);
          expect.soft(d.y, "inside top").toBeGreaterThanOrEqual(r.panel.y - 0.5);
          expect.soft(d.x + d.width, "inside right").toBeLessThanOrEqual(r.panel.x + r.panel.width + 0.5);
          expect.soft(d.y + d.height, "inside bottom").toBeLessThanOrEqual(r.panel.y + r.panel.height + 0.5);
          // ...at the image's whole aspect ratio...
          expect.soft(d.width / d.height, "aspect").toBeCloseTo(MAT_W / MAT_H, 2);
          // ...touching the panel on at least one axis (as large as it can be)...
          const slackX = r.panel.width - d.width;
          const slackY = r.panel.height - d.height;
          expect.soft(Math.min(slackX, slackY), "as large as fits").toBeLessThan(4);
          // ...and centred.
          const cx = d.x + d.width / 2 - (r.panel.x + r.panel.width / 2);
          const cy = d.y + d.height / 2 - (r.panel.y + r.panel.height / 2);
          expect.soft(Math.abs(cx), "centred x").toBeLessThan(3);
          expect.soft(Math.abs(cy), "centred y").toBeLessThan(3);
        }
      }
    } finally {
      await closeAll(...players);
    }
  });
}
