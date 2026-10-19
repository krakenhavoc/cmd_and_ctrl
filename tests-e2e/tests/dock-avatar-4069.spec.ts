import { expect, test } from "@playwright/test";
import { writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import { adminMoveByName, findCardOnBattlefield, openAdminClient } from "./s19-helpers";

// #4069: at a four-seat table, the declare-attackers dock is at its
// tallest (one attack button per opponent). It must leave the viewer's own
// avatar and life total visible at 1024x768 (tablet is the minimum
// screen), 1366x768 and 1920x1080, and opening an opponent's expanded
// board (ADR 0120) must log no "ResizeObserver loop" error.
//
// SHOT_DIR writes the screenshots; SHOT_ONLY=1 takes them without asserting
// (the "before" run).

const BEARS = ["Grizzly Bears", "Runeclaw Bear", "Balduvian Bears", "Alpine Grizzly"];
const SIZES = [
  { width: 1024, height: 768 },
  { width: 1366, height: 768 },
  { width: 1920, height: 1080 },
];

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      ...BEARS.map((n) => `1 ${n}`),
      `${99 - BEARS.length} Forest`,
    ].join("\n") + "\n"
  );
}

type Box = { x: number; y: number; width: number; height: number };
const overlap = (a: Box, b: Box): number => {
  const w = Math.min(a.x + a.width, b.x + b.width) - Math.max(a.x, b.x);
  const h = Math.min(a.y + a.height, b.y + b.height) - Math.max(a.y, b.y);
  return w > 0 && h > 0 ? w * h : 0;
};

test("#4069 the tall dock leaves your avatar and life visible; the overlay logs no observer loop", async ({
  browser,
  request,
}, testInfo) => {
  test.setTimeout(240_000);
  const players: JoinedPlayer[] = [];
  try {
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Dock avatar 4069 ${Date.now()}`);
    if (!game.invite_token) throw new Error("invite token missing");
    for (let i = 0; i < 4; i++) {
      players.push(await joinAsPlayer(browser, game.id, game.invite_token, `Archmage Thaddeus the ${["First","Second","Third","Fourth"][i]}`));
    }
    for (const p of players) await uploadDeckAs(request, adminToken, game.id, p.playerID, makeDeck());
    await startGameAs(request, adminToken, game.id);
    const admins = await Promise.all(
      players.map((p) => openAdminClient(adminToken, game.id, p.playerID, p.playerID)),
    );
    const admin = admins[0];
    let attacker!: JoinedPlayer;
    try {
      for (let i = 0; i < 20; i++) {
        const v = await admin.waitFor(
          (s) => s.mulligans_open !== true || s.seats.some((q) => q.mulligan_turn === true),
          "a seat to decide",
          10_000,
        );
        if (v.mulligans_open !== true) break;
        const who = v.seats.find((q) => q.mulligan_turn === true)!;
        await admins[players.findIndex((p) => p.playerID === who.id)].sendActionAsPlayer(who.id, "keep_hand", {});
        await admin.waitFor(
          (s) => s.mulligans_open !== true || s.seats.find((q) => q.mulligan_turn === true)?.id !== who.id,
          "next seat",
          10_000,
        );
      }
      await admin.waitFor(
        (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
        "the attacker's precombat main",
        20_000,
      );
      const snap = admin.snapshot();
      const activeID = snap.seats[snap.turn?.active_seat ?? 0]?.id;
      attacker = players.find((p) => p.playerID === activeID)!;
      const aAdmin = admins[players.indexOf(attacker)];
      for (const bear of BEARS) await adminMoveByName(aAdmin, attacker.playerID, bear, "library", "battlefield");
      await admin.waitFor((v) => BEARS.every((b) => findCardOnBattlefield(v, b) !== null), "bears out");
      await aAdmin.sendActionAsPlayer(attacker.playerID, "untap_all", {});
      await aAdmin.sendAction("advance_step", {});
      await admin.waitFor((v) => v.turn?.step === "declare_attackers", "declare_attackers", 20_000);
    } finally {
      admins.forEach((a) => a.close());
    }

    const page = attacker.page;
    const logs: string[] = [];
    page.on("console", (m) => logs.push(m.text()));
    page.on("pageerror", (e) => logs.push(String(e.message)));
    // The browser reports the loop as a window "error" event, which
    // Playwright surfaces as neither a console line nor a pageerror.
    await page.evaluate(() => {
      const w = window as unknown as { __errs: string[] };
      w.__errs = [];
      window.addEventListener("error", (e) => w.__errs.push(e.message));
    });

    const cluster = page
      .getByRole("region", { name: "actions", exact: true })
      .getByRole("group", { name: "declare attackers" });
    await expect(cluster).toBeVisible({ timeout: 15_000 });

    for (const size of SIZES) {
      await page.setViewportSize(size);
      await page.waitForTimeout(800);
      const tag = `dock-${size.width}x${size.height}`;
      const shot = await page.screenshot();
      await testInfo.attach(tag, { body: shot, contentType: "image/png" });
      if (process.env.SHOT_DIR) writeFileSync(path.join(process.env.SHOT_DIR, `${tag}.png`), shot);
      if (process.env.SHOT_ONLY) continue;
      const boxes = await page.evaluate((id) => {
        const r = (b: DOMRect) => ({ x: b.x, y: b.y, width: b.width, height: b.height });
        const dock = document.querySelector<HTMLElement>(".action-dock")!;
        const avatar = document.querySelector<HTMLElement>(`.avatar-wrap[data-seat-id="${id}"]`)!;
        const life = avatar.closest("[data-seat-identity]")!.querySelector<HTMLElement>(".life-chip")!;
        return {
          dock: r(dock.getBoundingClientRect()),
          avatar: r(avatar.getBoundingClientRect()),
          life: r(life.getBoundingClientRect()),
        };
      }, attacker.playerID);
      expect(overlap(boxes.dock, boxes.avatar), `${tag}: dock over avatar`).toBe(0);
      expect(overlap(boxes.dock, boxes.life), `${tag}: dock over life`).toBe(0);
      for (const b of [boxes.avatar, boxes.life]) {
        expect(b.x).toBeGreaterThanOrEqual(0);
        expect(b.y + b.height).toBeLessThanOrEqual(size.height);
      }
    }

    // ADR 0120 overlay: open and close it a few times, then let frames pass.
    await page.setViewportSize({ width: 1366, height: 768 });
    const opp = players.find((p) => p !== attacker)!;
    const oppAvatar = page.locator(`.avatar-wrap[data-seat-id="${opp.playerID}"]`).first();
    for (let i = 0; i < 3; i++) {
      await page.mouse.move(2, 2);
      await oppAvatar.hover();
      await page
        .getByRole("region", { name: /board, expanded/ })
        .first()
        .waitFor({ state: "visible" });
      await page.waitForTimeout(400);
      await page.mouse.move(2, 2);
      await page.waitForTimeout(400);
    }
    logs.push(...(await page.evaluate(() => (window as unknown as { __errs: string[] }).__errs)));
    expect(logs.filter((l) => /ResizeObserver loop/.test(l))).toEqual([]);
  } finally {
    await closeAll(...players);
  }
});
