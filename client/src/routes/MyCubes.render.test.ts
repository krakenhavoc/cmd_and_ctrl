// @vitest-environment jsdom
//
// #5441 (ADR 0222 §8): the My cubes page. The list with each cube's
// check, a new cube, the cards that block one named and dropped, and a
// delete that asks first, drawn from /me/cubes as docs/lobby.md gives it.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import MyCubes from "./MyCubes.svelte";
import { L } from "../lib/labels";
import { sessionFromOAuth, setSession } from "../lib/session";
import { cleanup, flushSync, render } from "../lib/test/render.svelte";

const realFetch = globalThis.fetch;
const USER = "6f9619ff-8b86-4011-b42d-00c04fc964ff";
const READY = "11111111-0000-4000-8000-000000000000";
const BLOCKED = "22222222-0000-4000-8000-000000000000";

interface Call {
  url: string;
  method: string;
  body: Record<string, unknown> | undefined;
}
type Answer = { status?: number; body?: unknown };
let calls: Call[] = [];

function serve(routes: Record<string, Answer>): void {
  calls = [];
  globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init: RequestInit = {}) => {
    const url = String(input);
    const method = init.method ?? "GET";
    calls.push({
      url,
      method,
      body: init.body ? (JSON.parse(String(init.body)) as Record<string, unknown>) : undefined,
    });
    const a = routes[`${method} ${url}`] ?? {};
    const status = a.status ?? 200;
    return new Response(status === 204 ? null : JSON.stringify(a.body ?? {}), {
      status,
      headers: { "Content-Type": "application/json" },
    });
  }) as never;
}
const sent = (method: string, url: string): Call[] =>
  calls.filter((c) => c.method === method && c.url === url);

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

const text = (el: Element | null | undefined): string =>
  (el?.textContent ?? "").replace(/\s+/g, " ").trim();
const button = (c: Element, name: string): HTMLButtonElement => {
  const b = [...c.querySelectorAll<HTMLButtonElement>("button")].find((el) => text(el) === name);
  if (!b) throw new Error(`no button named "${name}"`);
  return b;
};
const click = (el: Element): void => {
  el.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
  flushSync();
};
function type(el: HTMLInputElement | HTMLTextAreaElement, value: string): void {
  el.value = value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

const LIMITS = {
  max_cubes: 20,
  max_cards: 1080,
  max_name_length: 60,
  pack_size: 15,
  min_pack_size: 9,
  max_pack_size: 20,
  packs_per_seat: 3,
};
const ready = {
  id: READY,
  name: "Pauper cube",
  singleton: true,
  check: {
    cards: 360,
    unknown: [],
    not_full: [],
    repeated: [],
    too_large: false,
    ready: true,
    max_seats: 8,
  },
  updated_at: "2026-11-11T19:02:11Z",
};
const blocked = {
  id: BLOCKED,
  name: "Vintage cube",
  singleton: true,
  check: {
    cards: 4,
    unknown: ["Typo Card"],
    not_full: ["Not Built"],
    repeated: ["Sol Ring"],
    too_large: false,
    ready: false,
    max_seats: 0,
  },
  updated_at: "2026-11-11T19:02:11Z",
};
const BLOCKED_LIST = "1 Not Built\n1 Typo Card\n2 Sol Ring\n";

beforeEach(() => {
  setSession({
    ...sessionFromOAuth({ token: "tok", expiresAt: "2099-01-01T00:00:00Z", userID: USER }),
  });
});

afterEach(() => {
  cleanup();
  setSession(null);
  globalThis.fetch = realFetch;
});

const rows = (c: Element): Element[] => [...c.querySelectorAll('[data-testid="my-cubes"] > li')];

describe("My cubes", () => {
  it("lists the caller's cubes with each one's check, and a ready one's way to a draft", async () => {
    serve({ "GET /me/cubes": { body: { cubes: [ready, blocked], ...LIMITS } } });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    expect(text(c.querySelector("h1"))).toBe(L.myCubes);
    const [a, b] = rows(c);
    expect(text(a)).toContain("Pauper cube");
    expect(text(a)).toContain("360 cards");
    expect(text(a)).toContain("Ready for up to 8 seats");
    const draft = [...a.querySelectorAll("a")].find((x) => text(x) === L.cubeDraft);
    expect(draft?.getAttribute("href")).toBe(`#/draft?cube=${READY}`);
    expect(text(b)).toContain("3 cards block it");
    // A cube that cannot be drafted has no link to a draft.
    expect([...b.querySelectorAll("a")].some((x) => text(x) === L.cubeDraft)).toBe(false);
  });

  it("says so when there are none, and states the sizes a draft needs", async () => {
    serve({ "GET /me/cubes": { body: { cubes: [], ...LIMITS } } });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    expect(text(c)).toContain("No cubes yet.");
    expect(text(c.querySelector('[data-testid="cube-sizes"]'))).toBe(
      "A draft deals 3 packs a seat: 90 cards for 2 seats and 360 for 8, with packs of 15. A cube holds at most 1080 cards, and every card has to be fully built before the cube can be drafted.",
    );
  });

  it("saves a pasted list as a new cube and shows its check", async () => {
    serve({
      "GET /me/cubes": { body: { cubes: [], ...LIMITS } },
      "POST /me/cubes": { status: 201, body: { cube: { ...blocked, list: BLOCKED_LIST } } },
    });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    // Nothing to save yet.
    expect(button(c, L.cubeSave).disabled).toBe(true);
    type(c.querySelector<HTMLInputElement>("#cube-name")!, "Vintage cube");
    type(c.querySelector<HTMLTextAreaElement>("#cube-list")!, BLOCKED_LIST);
    expect(c.querySelector<HTMLInputElement>('.form input[type="checkbox"]')!.checked).toBe(true);
    c.querySelector("form")!.dispatchEvent(
      new Event("submit", { bubbles: true, cancelable: true }),
    );
    await settle();
    expect(sent("POST", "/me/cubes")[0].body).toEqual({
      name: "Vintage cube",
      text: BLOCKED_LIST.trim(),
      singleton: true,
    });
    // The new cube is in the list, open, with the cards that block it named.
    const shown = c.querySelector('[data-testid="cube-check"]')!;
    expect(text(shown)).toContain("Not fully built yet 1");
    expect(text(shown)).toContain("Not Built");
    expect(text(shown)).toContain("Not recognised 1");
    expect(text(shown)).toContain("Typo Card");
    expect(text(shown)).toContain("More than one copy 1");
    expect(text(shown)).toContain("Sol Ring");
    expect(text(c.querySelector('[role="status"]'))).toContain("Saved Vintage cube");
  });

  it("saves a link in place of a list, with the name left to the deck", async () => {
    serve({
      "GET /me/cubes": { body: { cubes: [], ...LIMITS } },
      "POST /me/cubes": { status: 201, body: { cube: { ...ready, list: "1 A\n" } } },
    });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    const link = c.querySelector<HTMLInputElement>('input[name="cube-from"][value="url"]')!;
    link.checked = true;
    link.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    expect(c.querySelector("#cube-list")).toBeNull();
    type(c.querySelector<HTMLInputElement>("#cube-link")!, "https://archidekt.com/decks/1");
    c.querySelector("form")!.dispatchEvent(
      new Event("submit", { bubbles: true, cancelable: true }),
    );
    await settle();
    expect(sent("POST", "/me/cubes")[0].body).toEqual({
      url: "https://archidekt.com/decks/1",
      singleton: true,
    });
  });

  it("shows the server's sentence for a save it refused", async () => {
    serve({
      "GET /me/cubes": { body: { cubes: [], ...LIMITS } },
      "POST /me/cubes": {
        status: 409,
        body: { error: "You already have a cube with that name.", code: "cube_name_taken" },
      },
    });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    type(c.querySelector<HTMLInputElement>("#cube-name")!, "Pauper cube");
    type(c.querySelector<HTMLTextAreaElement>("#cube-list")!, "1 A");
    c.querySelector("form")!.dispatchEvent(
      new Event("submit", { bubbles: true, cancelable: true }),
    );
    await settle();
    expect(text(c.querySelector('.form [role="alert"]'))).toBe(
      "You already have a cube with that name.",
    );
  });

  it("checks a cube, then drops the cards that are not ready and keeps the repeated one", async () => {
    const after = {
      ...blocked,
      list: "2 Sol Ring\n",
      check: { ...blocked.check, cards: 2, unknown: [], not_full: [] },
    };
    serve({
      "GET /me/cubes": { body: { cubes: [blocked], ...LIMITS } },
      [`GET /me/cubes/${BLOCKED}`]: { body: { cube: { ...blocked, list: BLOCKED_LIST } } },
      [`PUT /me/cubes/${BLOCKED}`]: { body: { cube: after } },
    });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    expect(c.querySelector('[data-testid="cube-check"]')).toBeNull();
    click(button(c, L.cubeCheck));
    await settle();
    expect(sent("GET", `/me/cubes/${BLOCKED}`)).toHaveLength(1);
    expect(text(c.querySelector('[data-testid="cube-check"]'))).toContain(
      "This cube can’t be drafted until the cards below are dropped or fixed.",
    );
    click(button(c, L.cubeDropBlockers));
    await settle();
    expect(sent("PUT", `/me/cubes/${BLOCKED}`)[0].body).toEqual({ text: "2 Sol Ring" });
    const shown = text(c.querySelector('[data-testid="cube-check"]'));
    expect(shown).not.toContain("Not Built");
    expect(shown).toContain("More than one copy 1");
    expect(text(rows(c)[0])).toContain("1 card blocks it");
    // Nothing is left to drop: the repeated card is the owner's to cut.
    expect(() => button(c, L.cubeDropBlockers)).toThrow();
  });

  it("turns singleton off for a cube, and edits its list", async () => {
    const waived = {
      ...blocked,
      singleton: false,
      list: "2 Sol Ring\n",
      check: { ...ready.check, cards: 2, max_seats: 0 },
    };
    serve({
      "GET /me/cubes": { body: { cubes: [blocked], ...LIMITS } },
      [`GET /me/cubes/${BLOCKED}`]: { body: { cube: { ...blocked, list: BLOCKED_LIST } } },
      [`PUT /me/cubes/${BLOCKED}`]: { body: { cube: waived } },
    });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    click(button(c, L.cubeCheck));
    await settle();
    const box = c.querySelector<HTMLInputElement>(
      '[data-testid="cube-check"] input[type="checkbox"]',
    )!;
    expect(box.checked).toBe(true);
    box.checked = false;
    box.dispatchEvent(new Event("change", { bubbles: true }));
    await settle();
    expect(sent("PUT", `/me/cubes/${BLOCKED}`)[0].body).toEqual({ singleton: false });
    expect(text(rows(c)[0])).toContain("any number of copies");
    expect(text(c.querySelector('[data-testid="cube-check"]'))).toContain(
      "Every card is ready, but 2 cards is too few to draft.",
    );

    click(button(c, L.cubeEditList));
    const area = c.querySelector<HTMLTextAreaElement>(`#cube-edit-${BLOCKED}`)!;
    expect(area.value).toBe("2 Sol Ring\n");
    type(area, "2 Sol Ring\n1 Island\n");
    click(button(c, L.cubeSaveList));
    await settle();
    expect(sent("PUT", `/me/cubes/${BLOCKED}`)[1].body).toEqual({ text: "2 Sol Ring\n1 Island" });
  });

  it("asks before it deletes a cube", async () => {
    serve({
      "GET /me/cubes": { body: { cubes: [ready], ...LIMITS } },
      [`DELETE /me/cubes/${READY}`]: { status: 204 },
    });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    click(button(c, L.cubeDelete));
    expect(sent("DELETE", `/me/cubes/${READY}`)).toHaveLength(0);
    expect(text(rows(c)[0])).toContain("Delete Pauper cube? This can’t be undone.");
    click(button(c, "Keep it"));
    expect(rows(c)).toHaveLength(1);
    click(button(c, L.cubeDelete));
    click(button(c, L.cubeDeleteConfirm));
    await settle();
    expect(sent("DELETE", `/me/cubes/${READY}`)).toHaveLength(1);
    expect(rows(c)).toHaveLength(0);
    expect(text(c)).toContain("No cubes yet.");
  });

  it("has no form at the cap, and says how to make room", async () => {
    const many = Array.from({ length: 20 }, (_, i) => ({
      ...ready,
      id: `${i}`.padStart(8, "0") + "-0000-4000-8000-000000000000",
      name: `Cube ${i}`,
    }));
    serve({ "GET /me/cubes": { body: { cubes: many, ...LIMITS } } });
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    expect(text(c.querySelector('[data-testid="cubes-full"]'))).toBe(
      "You have 20 cubes, the most you can keep. Delete one to add another.",
    );
    expect(c.querySelector("#cube-name")).toBeNull();
  });

  it("asks nothing of the server for someone who is not signed in", async () => {
    setSession(null);
    serve({});
    const { container: c } = render(MyCubes as never, {} as never);
    await settle();
    expect(sent("GET", "/me/cubes")).toHaveLength(0);
    const card = c.querySelector('[data-testid="cubes-signed-out"]')!;
    expect(text(card)).toContain("Sign in to keep a cube.");
    expect(text(card)).toContain("You can draft a set without one.");
    expect(c.querySelector("form")).toBeNull();
  });
});
