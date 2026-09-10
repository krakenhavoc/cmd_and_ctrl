import { expect, type Browser, type BrowserContext, type Page } from "@playwright/test";

// Shared browser-side player onboarding. full-game, board-layout,
// mulligan and zone-browser all need "a fresh context that walked the
// public invite flow and is now sitting on the game route"; three
// byte-identical copies of that function is exactly the kind of
// duplication that let the suite rot in the first place (a selector
// change had to be found and fixed in N places).
//
// The S19 helper keeps its own variant: it waits on localStorage
// rather than the lobby URL because it drives many more sessions per
// run and needs the longer, more forgiving handshake.

export interface JoinedPlayer {
  context: BrowserContext;
  page: Page;
  name: string;
  token: string;
  playerID: string;
}

// joinAsPlayer walks the public invite-link flow in a fresh browser
// context, enters `name`, and leaves the page on the game route with
// the issued session captured. The returned token + playerID let a
// test drive that seat either through this page or through the
// server API (admin helpers in lobby-api.ts).
export async function joinAsPlayer(
  browser: Browser,
  gameID: string,
  inviteToken: string,
  name: string,
): Promise<JoinedPlayer> {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto(`/#/games/${gameID}/join?t=${encodeURIComponent(inviteToken)}`);
  await page.getByPlaceholder("your name").fill(name);
  await page.getByRole("button", { name: "join" }).click();
  // Players land in the lobby first — s085 (#43) — then a seated
  // session can open the game route directly.
  await expect(page).toHaveURL(/#\/lobby$/, { timeout: 10_000 });
  await page.goto(`/#/games/${gameID}`);
  await expect(page).toHaveURL(new RegExp(`#/games/${gameID}$`), { timeout: 10_000 });

  const session = await page.evaluate(() =>
    JSON.parse(localStorage.getItem("cmdctrl.session") ?? "null"),
  );
  if (!session?.token || !session?.playerID) {
    throw new Error(`${name}: session missing after join`);
  }
  return { context, page, name, token: session.token, playerID: session.playerID };
}

// closeAll tears down every context a test opened. Safe to call in a
// finally / afterEach even if some entries are undefined.
export async function closeAll(...players: (JoinedPlayer | null | undefined)[]): Promise<void> {
  for (const p of players) {
    if (p) await p.context.close();
  }
}
