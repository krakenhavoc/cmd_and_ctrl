// @vitest-environment jsdom
//
// ADR 0213 §9, slice 3: the #/friends page (your friends, requests sent
// and waiting, people you played with, blocks, and the two settings) and
// Add friend on another person's seat at your table.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Friends from "../routes/Friends.svelte";
import SeatTile from "./components/SeatTile.svelte";
import { L } from "./labels";
import { setSession, type Session } from "./session";
import { seatTileText } from "./tablesPage";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const BOB = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const CARA = "7a1d3c10-0f6e-4d46-8d1b-3a1f4d7a9c22";
const DEV = "8b2e4f60-1c3d-4e5f-9a7b-6c5d4e3f2a10";
const ZED = "9c3f5a71-2d4e-4f60-8b8c-7d6e5f4a3b21";

function sess(role: Session["principal"]["role"], userID?: string): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: `${role}-tok`,
    expiresAt: expires,
    principal: { role, user_id: userID, name: "Alice", issued_at: expires, expires_at: expires },
  } as Session;
}

function reply(status: number, payload: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: "x",
    json: async () => payload,
    clone() {
      return this;
    },
  } as unknown as Response;
}

let calls: { method: string; path: string; body?: unknown }[] = [];
let friendsStatus = 200;

beforeEach(() => {
  calls = [];
  friendsStatus = 200;
  setSession(sess("identified", USER));
  vi.stubGlobal(
    "fetch",
    vi.fn(async (path: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      const body = init.body ? JSON.parse(init.body as string) : undefined;
      calls.push({ method, path, body });
      if (path === "/me/friends") {
        if (friendsStatus !== 200) return reply(friendsStatus, { error: "off" });
        return reply(200, {
          friends: [{ user_id: BOB, display_name: "Bob", since: Date.now() }],
          incoming: [{ user_id: DEV, display_name: "Dev", since: Date.now() }],
          outgoing: [{ user_id: CARA, display_name: "Cara", since: Date.now() }],
          max_friends: 200,
          max_outgoing: 20,
        });
      }
      if (path === "/me/blocks")
        return reply(200, { blocks: [{ user_id: ZED, display_name: "Zed", since: Date.now() }] });
      if (path === "/me/tablemates")
        return reply(200, {
          tablemates: [
            { user_id: BOB, display_name: "Bob", last_played_at: Date.now(), friend: true },
            {
              user_id: CARA,
              display_name: "Cara",
              last_played_at: Date.now(),
              request: "outgoing",
            },
          ],
        });
      if (path === "/me/social" && method === "GET")
        return reply(200, { presence: "friends", requests: "played_with" });
      if (path === "/me/social" && method === "PUT")
        return reply(200, { presence: "friends", requests: "nobody", ...body });
      if (path === "/me/friends/requests" && method === "POST")
        return reply(200, { outcome: "sent", display_name: "Bea" });
      if (method === "DELETE" || method === "PUT") return reply(204, null);
      return reply(404, { error: "unexpected " + path });
    }),
  );
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

function section(c: HTMLElement, id: string): HTMLElement {
  return c.querySelector<HTMLElement>(`#${id}`)!.closest("section")!;
}
function button(root: HTMLElement, text: string): HTMLButtonElement | undefined {
  return [...root.querySelectorAll<HTMLButtonElement>("button")].find(
    (b) => b.textContent?.trim() === text,
  );
}

describe("the #/friends page", () => {
  it("lists friends, requests both ways, people you played with and blocks", async () => {
    const { container } = render(Friends as never, {} as never);
    await settle();
    expect(section(container, "f-friends").textContent).toContain("Bob");
    expect(section(container, "f-friends").textContent).toContain("1/200");
    const req = section(container, "f-requests").textContent!;
    expect(req).toContain("Waiting on you");
    expect(req).toContain("Dev");
    expect(req).toContain("You sent");
    expect(req).toContain("Cara");
    // A friend is not offered again under played with.
    const played = section(container, "f-played").textContent!;
    expect(played).toContain("Cara");
    expect(played).not.toContain("Bob");
    expect(section(container, "f-blocked").textContent).toContain("Zed");
  });

  it("asks before blocking, says what it ends, then blocks", async () => {
    const { container } = render(Friends as never, {} as never);
    await settle();
    const friends = section(container, "f-friends");
    click(button(friends, L.blockPerson)!);
    flushSync();
    expect(friends.textContent).toContain("It ends any friendship and request between you");
    expect(calls.some((c) => c.method === "PUT")).toBe(false);
    click(button(friends, L.blockPerson)!);
    await settle();
    expect(calls.some((c) => c.method === "PUT" && c.path === `/me/blocks/${BOB}`)).toBe(true);
    expect(container.textContent).toContain("You blocked Bob.");
  });

  it("Keep backs out of a removal", async () => {
    const { container } = render(Friends as never, {} as never);
    await settle();
    const friends = section(container, "f-friends");
    click(button(friends, L.removeFriend)!);
    flushSync();
    click(button(friends, "Keep")!);
    flushSync();
    expect(calls.some((c) => c.method === "DELETE")).toBe(false);
    expect(button(friends, L.removeFriend)).toBeDefined();
  });

  it("cancels a request you sent, and unblocks", async () => {
    const { container } = render(Friends as never, {} as never);
    await settle();
    click(button(section(container, "f-requests"), L.cancelFriendRequest)!);
    await settle();
    expect(
      calls.some((c) => c.method === "DELETE" && c.path === `/me/friends/requests/${CARA}`),
    ).toBe(true);
    click(button(section(container, "f-blocked"), L.unblockPerson)!);
    await settle();
    expect(calls.some((c) => c.method === "DELETE" && c.path === `/me/blocks/${ZED}`)).toBe(true);
  });

  it("saves a setting as soon as it is chosen", async () => {
    const { container } = render(Friends as never, {} as never);
    await settle();
    const group = container.querySelector(`[role="radiogroup"][aria-label="${L.whoCanRequest}"]`)!;
    const nobody = [...group.querySelectorAll<HTMLInputElement>("input")].find(
      (i) => i.value === "nobody",
    )!;
    nobody.checked = true;
    nobody.dispatchEvent(new Event("change", { bubbles: true }));
    await settle();
    expect(calls.find((c) => c.method === "PUT" && c.path === "/me/social")?.body).toEqual({
      requests: "nobody",
    });
    expect(
      container.querySelector(`[role="radiogroup"][aria-label="${L.whoCanSee}"]`),
    ).not.toBeNull();
  });

  it("says friends are off on a server with no friends store", async () => {
    friendsStatus = 503;
    const { container } = render(Friends as never, {} as never);
    await settle();
    expect(container.textContent).toContain("Friends aren’t available on this server.");
  });

  it("asks a guest seat or the admin token to sign in instead", async () => {
    for (const s of [sess("player"), sess("admin")]) {
      setSession(s);
      const { container, destroy } = render(Friends as never, {} as never);
      await settle();
      expect(container.textContent).toContain("Sign in with Discord to have friends.");
      destroy();
    }
    const social = ["/me/friends", "/me/blocks", "/me/social", "/me/tablemates"];
    expect(calls.filter((c) => social.includes(c.path))).toHaveLength(0);
  });
});

describe("Add friend on a seat", () => {
  const seat = {
    player_id: "p2",
    name: "Bea",
    seat: 1,
    discord_id: "123",
    display_name: "Bea",
    deck_uploaded: false,
  };

  it("is a button until asked, then says where it stands", () => {
    const onclick = vi.fn();
    const { container, destroy } = render(
      SeatTile as never,
      {
        seat: 1,
        p: seat,
        text: seatTileText(seat as never, undefined),
        befriend: { state: "none", disabled: false, onclick },
      } as never,
    );
    click(button(container, L.addFriend)!);
    expect(onclick).toHaveBeenCalledTimes(1);
    destroy();

    const after = render(
      SeatTile as never,
      {
        seat: 1,
        p: seat,
        text: seatTileText(seat as never, undefined),
        befriend: {
          state: "outgoing",
          disabled: false,
          note: "Friend request sent to Bea",
          onclick,
        },
      } as never,
    ).container;
    expect(button(after, L.addFriend)).toBeUndefined();
    expect(after.textContent).toContain("Request sent");
    expect(after.textContent).toContain("Friend request sent to Bea");
  });

  it("is not drawn without the prop", () => {
    const { container } = render(
      SeatTile as never,
      { seat: 1, p: seat, text: seatTileText(seat as never, undefined) } as never,
    );
    expect(button(container, L.addFriend)).toBeUndefined();
  });
});
