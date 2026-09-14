import { describe, expect, it } from "vitest";

import {
  BUG_DESC_MAX,
  BUG_IMAGE_TYPES,
  BUG_KINDS,
  BUG_MAX_IMAGES,
  BUG_MAX_IMAGE_BYTES,
  BUG_TITLE_MAX,
  DEFAULT_BUG_KIND,
  acceptableImages,
  buildBugContext,
  bugKindSpec,
  collectBugLog,
  describeBugAttachments,
  formatBytes,
  joinPhrases,
  validateAttachments,
  validateBugReport,
} from "./bugReport";
import type { GameView } from "./protocol";

function viewWithTurn(turn: Partial<GameView["turn"]>): GameView {
  return {
    turn: {
      number: 0,
      active_seat: 0,
      priority_holder: 0,
      phase: "",
      step: "",
      ...turn,
    },
  } as GameView;
}

describe("buildBugContext", () => {
  it("degrades to id + connection when the view is null", () => {
    const ctx = buildBugContext("g-123", null, 0, "reconnecting");
    expect(ctx).toEqual({ game_id: "g-123", connection: "reconnecting" });
  });

  it("captures turn/phase/step/seq from a live view", () => {
    const view = viewWithTurn({ number: 7, phase: "combat", step: "declare_blockers" });
    const ctx = buildBugContext("g-1", view, 412, "connected");
    expect(ctx).toEqual({
      game_id: "g-1",
      connection: "connected",
      seq: 412,
      turn: 7,
      phase: "combat",
      step: "declare_blockers",
    });
  });

  it("omits seq when the watermark is still zero", () => {
    const ctx = buildBugContext("g-1", null, 0, "connected");
    expect(ctx.seq).toBeUndefined();
  });
});

describe("validateBugReport", () => {
  it("requires a non-blank title", () => {
    expect(validateBugReport("   ", "")).toMatch(/title is required/);
    expect(validateBugReport("it broke", "")).toBeNull();
  });

  it("accepts titles exactly at the cap and rejects one past it", () => {
    expect(validateBugReport("t".repeat(BUG_TITLE_MAX), "")).toBeNull();
    expect(validateBugReport("t".repeat(BUG_TITLE_MAX + 1), "")).toMatch(/too long/);
  });

  it("bounds the description", () => {
    expect(validateBugReport("ok", "d".repeat(BUG_DESC_MAX))).toBeNull();
    expect(validateBugReport("ok", "d".repeat(BUG_DESC_MAX + 1))).toMatch(/too long/);
  });
});

// --- attachments + client log (added with the log/screenshot pass) ---

describe("validateAttachments", () => {
  const png = (name: string, size: number) => ({ name, type: "image/png", size });

  it("accepts a legal set", () => {
    expect(validateAttachments([png("a.png", 1024), png("b.png", 2048)])).toBeNull();
  });

  it("accepts every type the server renders", () => {
    for (const type of BUG_IMAGE_TYPES) {
      expect(validateAttachments([{ name: "x", type, size: 10 }])).toBeNull();
    }
  });

  // SVG is refused by the server (script carrier on an unauthenticated
  // route), so the client must not let one through and produce a
  // confusing rejection after the upload.
  it("rejects non-raster types including SVG", () => {
    for (const type of ["image/svg+xml", "application/pdf", "text/html", ""]) {
      expect(validateAttachments([{ name: "evil", type, size: 10 }])).toMatch(/isn't a PNG/);
    }
  });

  it("rejects a single oversized image", () => {
    expect(validateAttachments([png("huge.png", BUG_MAX_IMAGE_BYTES + 1)])).toMatch(/the limit is/);
  });

  // Each file legal, the set not — the aggregate cap is a property of
  // the set, which is why the whole list is validated at once.
  it("rejects a set that busts the aggregate cap", () => {
    const each = BUG_MAX_IMAGE_BYTES - 16;
    const problem = validateAttachments([
      png("a.png", each),
      png("b.png", each),
      png("c.png", each),
    ]);
    expect(problem).toMatch(/attachments total/);
  });

  it("rejects too many images", () => {
    const many = Array.from({ length: BUG_MAX_IMAGES + 1 }, (_, i) => png(`${i}.png`, 16));
    expect(validateAttachments(many)).toMatch(/at most/);
  });
});

describe("acceptableImages", () => {
  // A clipboard paste carries text/html and text/plain parts next to
  // the image; filtering rather than rejecting is what makes Ctrl+V
  // work at all.
  it("keeps images and drops everything else", () => {
    const got = acceptableImages([
      { name: "shot.png", type: "image/png", size: 1 },
      { name: "clip.html", type: "text/html", size: 1 },
      { name: "note.txt", type: "text/plain", size: 1 },
      { name: "photo.jpg", type: "image/jpeg", size: 1 },
    ]);
    expect(got.map((f) => f.name)).toEqual(["shot.png", "photo.jpg"]);
  });
});

describe("collectBugLog", () => {
  const at = (ms: number) => new Date(ms);

  // The interleaving is the diagnosis: action sent, then a TypeError,
  // then no snapshot. Two separate lists would lose that.
  it("merges protocol and console entries into one timeline", () => {
    const got = collectBugLog(
      [
        { at: at(1000), direction: "sent", text: "action cast_spell" },
        { at: at(3000), direction: "received", text: "snapshot seq=2" },
      ],
      [{ at: 2000, text: "console.error: TypeError" }],
    );
    expect(got.map((e) => e.text)).toEqual([
      "action cast_spell",
      "console.error: TypeError",
      "snapshot seq=2",
    ]);
    expect(got.map((e) => e.kind)).toEqual(["sent", "console", "received"]);
  });

  it("normalises an unrecognised direction to info", () => {
    const got = collectBugLog([{ at: at(1), direction: "weird", text: "x" }], []);
    expect(got[0].kind).toBe("info");
  });

  it("keeps the tail when over the limit", () => {
    const entries = Array.from({ length: 10 }, (_, i) => ({
      at: at(i),
      direction: "info",
      text: `e${i}`,
    }));
    const got = collectBugLog(entries, [], 3);
    expect(got.map((e) => e.text)).toEqual(["e7", "e8", "e9"]);
  });

  it("handles both sides being empty", () => {
    expect(collectBugLog([], [])).toEqual([]);
  });
});

describe("formatBytes", () => {
  it("scales to a unit a human reads", () => {
    expect(formatBytes(512)).toBe("512 B");
    expect(formatBytes(2048)).toBe("2 KB");
    expect(formatBytes(3 * 1024 * 1024)).toBe("3.0 MB");
  });
});

describe("report kinds", () => {
  it("opens on bug — the kind a client without a picker files", () => {
    expect(DEFAULT_BUG_KIND).toBe("bug");
    expect(bugKindSpec(DEFAULT_BUG_KIND).label).toBe("bug");
  });

  it("maps each kind to an EXISTING repo label", () => {
    // Not a taxonomy we invent: these three already exist on the
    // tracker, so a report never causes GitHub to create a label.
    const existing = ["bug", "enhancement", "documentation", "question"];
    expect(BUG_KINDS.map((k) => k.label)).toEqual(["bug", "enhancement", "question"]);
    for (const k of BUG_KINDS) {
      expect(existing).toContain(k.label);
    }
  });

  it("falls back to bug rather than throwing on an unknown kind", () => {
    expect(bugKindSpec("nonsense").kind).toBe("bug");
  });

  it("mirrors the server's per-kind artifact policy", () => {
    expect(bugKindSpec("bug")).toMatchObject({ log: true, replay: true, gameLog: true });
    // A question is answered from what happened, not from
    // frame-by-frame state — the expensive pin stays off.
    expect(bugKindSpec("question")).toMatchObject({ log: true, replay: false, gameLog: true });
    // An idea is about what the game SHOULD do: a screenshot is the
    // evidence, and a 30 MiB replay would be pure waste.
    expect(bugKindSpec("idea")).toMatchObject({ log: false, replay: false, gameLog: false });
  });

  it("describes what a bug attaches, in the reporter's words", () => {
    const got = describeBugAttachments(bugKindSpec("bug"), true);
    expect(got).toEqual([
      "your recent activity log",
      "an admin-only snapshot of this game's replay and log",
    ]);
  });

  it("names only the log when that is all a kind pins", () => {
    expect(describeBugAttachments(bugKindSpec("question"), true)).toEqual([
      "your recent activity log",
      "an admin-only snapshot of this game's log",
    ]);
  });

  it("joins phrases the way a person would read them", () => {
    expect(joinPhrases([])).toBe("");
    expect(joinPhrases(["one"])).toBe("one");
    expect(joinPhrases(["one", "two"])).toBe("one and two");
    expect(joinPhrases(["one", "two", "three"])).toBe("one, two and three");
  });

  it("says nothing extra rides along with an idea", () => {
    expect(describeBugAttachments(bugKindSpec("idea"), true)).toEqual([]);
  });

  it("drops the game-bound artifacts when there is no game", () => {
    expect(describeBugAttachments(bugKindSpec("question"), false)).toEqual([
      "your recent activity log",
    ]);
  });
});
