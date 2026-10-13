// @vitest-environment jsdom
//
// blockList.render.test.ts — #3113, rendered: the block list with its
// Unblock buttons (by label, never a Discord id; quiet when the server
// refuses; an empty state only where asked), the ⋯ menu's spectators in
// the report form, and the Spectators panel with its Remove and Block
// buttons, and the admin seat list's "removed".

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import BlockList from "./components/BlockList.svelte";
import GameMenu from "./components/board/GameMenu.svelte";
import SeatList from "./components/admin/SeatList.svelte";
import type { BlockEntry, SpectatorLine } from "./api";
import type { GameMenuOptions } from "./gameMenu";
import { _resetForTests as resetModals } from "./modalLayers";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const GAME = "9c2f0e4c-1d1e-4c3b-9d7e-2f8a1b2c3d4e";

interface Call {
  url: string;
  method: string;
}
let calls: Call[];
let listStatus: number;
let blocks: BlockEntry[];

function json(body: unknown, code = 200): Response {
  return new Response(JSON.stringify(body), { status: code });
}

beforeEach(() => {
  calls = [];
  listStatus = 200;
  blocks = [
    { id: "e1", label: "Dave", kind: "account", created_at: 1 },
    { id: "e2", label: "bob", kind: "guest", created_at: 2 },
  ];
  resetModals();
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      calls.push({ url, method });
      if (listStatus !== 200) return json({ error: "no" }, listStatus);
      if (url === `/games/${GAME}/blocks` && method === "GET") return json({ blocks });
      if (url.startsWith(`/games/${GAME}/blocks/`) && method === "DELETE") {
        const id = decodeURIComponent(url.split("/").at(-1)!);
        blocks = blocks.filter((b) => b.id !== id);
        return json({ blocks });
      }
      return json({ error: "nope" }, 404);
    }),
  );
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

const text = (el: Element | null | undefined): string =>
  (el?.textContent ?? "").replace(/\s+/g, " ");

describe("the block list", () => {
  it("lists each block by its label with what it holds against, and an Unblock", async () => {
    const { container } = render(BlockList as never, { gameID: GAME } as never);
    await settle();
    expect(calls).toEqual([{ url: `/games/${GAME}/blocks`, method: "GET" }]);
    const rows = [...container.querySelectorAll("li")];
    expect(rows).toHaveLength(2);
    expect(text(rows[0])).toContain("Dave");
    expect(text(rows[0])).toContain("account");
    expect(text(rows[1])).toContain("bob");
    expect(text(rows[1])).toContain("guest name");
    expect(container.querySelectorAll("button.unblock")).toHaveLength(2);
    // The hint says what a guest block is worth.
    expect(text(container)).toContain("holds against the name only");
  });

  it("unblocks one with DELETE and takes it off the list", async () => {
    const { container } = render(BlockList as never, { gameID: GAME } as never);
    await settle();
    click(container.querySelector<HTMLButtonElement>('button[aria-label="Unblock bob"]')!);
    await settle();
    expect(calls.at(-1)).toEqual({ url: `/games/${GAME}/blocks/e2`, method: "DELETE" });
    const rows = [...container.querySelectorAll("li")];
    expect(rows).toHaveLength(1);
    expect(text(rows[0])).toContain("Dave");
  });

  it("reads the list again when the host removes and blocks someone", async () => {
    const { container, setProps } = render(
      BlockList as never,
      {
        gameID: GAME,
        refreshKey: 0,
      } as never,
    );
    await settle();
    blocks = [...blocks, { id: "e3", label: "carol", kind: "guest", created_at: 3 }];
    setProps({ refreshKey: 1 } as never);
    await settle();
    expect(container.querySelectorAll("li")).toHaveLength(3);
    expect(calls.filter((c) => c.method === "GET")).toHaveLength(2);
  });

  it("draws nothing for an empty list on a waiting table, and says so where asked", async () => {
    blocks = [];
    const quiet = render(BlockList as never, { gameID: GAME } as never);
    await settle();
    expect(quiet.container.querySelector(".blocks")).toBeNull();
    cleanup();

    const asked = render(BlockList as never, { gameID: GAME, showEmpty: true } as never);
    await settle();
    expect(text(asked.container)).toContain("Nobody is blocked from this table.");
  });

  it("stays silent when the server refuses the list (a non-manager, a started table)", async () => {
    for (const code of [403, 404, 409]) {
      listStatus = code;
      const { container } = render(BlockList as never, { gameID: GAME, showEmpty: true } as never);
      await settle();
      expect(container.querySelector(".blocks")).toBeNull();
      cleanup();
    }
  });
});

// ---- the ⋯ menu ------------------------------------------------------------

const WATCHERS: SpectatorLine[] = [
  { id: "s1", name: "Eve", signed_in: true, since: 1 },
  { id: "s2", name: "Spectator", since: 2 },
];

function mountMenu(over: Partial<GameMenuOptions> = {}) {
  const reports: unknown[][] = [];
  const removed: unknown[][] = [];
  const r = render(
    GameMenu as never,
    {
      seated: true,
      canManage: true,
      spawnAvailable: false,
      myGames: false,
      voteOpen: false,
      onDraw: () => {},
      onUntapAll: () => {},
      onShuffle: () => {},
      onMulligan: () => {},
      onLifeHistory: () => {},
      onTableSettings: () => {},
      onSpawn: () => {},
      onMyGames: () => {},
      onBack: () => {},
      onConcede: () => {},
      onStartVote: () => {},
      reportTargets: [{ id: "p2", name: "Bob" }],
      onReport: async (...args: unknown[]) => {
        reports.push(args);
      },
      loadSpectators: async () => WATCHERS,
      canRemoveSpectators: true,
      onRemoveSpectator: async (...args: unknown[]) => {
        removed.push(args);
        return { blocked: args[1] === true && args[0] === "s1", spectators: [WATCHERS[1]] };
      },
      ...over,
    } as never,
  );
  const trigger = r.container.querySelector<HTMLButtonElement>(
    'button[aria-label="more actions"]',
  )!;
  return { ...r, trigger, reports, removed };
}

const menuItem = (c: ParentNode, name: string): HTMLElement | undefined =>
  [...c.querySelectorAll<HTMLElement>('[role="menuitem"]')].find((el) =>
    (el.textContent ?? "").replace(/\s+/g, " ").trim().startsWith(name),
  );

function type(el: HTMLTextAreaElement | HTMLSelectElement, value: string): void {
  el.value = value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
  el.dispatchEvent(new Event("change", { bubbles: true }));
  flushSync();
}

describe("reporting a spectator from the ⋯ menu", () => {
  it("offers the people watching beside the seats, and reports one as a spectator", async () => {
    const m = mountMenu();
    click(m.trigger);
    await settle();
    click(menuItem(m.container, "Report a player…")!);
    await settle();
    const form = m.container.querySelector<HTMLFormElement>("form[aria-label='report a player']")!;
    const groups = [...form.querySelectorAll("optgroup")].map((g) => g.getAttribute("label"));
    expect(groups).toEqual(["Players", "Spectators"]);
    const options = [...form.querySelectorAll("option")].map((o) => [o.value, o.textContent]);
    expect(options).toEqual([
      ["p2", "Bob"],
      ["s1", "Eve"],
      ["s2", "Spectator"],
    ]);
    type(form.querySelector("select")!, "s1");
    type(form.querySelector("textarea")!, "  spamming chat ");
    form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(m.reports).toEqual([["s1", "spamming chat", "spectator"]]);
    expect(text(m.container)).toContain("Report sent. Thank you.");
  });

  it("still reports a seat with two arguments, and offers the entry for spectators alone", async () => {
    const m = mountMenu({ reportTargets: [] });
    click(m.trigger);
    await settle();
    expect(menuItem(m.container, "Report a player…")).toBeDefined();
    click(menuItem(m.container, "Report a player…")!);
    await settle();
    const form = m.container.querySelector<HTMLFormElement>("form[aria-label='report a player']")!;
    // The spectators arrive after the form opens; with no seats the first becomes the target.
    expect(form.querySelector("select")!.value).toBe("s1");
    cleanup();

    const seats = mountMenu({ loadSpectators: async () => [] });
    click(seats.trigger);
    await settle();
    click(menuItem(seats.container, "Report a player…")!);
    await settle();
    const f2 = seats.container.querySelector<HTMLFormElement>(
      "form[aria-label='report a player']",
    )!;
    expect(f2.querySelectorAll("optgroup")).toHaveLength(0);
    type(f2.querySelector("textarea")!, "rude");
    f2.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(seats.reports).toEqual([["p2", "rude"]]);
  });
});

describe("the Spectators panel", () => {
  async function openPanel(m: ReturnType<typeof mountMenu>): Promise<HTMLElement> {
    click(m.trigger);
    await settle();
    click(menuItem(m.container, "Who's watching…")!);
    await settle();
    return m.container.querySelector<HTMLElement>("[aria-label='spectators']")!;
  }

  it("is offered to the host and admin, and to nobody else", async () => {
    const yes = mountMenu();
    click(yes.trigger);
    await settle();
    expect(menuItem(yes.container, "Who's watching…")).toBeDefined();
    cleanup();

    for (const over of [{ canRemoveSpectators: false }, { onRemoveSpectator: undefined }]) {
      const no = mountMenu(over as Partial<GameMenuOptions>);
      click(no.trigger);
      await settle();
      expect(menuItem(no.container, "Who's watching…")).toBeUndefined();
      cleanup();
    }
  });

  it("lists who is watching with a Remove and a Remove and block each", async () => {
    const m = mountMenu();
    const panel = await openPanel(m);
    const rows = [...panel.querySelectorAll("li")];
    expect(rows).toHaveLength(2);
    expect(text(rows[0])).toContain("Eve");
    expect(text(rows[0])).toContain("signed in");
    expect(rows[1].querySelector('button[aria-label="Remove and block Spectator"]')).not.toBeNull();
    // The panel says what a guest with no name is worth.
    expect(text(panel)).toContain("guest with no name cannot be blocked");
  });

  it("removes one without a block, then with one, and says what happened", async () => {
    const m = mountMenu();
    const panel = await openPanel(m);
    click(panel.querySelector<HTMLButtonElement>('button[aria-label="Remove Eve"]')!);
    await settle();
    expect(m.removed).toEqual([["s1", false]]);
    expect(text(panel)).toContain("Removed.");
    // The answer's list replaces the old one.
    expect(panel.querySelectorAll("li")).toHaveLength(1);

    click(
      panel.querySelector<HTMLButtonElement>('button[aria-label="Remove and block Spectator"]')!,
    );
    await settle();
    expect(m.removed.at(-1)).toEqual(["s2", true]);
    // The server said it blocked nothing (a guest with no name): the panel says so.
    expect(text(panel)).toContain("nothing to block");
  });

  it("says a block took when the server says it did", async () => {
    const m = mountMenu();
    const panel = await openPanel(m);
    click(panel.querySelector<HTMLButtonElement>('button[aria-label="Remove and block Eve"]')!);
    await settle();
    expect(m.removed).toEqual([["s1", true]]);
    expect(text(panel)).toContain("Removed and blocked from this table.");
  });

  it("shows a refusal and reads the list again", async () => {
    const { LobbyApiError } = await import("./session");
    let loads = 0;
    const m = mountMenu({
      loadSpectators: async () => {
        loads += 1;
        return WATCHERS;
      },
      onRemoveSpectator: async () => {
        throw new LobbyApiError(404, "gone");
      },
    });
    const panel = await openPanel(m);
    click(panel.querySelector<HTMLButtonElement>('button[aria-label="Remove Eve"]')!);
    await settle();
    expect(text(panel)).toMatch(/isn't watching any more/);
    expect(loads).toBe(2);
  });
});

// ---- the admin seat list --------------------------------------------------

describe("the admin seat list reads a removed seat as removed", () => {
  it("shows 'removed' where a disconnected seat would say 'not connected'", () => {
    const seats = [
      { seat: 0, kind: "human", guest_name: "Alice", host: true, connected: 1 },
      { seat: 1, kind: "human", guest_name: "Bob", host: false, connected: 0, removed: true },
      { seat: 2, kind: "human", guest_name: "Cat", host: false, connected: 0 },
    ];
    const { container } = render(
      SeatList as never,
      {
        seats,
        now: Date.now(),
        label: "seats",
      } as never,
    );
    const rows = [...container.querySelectorAll("li.seat")];
    expect(text(rows[1])).toContain("removed");
    expect(text(rows[1])).not.toContain("not connected");
    expect(text(rows[2])).toContain("not connected");
    expect(text(rows[2])).not.toContain("removed");
    expect(text(rows[0])).not.toContain("removed");
  });
});
