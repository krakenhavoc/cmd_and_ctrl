import { expect, test, type Page, type Route } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import { keepAllHands, openAdminClient } from "./s19-helpers";
import { L } from "../../client/src/lib/labels";

// #4116: the report form can include a picture of the board, from a
// checkbox that starts unticked. It is taken in the reporter's own browser
// from the table they are looking at, shows among the screenshots before
// anything is sent, goes up as an ordinary attachment (the same caps and
// storage as one they pasted), and leaves out the form itself. A pasted or
// picked image still works beside it.
//
// The server here has no GitHub token, so reporting is switched on for
// the page by answering GET /bugreport/config, and the POST is caught and
// read rather than filed.
//
// SHOT_DIR writes the picture the form sent (and a screenshot of the form
// with it attached) for the PR.

const MAX_IMAGE_BYTES = 4 * 1024 * 1024;

// One byte-for-byte 1x1 PNG, for the "pick or paste one yourself" half.
const TINY_PNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
  "base64",
);

interface Posted {
  fields: Record<string, string>;
  files: { name: string; type: string; bytes: Buffer }[];
}

// A minimal multipart/form-data reader: enough to pull the images out
// of what the form posts.
function parseMultipart(contentType: string, body: Buffer): Posted {
  // With no images the form posts plain JSON.
  if (contentType.startsWith("application/json")) {
    return { fields: JSON.parse(body.toString("utf8")), files: [] };
  }
  const boundary = /boundary=(.+)$/.exec(contentType)?.[1];
  if (!boundary) throw new Error(`no boundary in ${contentType}`);
  const sep = Buffer.from(`--${boundary}`);
  const out: Posted = { fields: {}, files: [] };
  let at = body.indexOf(sep);
  while (at >= 0) {
    const next = body.indexOf(sep, at + sep.length);
    if (next < 0) break;
    const part = body.subarray(at + sep.length + 2, next - 2);
    const split = part.indexOf("\r\n\r\n");
    const head = part.subarray(0, split).toString("utf8");
    const data = part.subarray(split + 4);
    const name = /name="([^"]*)"/.exec(head)?.[1] ?? "";
    const file = /filename="([^"]*)"/.exec(head)?.[1];
    if (file !== undefined) {
      const type = /Content-Type:\s*([^\r\n]+)/i.exec(head)?.[1] ?? "";
      out.files.push({ name: file, type, bytes: Buffer.from(data) });
    } else {
      out.fields[name] = data.toString("utf8");
    }
    at = next;
  }
  return out;
}

const pngSize = (b: Buffer): { w: number; h: number } => ({
  w: b.readUInt32BE(16),
  h: b.readUInt32BE(20),
});

async function enableReporting(page: Page, posted: Posted[]): Promise<void> {
  await page.route("**/bugreport/config", (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ enabled: true, attachments: true }),
    }),
  );
  await page.route("**/bugreport", async (route: Route) => {
    const req = route.request();
    if (req.method() !== "POST") return route.continue();
    posted.push(
      parseMultipart(
        req.headers()["content-type"] ?? "",
        req.postDataBuffer() ?? Buffer.alloc(0),
      ),
    );
    await route.fulfill({
      status: 201,
      contentType: "application/json",
      body: JSON.stringify({
        number: 7,
        url: "https://github.com/x/y/issues/7",
      }),
    });
  });
}

test("#4116 the report form takes a picture of the board, only when asked", async ({
  browser,
  request,
}) => {
  test.setTimeout(180_000);
  const players: JoinedPlayer[] = [];
  try {
    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `Board shot 4116 ${Date.now()}`,
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
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        p.playerID,
        makeCommanderDeck(),
      );
    await startGameAs(request, adminToken, game.id);
    const admin = await openAdminClient(
      adminToken,
      game.id,
      players[0].playerID,
      players[1].playerID,
    );
    try {
      await keepAllHands(admin);
    } finally {
      admin.close();
    }

    const page = players[0].page;
    const posted: Posted[] = [];
    await enableReporting(page, posted);
    await page.reload();
    await page.setViewportSize({ width: 1280, height: 720 });

    // Wait for the table to be on screen.
    await expect(page.locator("[data-pile='library']").first()).toBeVisible({
      timeout: 20_000,
    });
    await page.waitForTimeout(1500);

    await page.getByRole("button", { name: L.reportABug }).first().click();
    const form = page.locator(
      '[role="dialog"][aria-labelledby="bug-report-title"]',
    );
    await expect(form).toBeVisible();
    await form.locator('input[type="text"]').fill("Board picture check");

    // Offered, unticked, and nothing taken yet.
    const box = form.getByRole("checkbox", {
      name: /include a picture of the board/,
    });
    await expect(box).toBeVisible();
    await expect(box).not.toBeChecked();
    await expect(form.locator(".shots img")).toHaveCount(0);

    // Ticked: the picture appears among the screenshots.
    await box.check();
    await expect(form.locator(".shots img")).toHaveCount(1, {
      timeout: 30_000,
    });
    await expect(box).toBeChecked();
    const thumb = form.locator(".shots img").first();
    expect(
      await thumb.evaluate((i: HTMLImageElement) => i.naturalWidth),
    ).toBeGreaterThan(300);

    // A pasted or picked image still goes up beside it.
    await form.locator('input[type="file"]').setInputFiles({
      name: "mine.png",
      mimeType: "image/png",
      buffer: TINY_PNG,
    });
    await expect(form.locator(".shots img")).toHaveCount(2);

    const shotDir = process.env.SHOT_DIR;
    if (shotDir) {
      mkdirSync(shotDir, { recursive: true });
      writeFileSync(
        path.join(shotDir, "report-form-board-shot.png"),
        await page.screenshot(),
      );
    }

    await form.getByRole("button", { name: /^(Send|File|Submit)/i }).click();
    await expect.poll(() => posted.length, { timeout: 30_000 }).toBe(1);

    const [sent] = posted;
    expect(sent.files.map((f) => f.type)).toEqual(["image/png", "image/png"]);
    const board = sent.files.find((f) => f.name.startsWith("board-"))!;
    expect(board).toBeTruthy();
    expect(board.bytes.subarray(0, 4).toString("hex")).toBe("89504e47");
    // The table, not a stub: wider than the form's own width and tall
    // enough to hold two seats, and inside the caps.
    const { w, h } = pngSize(board.bytes);
    expect(w).toBeGreaterThan(1000);
    expect(h).toBeGreaterThan(500);
    expect(board.bytes.length).toBeLessThanOrEqual(MAX_IMAGE_BYTES);
    expect(board.bytes.length).toBeGreaterThan(20_000);
    expect(
      sent.files.reduce((n, f) => n + f.bytes.length, 0),
    ).toBeLessThanOrEqual(10 * 1024 * 1024);
    if (shotDir)
      writeFileSync(path.join(shotDir, "board-picture-sent.png"), board.bytes);
  } finally {
    await closeAll(...players);
  }
});

test("#4116 a report sent without ticking the box carries no picture of the board", async ({
  browser,
  request,
}) => {
  test.setTimeout(120_000);
  const players: JoinedPlayer[] = [];
  try {
    const adminToken = await adminLogin(request);
    const game = await createGame(
      request,
      adminToken,
      `No board shot 4116 ${Date.now()}`,
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
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        p.playerID,
        makeCommanderDeck(),
      );
    await startGameAs(request, adminToken, game.id);
    const admin = await openAdminClient(
      adminToken,
      game.id,
      players[0].playerID,
      players[1].playerID,
    );
    try {
      await keepAllHands(admin);
    } finally {
      admin.close();
    }

    const page = players[0].page;
    const posted: Posted[] = [];
    await enableReporting(page, posted);
    await page.reload();
    await expect(page.locator("[data-pile='library']").first()).toBeVisible({
      timeout: 20_000,
    });

    await page.getByRole("button", { name: L.reportABug }).first().click();
    const form = page.locator(
      '[role="dialog"][aria-labelledby="bug-report-title"]',
    );
    await form.locator('input[type="text"]').fill("No picture");
    await form.getByRole("button", { name: /^(Send|File|Submit)/i }).click();
    await expect.poll(() => posted.length, { timeout: 30_000 }).toBe(1);
    expect(posted[0].files).toEqual([]);
  } finally {
    await closeAll(...players);
  }
});
