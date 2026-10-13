// @vitest-environment jsdom
//
// moderationTables.render.test.ts — #2984 on the pages that meet a
// person: the lobby's Remove confirm (the block box, unticked each time,
// and what it sends) and the Join page's two refusals (blocked from the
// table, and a name on the filter's list) shown as the server's own
// sentence.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Join from "../routes/Join.svelte";
import Lobby from "../routes/Lobby.svelte";
import { BLOCK_HINT, BLOCK_LABEL } from "./moderation";
import { setSession } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const GAME = "11111111-2222-3333-4444-555555555555";
const HOST = "p-host";
const BOB = "p-bob";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
  localStorage.clear();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    for (let j = 0; j < 5; j++) await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

interface Sent {
  url: string;
  method: string;
}
let sent: Sent[];

function stubServer(handler: (url: string, method: string) => Response | undefined): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      sent.push({ url: input, method });
      return handler(input, method) ?? new Response("{}", { status: 200 });
    }),
  );
}

const json = (body: unknown, status = 200): Response =>
  new Response(JSON.stringify(body), { status });

describe("the lobby's Remove confirm", () => {
  const meta = {
    id: GAME,
    name: "Friday",
    created_at: new Date().toISOString(),
    state: "lobby",
    is_creator: true,
    players: [
      { player_id: HOST, name: "Alice", seat: 0, deck_uploaded: true, is_host: true },
      { player_id: BOB, name: "Bob", seat: 1, deck_uploaded: false },
    ],
  };

  beforeEach(() => {
    sent = [];
    location.hash = "#/lobby";
    setSession({
      token: "t",
      expiresAt: new Date(Date.now() + 86_400_000).toISOString(),
      principal: {
        role: "identified",
        user_id: "6f9619ff-8b86-d011-b42d-00c04fc964ff",
        name: "Alice",
        issued_at: new Date().toISOString(),
        expires_at: new Date(Date.now() + 86_400_000).toISOString(),
      },
    } as never);
    stubServer((url) => {
      if (url === "/games") return json({ games: [meta] });
      if (url === "/me") return json({ admin: false });
      if (url === "/me/games") return json({ games: [] });
      if (url.startsWith(`/games/${GAME}/seats/`)) return json(meta);
      return undefined;
    });
  });

  async function openConfirm(): Promise<HTMLElement> {
    const { container } = render(Lobby as never, {} as never);
    await settle();
    click(container.querySelector<HTMLButtonElement>('button[aria-label="remove Bob"]')!);
    flushSync();
    return container;
  }

  it("offers 'also block them', unticked, with the limit stated", async () => {
    const c = await openConfirm();
    const box = c.querySelector<HTMLInputElement>(".confirm input[type=checkbox]")!;
    expect(box).not.toBeNull();
    expect(box.checked).toBe(false);
    expect(c.querySelector(".confirm")?.textContent).toContain(BLOCK_LABEL);
    expect(c.querySelector(".confirm")?.textContent).toContain(BLOCK_HINT);
  });

  it("removes without a block unless the box is ticked", async () => {
    const c = await openConfirm();
    click(
      [...c.querySelectorAll<HTMLButtonElement>(".confirm button")].find((b) =>
        b.textContent?.includes("remove from table"),
      )!,
    );
    await settle();
    expect(sent.filter((s) => s.method === "DELETE")).toEqual([
      { url: `/games/${GAME}/seats/${BOB}`, method: "DELETE" },
    ]);
  });

  it("removes with ?block=true when the box is ticked", async () => {
    const c = await openConfirm();
    const box = c.querySelector<HTMLInputElement>(".confirm input[type=checkbox]")!;
    box.click();
    flushSync();
    click(
      [...c.querySelectorAll<HTMLButtonElement>(".confirm button")].find((b) =>
        b.textContent?.includes("remove from table"),
      )!,
    );
    await settle();
    expect(sent.filter((s) => s.method === "DELETE")).toEqual([
      { url: `/games/${GAME}/seats/${BOB}?block=true`, method: "DELETE" },
    ]);
  });
});

describe("the Join page's refusals", () => {
  beforeEach(() => {
    sent = [];
  });

  async function tryJoin(status: number, body: unknown): Promise<string> {
    stubServer((url, method) => {
      if (url === "/auth/discord/config") return json({ enabled: false });
      if (url.startsWith(`/games/${GAME}/preview`)) {
        return json({
          game: { id: GAME, name: "Friday", players: [], state: "lobby" },
          invite: "player",
          max_seats: 4,
        });
      }
      if (method === "POST" && url === `/games/${GAME}/join`) return json(body, status);
      return undefined;
    });
    const { container } = render(
      Join as never,
      { gameID: GAME, inviteToken: "inv", spectator: false } as never,
    );
    await settle();
    const input = container.querySelector<HTMLInputElement>("input[type=text], input:not([type])")!;
    input.value = "Bob";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    return container.querySelector('[role="alert"].error, p.error')?.textContent?.trim() ?? "";
  }

  it("shows the server's sentence for a block", async () => {
    const msg = "you were removed from this table and can't take a seat here again";
    expect(await tryJoin(403, { error: msg, code: "blocked_from_table" })).toBe(msg);
  });

  it("shows the server's sentence for a name on the list", async () => {
    const msg = "that name isn't allowed here; pick another";
    expect(await tryJoin(400, { error: msg, code: "name_not_allowed" })).toBe(msg);
  });
});
