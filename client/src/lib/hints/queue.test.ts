// queue.test.ts — which hint shows, and when (ADR 0125 §3.5, §8): one at
// a time; one per site-page visit; the order; `when`; tipsOff; a missing
// anchor skips and logs; replays ignore tipsOff and the one-per-visit
// rule; at the table only in a quiet moment, stepping aside when it ends.

import { describe, expect, it } from "vitest";

import { emptyContext, type Hint, type HintContext, type HintPlace } from "./hint";
import {
  SITE_WINDOW_MS,
  candidates,
  createHintEngine,
  missingAnchorMessage,
  pickSiteHint,
  type EngineEnv,
} from "./queue";
import type { SeenMap } from "./seen";
import { TABLE_GAP_MS } from "./tableMoment";
import { L } from "../labels";
import { moment } from "../test/hintContexts";
import { BOT, ME } from "../test/tutorialBoards";

function hint(id: `${HintPlace}.${string}`, order = 0, extra: Partial<Hint> = {}): Hint {
  return {
    id,
    version: 1,
    place: id.split(".")[0] as HintPlace,
    order,
    anchor: { label: L.actions },
    title: "A title",
    body: "A body.",
    ...extra,
  };
}

const LOBBY_A = hint("lobby.a", 1);
const LOBBY_B = hint("lobby.b", 2);
const SITE_HELP = hint("site.help", 0);
const SIGNED_IN_ONLY = hint("lobby.signed", 0, { when: (c) => c.signedIn });
const TABLE_A = hint("table.a", 1);
const TABLE_B = hint("table.b", 2);

describe("candidates", () => {
  it("offers a page's own hints in order, then the site's", () => {
    const got = candidates({
      hints: [SITE_HELP, LOBBY_B, LOBBY_A],
      ctx: emptyContext("lobby"),
      seen: {},
      tipsOff: false,
    });
    expect(got.map((h) => h.id)).toEqual(["lobby.a", "lobby.b", "site.help"]);
  });

  it("skips seen hints, and offers one again when its version is bumped", () => {
    const seen: SeenMap = { "lobby.a": 1 };
    const ids = (hs: Hint[]) =>
      candidates({ hints: hs, ctx: emptyContext("lobby"), seen, tipsOff: false }).map((h) => h.id);
    expect(ids([LOBBY_A, LOBBY_B])).toEqual(["lobby.b"]);
    expect(ids([{ ...LOBBY_A, version: 2 }, LOBBY_B])).toEqual(["lobby.a", "lobby.b"]);
  });

  it("honours when", () => {
    const guest = emptyContext("lobby");
    const signed: HintContext = { ...guest, signedIn: true };
    const base = { hints: [SIGNED_IN_ONLY, LOBBY_A], seen: {}, tipsOff: false };
    expect(candidates({ ...base, ctx: guest }).map((h) => h.id)).toEqual(["lobby.a"]);
    expect(candidates({ ...base, ctx: signed }).map((h) => h.id)).toEqual([
      "lobby.signed",
      "lobby.a",
    ]);
  });

  it("offers nothing with tips off, except a replay's own hints", () => {
    const base = { hints: [LOBBY_A, LOBBY_B], ctx: emptyContext("lobby"), seen: {} };
    expect(candidates({ ...base, tipsOff: true })).toEqual([]);
    const replay = new Set(["lobby.b"]);
    expect(candidates({ ...base, tipsOff: true, replay }).map((h) => h.id)).toEqual(["lobby.b"]);
  });

  it("keeps the table and the settings dialog to their own hints", () => {
    const all = [SITE_HELP, LOBBY_A, TABLE_A, hint("settings.display")];
    const at = (p: HintPlace) =>
      candidates({ hints: all, ctx: emptyContext(p), seen: {}, tipsOff: false }).map((h) => h.id);
    expect(at("table")).toEqual(["table.a"]);
    expect(at("settings")).toEqual(["settings.display"]);
  });
});

describe("pickSiteHint", () => {
  const on =
    (...ids: string[]) =>
    (h: Hint) =>
      ids.includes(h.id);

  it("shows the first hint as soon as its anchor resolves", () => {
    expect(pickSiteHint([LOBBY_A, LOBBY_B], on("lobby.a", "lobby.b"), 0).hint).toBe(LOBBY_A);
  });

  it("waits for the first hint's anchor through the window before passing it over", () => {
    const early = pickSiteHint([LOBBY_A, LOBBY_B], on("lobby.b"), SITE_WINDOW_MS - 1);
    expect(early).toEqual({ hint: null, missing: [], settled: false });
    const late = pickSiteHint([LOBBY_A, LOBBY_B], on("lobby.b"), SITE_WINDOW_MS);
    expect(late).toEqual({ hint: LOBBY_B, missing: [LOBBY_A], settled: true });
  });

  it("settles with nothing when no anchor comes", () => {
    expect(pickSiteHint([LOBBY_A], on(), SITE_WINDOW_MS)).toEqual({
      hint: null,
      missing: [LOBBY_A],
      settled: true,
    });
    expect(pickSiteHint([], on(), 0).settled).toBe(true);
  });
});

/** A small driver: an engine, a clock, a seen map and the anchors on the page. */
function driver(hints: Hint[], place: HintPlace = "lobby") {
  const logs: string[] = [];
  const engine = createHintEngine({ log: (m) => logs.push(m) });
  const st = {
    now: 1_000_000,
    seen: {} as SeenMap,
    tipsOff: false,
    visit: "v1",
    ctx: emptyContext(place) as HintContext,
    anchors: new Set(hints.map((h) => h.id)),
  };
  const env = (): EngineEnv => ({
    hints,
    ctx: st.ctx,
    visit: st.visit,
    seen: st.seen,
    tipsOff: st.tipsOff,
    now: st.now,
    resolves: (h) => st.anchors.has(h.id),
  });
  return {
    st,
    logs,
    engine,
    tick: () => engine.tick(env())?.id ?? null,
    dismiss: () => {
      const h = engine.dismiss(st.now);
      if (h) st.seen = { ...st.seen, [h.id]: h.version };
      return h?.id ?? null;
    },
  };
}

describe("the engine on a site page", () => {
  it("shows one hint at a time, and only one per visit", () => {
    const d = driver([LOBBY_A, LOBBY_B]);
    expect(d.tick()).toBe("lobby.a");
    expect(d.tick()).toBe("lobby.a");
    expect(d.dismiss()).toBe("lobby.a");
    d.st.now += 60_000;
    expect(d.tick()).toBeNull();
    // The next visit offers the next one.
    d.st.visit = "v2";
    expect(d.tick()).toBe("lobby.b");
  });

  it("skips a hint whose anchor never comes, logs it once, and keeps it unseen", () => {
    const d = driver([LOBBY_A, LOBBY_B]);
    d.st.anchors.delete("lobby.a");
    expect(d.tick()).toBeNull();
    d.st.now += SITE_WINDOW_MS;
    expect(d.tick()).toBe("lobby.b");
    d.tick();
    expect(d.logs).toEqual([missingAnchorMessage("lobby.a")]);
    expect(d.logs[0]).toBe("hint: lobby.a has no anchor on the page");
    d.dismiss();
    expect(d.st.seen).toEqual({ "lobby.b": 1 });
  });

  it("never makes a page wait: with no anchor at all the visit simply ends", () => {
    const d = driver([LOBBY_A]);
    d.st.anchors.clear();
    expect(d.tick()).toBeNull();
    d.st.now += SITE_WINDOW_MS;
    expect(d.tick()).toBeNull();
    d.st.anchors.add("lobby.a");
    expect(d.tick()).toBeNull();
  });

  it("offers nothing with tips off", () => {
    const d = driver([LOBBY_A]);
    d.st.tipsOff = true;
    expect(d.tick()).toBeNull();
  });

  it("hides the card while its anchor is gone, and keeps it as the visit's one hint", () => {
    const d = driver([LOBBY_A, LOBBY_B]);
    expect(d.tick()).toBe("lobby.a");
    d.st.anchors.delete("lobby.a");
    expect(d.tick()).toBeNull();
    d.st.now += 60_000;
    expect(d.tick()).toBeNull();
    d.st.anchors.add("lobby.a");
    expect(d.tick()).toBe("lobby.a");
  });

  it("drops a hint that was marked seen elsewhere (another tab)", () => {
    const d = driver([LOBBY_A]);
    expect(d.tick()).toBe("lobby.a");
    d.st.seen = { "lobby.a": 1 };
    expect(d.tick()).toBeNull();
  });

  it("opening Settings and closing it again does not give the page a second hint", () => {
    const settingsHint = hint("settings.display");
    const d = driver([LOBBY_A, LOBBY_B, settingsHint]);
    expect(d.tick()).toBe("lobby.a");
    // Settings opens over the page: its own visit, its own hint.
    d.st.visit = "settings#1";
    d.st.ctx = emptyContext("settings");
    expect(d.tick()).toBe("settings.display");
    d.dismiss();
    // Back on the page: the page's hint is still its one.
    d.st.visit = "v1";
    d.st.ctx = emptyContext("lobby");
    expect(d.tick()).toBe("lobby.a");
    d.dismiss();
    d.st.now += 60_000;
    expect(d.tick()).toBeNull();
  });

  it("a replay shows the place's hints one after another, with tips off, past the visit's one", () => {
    const d = driver([LOBBY_A, LOBBY_B, SITE_HELP]);
    expect(d.tick()).toBe("lobby.a");
    d.dismiss();
    d.st.tipsOff = true;
    expect(d.tick()).toBeNull();
    // Help → Tips for this page: the caller forgets them first.
    d.st.seen = {};
    d.engine.replay("lobby", ["lobby.a", "lobby.b"]);
    expect(d.tick()).toBe("lobby.a");
    d.dismiss();
    expect(d.tick()).toBe("lobby.b");
    d.dismiss();
    // The replay is over; with tips off, site.help is not offered.
    expect(d.tick()).toBeNull();
  });

  it("a replay skips a hint whose anchor never comes", () => {
    const d = driver([LOBBY_A, LOBBY_B]);
    d.st.anchors.delete("lobby.a");
    d.engine.replay("lobby", ["lobby.a", "lobby.b"]);
    expect(d.tick()).toBeNull();
    d.st.now += SITE_WINDOW_MS;
    expect(d.tick()).toBe("lobby.b");
    expect(d.logs).toEqual([missingAnchorMessage("lobby.a")]);
  });
});

describe("the engine at the table", () => {
  function tableDriver(hints: Hint[] = [TABLE_A, TABLE_B]) {
    const d = driver(hints, "table");
    d.st.ctx = { ...emptyContext("table"), moment: moment() };
    const set = (over: Parameters<typeof moment>[0]) =>
      (d.st.ctx = { ...d.st.ctx, moment: moment(over) });
    return { ...d, set };
  }

  it("shows a hint in a quiet moment", () => {
    const d = tableDriver();
    expect(d.tick()).toBe("table.a");
  });

  it("shows nothing with no table moment (a spectator, or no table)", () => {
    const d = tableDriver();
    d.st.ctx = { ...d.st.ctx, moment: null };
    expect(d.tick()).toBeNull();
  });

  it("steps aside at once when a dock request opens, stays unseen, and comes back", () => {
    const d = tableDriver();
    expect(d.tick()).toBe("table.a");
    d.set({ dockRequest: true });
    expect(d.tick()).toBeNull();
    expect(d.st.seen).toEqual({});
    d.set({});
    expect(d.tick()).toBe("table.a");
  });

  it("never shows over an opponent's item while the viewer holds priority", () => {
    const d = tableDriver();
    d.set({ viewerHasPriority: true, stackTopController: BOT });
    expect(d.tick()).toBeNull();
    // The viewer's own item on top is quiet.
    d.set({ viewerHasPriority: true, stackTopController: ME });
    expect(d.tick()).toBe("table.a");
  });

  it("is off while the coach is visible", () => {
    const d = tableDriver();
    d.set({ coachVisible: true });
    expect(d.tick()).toBeNull();
  });

  it("waits twenty seconds after a hint closes before the next", () => {
    const d = tableDriver();
    expect(d.tick()).toBe("table.a");
    d.dismiss();
    d.st.now += TABLE_GAP_MS - 1;
    expect(d.tick()).toBeNull();
    d.st.now += 1;
    expect(d.tick()).toBe("table.b");
  });

  it("offers several over a game, one at a time, in order", () => {
    const d = tableDriver();
    expect(d.tick()).toBe("table.a");
    expect(d.tick()).toBe("table.a");
    d.dismiss();
    d.st.now += TABLE_GAP_MS;
    expect(d.tick()).toBe("table.b");
  });

  it("passes over a hint whose anchor is not on screen, and logs it once", () => {
    const d = tableDriver();
    d.st.anchors.delete("table.a");
    expect(d.tick()).toBe("table.b");
    d.tick();
    expect(d.logs).toEqual([missingAnchorMessage("table.a")]);
  });

  it("passes over a hint that waits for its anchor without logging it, and shows it once it is there", () => {
    const waiting: Hint = { ...TABLE_A, waitsForAnchor: true };
    const d = tableDriver([waiting, TABLE_B]);
    d.st.anchors.delete("table.a");
    expect(d.tick()).toBe("table.b");
    d.tick();
    expect(d.logs).toEqual([]);
    d.dismiss();
    d.st.now += TABLE_GAP_MS;
    d.st.anchors.add("table.a");
    expect(d.tick()).toBe("table.a");
  });

  it("a table replay shows its hints at the next quiet moments, with tips off", () => {
    const d = tableDriver();
    d.st.tipsOff = true;
    expect(d.tick()).toBeNull();
    d.engine.replay("table", ["table.a", "table.b"]);
    d.set({ dockRequest: true });
    expect(d.tick()).toBeNull();
    d.set({});
    expect(d.tick()).toBe("table.a");
  });
});
