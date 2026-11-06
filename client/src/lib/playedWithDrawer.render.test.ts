// @vitest-environment jsdom
//
// ADR 0213 §1, slice 2: the Played with drawer. It fetches the list when
// it opens and never again, offers Invite on each row (a Discord DM where
// the server has a bot, Copy invite where it has not), and the invite goes
// to one of your lobby tables with an empty seat or to a New table.
//
// Slice 3 makes it the friends drawer: Requests (Accept, Ignore) and
// Friends above Played with, Add friend on each Played with row, and a
// link to #/friends. With no friends store (GET /me/friends refused) it is
// Played with alone, which is what the slice 2 cases below run against.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { writable } from "svelte/store";

import type { Route } from "./router";

vi.mock("./router", async (importOriginal) => {
  const orig = await importOriginal<typeof import("./router")>();
  return { ...orig, route: writable<Route>({ name: "decks" } as Route), navigate: vi.fn() };
});

import PlayedWithDrawer from "./components/PlayedWithDrawer.svelte";
import SiteHeader from "./components/SiteHeader.svelte";
import { _resetCommunityForTests, loadCommunity } from "./community";
import { resetBugReportConfig } from "./bugReportConfig";
import { L } from "./labels";
import { setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const BOB = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const CARA = "7a1d3c10-0f6e-4d46-8d1b-3a1f4d7a9c22";

function sess(role: Session["principal"]["role"], userID?: string): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: `${role}-tok`,
    expiresAt: expires,
    principal: {
      role,
      user_id: userID,
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
  } as Session;
}

interface Call {
  method: string;
  path: string;
  body?: unknown;
}
let calls: Call[] = [];
let dmInvites = true;
let dmStatus = 200;
let mates: unknown[] = [];
// GET /me/friends' body, or null for a server with no friends store.
let friendsBody: unknown = null;
let askOutcome = "sent";

const lobbyTable = {
  id: "g-lobby",
  name: "Friday Pod",
  created_at: "2026-11-05T00:00:00Z",
  state: "lobby",
  is_creator: true,
  players: [{ player_id: "p1", name: "Alice", seat: 0 }],
};
const fullTable = {
  ...lobbyTable,
  id: "g-full",
  name: "Full house",
  players: [0, 1, 2, 3].map((i) => ({ player_id: `f${i}`, name: `F${i}`, seat: i })),
};

function reply(status: number, payload: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? "OK" : "Error",
    json: async () => payload,
    clone() {
      return this;
    },
  } as unknown as Response;
}

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      const body = init.body ? JSON.parse(init.body as string) : undefined;
      calls.push({ method, path: input, body });
      if (input === "/community")
        return reply(200, { invite_url: "", gate: false, dm_invites: dmInvites });
      if (input === "/me/tablemates") return reply(200, { tablemates: mates });
      if (input === "/me/friends") {
        return friendsBody
          ? reply(200, friendsBody)
          : reply(503, { error: "friends are not available on this server" });
      }
      if (input === "/me/friends/requests" && method === "POST")
        return reply(200, { outcome: askOutcome, user_id: body?.user_id, display_name: "Bob" });
      if (input.startsWith("/me/friends/requests/") && input.endsWith("/accept"))
        return reply(200, { outcome: "friends" });
      if (input.startsWith("/me/friends/requests/") && input.endsWith("/ignore"))
        return reply(204, null);
      if (input === "/games" && method === "GET")
        return reply(200, { games: [lobbyTable, fullTable] });
      if (input === "/games" && method === "POST") {
        return reply(200, {
          ...lobbyTable,
          id: "g-new",
          name: "Quiet Hollow",
          invite_token: "new-token",
          players: [{ player_id: "p9", name: "Alice", seat: 0 }],
        });
      }
      if (input === "/games/g-lobby")
        return reply(200, { ...lobbyTable, invite_token: "lobby-token" });
      if (input.endsWith("/invites/dm")) {
        return dmStatus === 200
          ? reply(200, { sent: true })
          : reply(dmStatus, { error: "that person has DMs closed" });
      }
      return reply(404, { error: "unexpected " + input });
    }),
  );
}

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

function inviteButton(c: HTMLElement, name = "Bob"): HTMLButtonElement {
  const row = [...c.querySelectorAll<HTMLElement>(".mate")].find((r) =>
    r.textContent?.includes(name),
  );
  return row!.querySelector<HTMLButtonElement>("button.invite")!;
}

function choice(c: HTMLElement, text: string): HTMLButtonElement | undefined {
  return [...c.querySelectorAll<HTMLButtonElement>(".choices button")].find((b) =>
    b.textContent?.includes(text),
  );
}

beforeEach(async () => {
  calls = [];
  dmInvites = true;
  dmStatus = 200;
  friendsBody = null;
  askOutcome = "sent";
  mates = [
    {
      user_id: BOB,
      display_name: "Bob",
      last_played_at: Date.now() - 2 * 86_400_000,
      games_together: 3,
      last_table: "FNM",
    },
    {
      user_id: CARA,
      display_name: "Cara",
      last_played_at: Date.now() - 40 * 86_400_000,
      games_together: 1,
      last_table: "Old lobby",
    },
  ];
  _resetCommunityForTests();
  resetBugReportConfig();
  stubServer();
  setSession(sess("identified", USER));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
});

async function open(): Promise<{ container: HTMLElement; closed: () => number }> {
  await loadCommunity();
  let closes = 0;
  const { container } = render(PlayedWithDrawer as never, { onclose: () => closes++ } as never);
  await settle();
  return { container, closed: () => closes };
}

describe("the Played with drawer", () => {
  it("fetches the list and your tables when it opens, and does not poll", async () => {
    const { container } = await open();
    const reads = () =>
      calls.filter(
        (c) => c.path === "/me/tablemates" || (c.path === "/games" && c.method === "GET"),
      );
    expect(
      reads()
        .map((c) => c.path)
        .sort(),
    ).toEqual(["/games", "/me/tablemates"]);
    const rows = container.querySelectorAll(".mate");
    expect(rows).toHaveLength(2);
    expect(rows[0].textContent).toContain("Bob");
    expect(rows[0].textContent).toContain("3 tables together · last: FNM · 2 days ago");

    await new Promise((r) => setTimeout(r, 50));
    await settle();
    expect(reads()).toHaveLength(2);
  });

  it("is a dialog named Friends list, with a Played with section", async () => {
    const { container } = await open();
    const dialog = container.querySelector('[role="dialog"]');
    expect(dialog?.getAttribute("aria-label")).toBe(L.friendsDrawer);
    expect(container.querySelector("#fr-played")?.textContent).toBe(L.playedWith);
    expect(dialog?.getAttribute("aria-modal")).toBe("true");
  });

  it("says so when nobody has sat with you yet", async () => {
    mates = [];
    const { container } = await open();
    expect(container.querySelector(".mates")).toBeNull();
    expect(container.textContent).toContain("Nobody yet");
  });

  it("Escape closes it", async () => {
    const { closed } = await open();
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(closed()).toBe(1);
  });

  it("offers Invite with a bot: your lobby tables with an empty seat, and New table", async () => {
    const { container } = await open();
    const btn = inviteButton(container);
    expect(btn.textContent?.trim()).toBe("Invite");
    click(btn);
    flushSync();
    const labels = [...container.querySelectorAll(".choices button")].map((b) =>
      b.textContent?.trim(),
    );
    expect(labels).toEqual(["Friday Pod · 1/4", "New table"]);
    expect(btn.getAttribute("aria-expanded")).toBe("true");
  });

  it("sends the DM to the table you pick", async () => {
    const { container } = await open();
    click(inviteButton(container));
    flushSync();
    click(choice(container, "Friday Pod")!);
    await settle();
    const dm = calls.find((c) => c.path === "/games/g-lobby/invites/dm");
    expect(dm?.method).toBe("POST");
    expect(dm?.body).toEqual({ user_id: BOB });
    expect(container.querySelector(".result")?.textContent).toContain(
      "Invite to Friday Pod sent to Bob",
    );
    expect(container.querySelector(".choose")).toBeNull();
  });

  it("New table makes a table from your last setup, then invites into it", async () => {
    const { container } = await open();
    const changed = vi.fn();
    window.addEventListener("c2:tables-changed", changed);
    click(inviteButton(container));
    flushSync();
    click(choice(container, "New table")!);
    await settle();
    window.removeEventListener("c2:tables-changed", changed);
    const create = calls.find((c) => c.path === "/games" && c.method === "POST");
    expect(create?.body).toMatchObject({ name: "", setup: "last" });
    expect(calls.some((c) => c.path === "/games/g-new/invites/dm")).toBe(true);
    expect(changed).toHaveBeenCalledTimes(1);
    expect(container.querySelector(".result")?.textContent).toContain("Quiet Hollow");
  });

  it("shows what went wrong and leaves the chooser open to try again", async () => {
    dmStatus = 422;
    const { container } = await open();
    click(inviteButton(container));
    flushSync();
    click(choice(container, "Friday Pod")!);
    await settle();
    const result = container.querySelector(".result");
    expect(result?.textContent).toContain("that person has DMs closed");
    expect(result?.getAttribute("role")).toBe("alert");
    expect(container.querySelector(".choose")).not.toBeNull();
  });

  describe("where the server has no bot", () => {
    beforeEach(() => {
      dmInvites = false;
    });

    it("the row says Copy invite, and the link goes on the clipboard", async () => {
      const writeText = vi.fn(async () => undefined);
      vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText } });
      const { container } = await open();
      const btn = inviteButton(container);
      expect(btn.textContent?.trim()).toBe(L.copyInvite);
      click(btn);
      flushSync();
      click(choice(container, "Friday Pod")!);
      await settle();
      expect(writeText).toHaveBeenCalledTimes(1);
      const link = (writeText.mock.calls[0] as unknown as [string])[0];
      expect(link).toContain("#/games/g-lobby/join?t=lobby-token");
      expect(calls.some((c) => c.path.endsWith("/invites/dm"))).toBe(false);
      expect(container.querySelector(".result")?.textContent).toContain(
        "Invite link for Friday Pod copied for Bob",
      );
    });

    it("shows the link to copy by hand when the clipboard refuses", async () => {
      const writeText = vi.fn(async () => {
        throw new Error("denied");
      });
      vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText } });
      const { container } = await open();
      click(inviteButton(container));
      flushSync();
      click(choice(container, "Friday Pod")!);
      await settle();
      const input = container.querySelector<HTMLInputElement>("input.link");
      expect(input?.value).toContain("#/games/g-lobby/join?t=lobby-token");
    });

    it("New table uses the link the create answer carried", async () => {
      const writeText = vi.fn(async () => undefined);
      vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText } });
      const { container } = await open();
      click(inviteButton(container));
      flushSync();
      click(choice(container, "New table")!);
      await settle();
      const link = (writeText.mock.calls[0] as unknown as [string])[0];
      expect(link).toContain("#/games/g-new/join?t=new-token");
      expect(calls.some((c) => c.path === "/games/g-new")).toBe(false);
    });
  });
});

describe("the site header's Friends button", () => {
  function button(c: HTMLElement): HTMLButtonElement | null {
    return c.querySelector<HTMLButtonElement>(`button[aria-label="${L.friendsButton}"]`);
  }

  it("is offered to someone signed in with Discord, and opens the drawer without fetching first", async () => {
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    expect(calls.some((c) => c.path === "/me/tablemates")).toBe(false);
    click(button(container)!);
    await settle();
    expect(
      container.querySelector(`[role="dialog"][aria-label="${L.friendsDrawer}"]`),
    ).not.toBeNull();
    expect(calls.some((c) => c.path === "/me/tablemates")).toBe(true);
  });

  it("is not offered to a guest, the admin token, or a signed-out visitor", async () => {
    for (const s of [sess("player"), sess("admin"), null]) {
      setSession(s);
      const { container, destroy } = render(SiteHeader as never, {} as never);
      await settle();
      expect(button(container), s?.principal.role ?? "signed out").toBeNull();
      destroy();
    }
  });
});

describe("friends in the drawer (ADR 0213 slice 3)", () => {
  const DEV = "8b2e4f60-1c3d-4e5f-9a7b-6c5d4e3f2a10";
  const EVE = "9c3f5a71-2d4e-4f60-8b8c-7d6e5f4a3b21";

  beforeEach(() => {
    friendsBody = {
      friends: [{ user_id: DEV, display_name: "Dev", since: Date.now() - 5 * 86_400_000 }],
      incoming: [{ user_id: EVE, display_name: "Eve", since: Date.now() }],
      outgoing: [],
    };
    mates = [
      ...mates,
      { user_id: DEV, display_name: "Dev", last_played_at: Date.now(), friend: true },
      { user_id: EVE, display_name: "Eve", last_played_at: Date.now(), request: "incoming" },
    ];
  });

  function section(c: HTMLElement, id: string): HTMLElement {
    return c.querySelector<HTMLElement>(`#${id}`)!.closest("section")!;
  }
  function named(root: HTMLElement, text: string): HTMLButtonElement | undefined {
    return [...root.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => b.textContent?.trim() === text,
    );
  }

  it("lists Requests, then Friends, then Played with, each person once", async () => {
    const { container } = await open();
    const heads = [...container.querySelectorAll("h3")].map((h) => h.id);
    expect(heads).toEqual(["fr-requests", "fr-friends", "fr-played"]);
    const names = (id: string) =>
      [...section(container, id).querySelectorAll(".name")].map((n) => n.textContent);
    expect(names("fr-requests")).toEqual(["Eve"]);
    expect(names("fr-friends")).toEqual(["Dev"]);
    expect(names("fr-played")).toEqual(["Bob", "Cara"]);
    expect(section(container, "fr-friends").textContent).toContain("friends since 5 days ago");
    const manage = [...container.querySelectorAll("a")].find(
      (a) => a.textContent?.trim() === L.manageFriends,
    );
    expect(manage?.getAttribute("href")).toBe("#/friends");
  });

  it("a friend can be invited like anyone you played with", async () => {
    const { container } = await open();
    click(inviteButton(container, "Dev"));
    flushSync();
    click(choice(container, "Friday Pod")!);
    await settle();
    expect(calls.find((c) => c.path.endsWith("/invites/dm"))?.body).toEqual({ user_id: DEV });
  });

  it("Accept answers the request and reads the lists again", async () => {
    const { container } = await open();
    const before = calls.filter((c) => c.path === "/me/friends").length;
    click(named(section(container, "fr-requests"), L.acceptFriend)!);
    await settle();
    expect(
      calls.some((c) => c.method === "POST" && c.path === `/me/friends/requests/${EVE}/accept`),
    ).toBe(true);
    expect(calls.filter((c) => c.path === "/me/friends").length).toBe(before + 1);
  });

  it("Ignore takes the request off the list without telling anyone", async () => {
    const { container } = await open();
    click(named(section(container, "fr-requests"), L.ignoreFriend)!);
    await settle();
    expect(calls.some((c) => c.path === `/me/friends/requests/${EVE}/ignore`)).toBe(true);
    expect(container.querySelector("#fr-requests")).toBeNull();
  });

  it("Add friend on a Played with row sends a request and says so", async () => {
    const { container } = await open();
    const played = section(container, "fr-played");
    const bobRow = [...played.querySelectorAll<HTMLElement>(".mate")].find((r) =>
      r.textContent?.includes("Bob"),
    )!;
    click(named(bobRow, L.addFriend)!);
    await settle();
    expect(calls.find((c) => c.path === "/me/friends/requests")?.body).toEqual({ user_id: BOB });
    const after = [...section(container, "fr-played").querySelectorAll<HTMLElement>(".mate")].find(
      (r) => r.textContent?.includes("Bob"),
    )!;
    expect(after.textContent).toContain("Friend request sent to Bob");
    expect(after.textContent).toContain("Request sent");
    expect(named(after, L.addFriend)).toBeUndefined();
  });

  it("says so when you have no friends yet", async () => {
    friendsBody = { friends: [], incoming: [], outgoing: [] };
    const { container } = await open();
    expect(section(container, "fr-friends").textContent).toContain("No friends yet");
    expect(container.querySelector("#fr-requests")).toBeNull();
  });

  it("with no friends store it is Played with alone, with no Add friend", async () => {
    friendsBody = null;
    const { container } = await open();
    expect(container.querySelector("#fr-friends")).toBeNull();
    expect(named(container, L.addFriend)).toBeUndefined();
    expect(
      [...container.querySelectorAll("a")].some((a) => a.textContent?.includes(L.manageFriends)),
    ).toBe(false);
  });
});
