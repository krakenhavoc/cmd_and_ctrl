// ADR 0209 §4 (#4397, #4416): the signature rules. Which card has which scene
// (held to the server's catalog, so an oracle_id is never guessed), the
// caps and Short, what a signature takes from the generic plan, and the
// lazy loader. The rendered half is in signatures.render.test.ts; the
// scene kit's geometry and the scenes' own timing in scenes/kit.test.ts.

import { readFileSync, readdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { join } from "node:path";

import { describe, expect, it, vi } from "vitest";

import type { LogEvent } from "../protocol";
import { SOUND_CATEGORY, SOUND_FALLBACK, SOUND_MANIFEST, isLazySound } from "../sounds";
import {
  EFFECT_SHORT_CAP_MS,
  WIPE_CARD_MS,
  WIPE_SPREAD_MS,
  type EffectCue,
  type EffectPlan,
} from "../tableEffects";
import {
  SCENES,
  SCENE_LOAD_ATTEMPTS,
  SIGNATURES,
  SIGNATURE_CAP_MS,
  SIGNATURE_SHORT_CAP_MS,
  SignatureLoader,
  planSignatures,
  sceneForCard,
  signatureCueTiming,
  signatureFor,
  signatureTiming,
  withSignatures,
  type SceneID,
} from "./signatures";

const SERVER = fileURLToPath(new URL("../../../../server/internal/cards/", import.meta.url));

const ACT = "7a2484a9-04fd-41a0-8224-610c1c07ed10";
const BOLAS = "cb33e07b-9599-4979-991f-df0c43ddac31";

describe("the signature catalog", () => {
  const effectsDir = join(SERVER, "effects");
  const specs = readdirSync(effectsDir)
    .filter((f) => f.endsWith(".go") && !f.endsWith("_test.go"))
    .map((f) => readFileSync(join(effectsDir, f), "utf8"));

  it.each(Object.entries(SIGNATURES))(
    "%s is a catalog card under the name it is matched by",
    (id, def) => {
      // The catalog's own Spec: OracleID, then Name.
      const spec = specs.find((s) => s.includes(`OracleID:`) && s.includes(`"${id}"`));
      expect(spec, `no Spec in server/internal/cards/effects has OracleID ${id}`).toBeTruthy();
      const name = new RegExp(`OracleID:\\s*"${id}",\\s*\\n\\s*Name:\\s*"([^"]+)"`).exec(
        spec!,
      )?.[1];
      expect(name).toBeTruthy();
      // And the oracle fixture generated from Scryfall agrees.
      const fixture = JSON.parse(
        readFileSync(join(SERVER, "coverage/testdata/oracle", `${id}.json`), "utf8"),
      ) as { name: string };
      expect(def.names).toContain(name);
      expect(def.names).toContain(fixture.name);
    },
  );

  it("every printing named Nicol Bolas in the catalog shares one scene", () => {
    const bolas = Object.values(SIGNATURES).filter((d) =>
      d.names.some((n) => n.startsWith("Nicol Bolas")),
    );
    expect(bolas.length).toBeGreaterThan(0);
    expect(new Set(bolas.map((d) => d.scene))).toEqual(new Set(["nicol_bolas"]));
  });

  it("every scene is used, plays in the effects category, and loads lazily", () => {
    const used = new Set(Object.values(SIGNATURES).map((d) => d.scene));
    for (const [id, spec] of Object.entries(SCENES)) {
      expect(used.has(id as SceneID)).toBe(true);
      expect(SOUND_CATEGORY[spec.sound]).toBe("effects");
      expect(SOUND_MANIFEST[spec.sound].length).toBeGreaterThan(0);
      expect(isLazySound(spec.sound)).toBe(true);
      // Until Ian's files land, a signature sounds like the cue it replaces.
      expect(SOUND_FALLBACK[spec.sound]).toBeTruthy();
    }
  });

  it("every scene fits its caps: its first beat in Short, its cards inside the whole", () => {
    for (const spec of Object.values(SCENES)) {
      expect(spec.ms).toBeLessThanOrEqual(SIGNATURE_CAP_MS);
      expect(spec.firstBeatMs).toBeLessThanOrEqual(SIGNATURE_SHORT_CAP_MS);
      expect(spec.firstBeatMs).toBeLessThanOrEqual(spec.ms);
      expect(spec.cardsAtMs + WIPE_SPREAD_MS + WIPE_CARD_MS).toBeLessThanOrEqual(SIGNATURE_CAP_MS);
    }
  });

  it("looks a card up by oracle_id, and a card on the table by its name", () => {
    expect(signatureFor(ACT)).toBe("blasphemous_act");
    expect(signatureFor(BOLAS)).toBe("nicol_bolas");
    expect(signatureFor("00000000-0000-0000-0000-000000000000")).toBeNull();
    expect(sceneForCard({ name: "Blasphemous Act" })).toBe("blasphemous_act");
    expect(sceneForCard({ name: "toxic deluge" })).toBe("toxic_deluge");
    expect(sceneForCard({ name: "Nicol Bolas, God-Pharaoh" })).toBe("nicol_bolas");
    expect(sceneForCard({ name: "Grizzly Bears" })).toBeNull();
    // A face-down card has no name; a plain token has no signature; a
    // token copy of a real card is that card.
    expect(sceneForCard({ name: "" })).toBeNull();
    expect(sceneForCard({ name: "Nicol Bolas, God-Pharaoh", is_token: true })).toBeNull();
    expect(
      sceneForCard({ name: "Nicol Bolas, God-Pharaoh", is_token: true, token_copy: true }),
    ).toBe("nicol_bolas");
  });
});

describe("the caps", () => {
  const long = { ms: 4000, firstBeatMs: 1600 };

  it("Full plays the whole scene, cut to 3.5 s (the #4416 amendment)", () => {
    expect(SIGNATURE_CAP_MS).toBe(3500);
    expect(signatureTiming({ ms: 2000, firstBeatMs: 600 }, "full", 1)).toEqual({ ms: 2000, k: 1 });
    const t = signatureTiming(long, "full", 1)!;
    expect(t.ms).toBe(SIGNATURE_CAP_MS);
    expect(4000 * t.k).toBeCloseTo(SIGNATURE_CAP_MS);
  });

  it("Short plays the first beat only, at most 0.8 s", () => {
    expect(signatureTiming({ ms: 2000, firstBeatMs: 600 }, "short", 1)).toEqual({ ms: 600, k: 1 });
    const t = signatureTiming(long, "short", 1)!;
    expect(t.ms).toBe(EFFECT_SHORT_CAP_MS);
    expect(1600 * t.k).toBeCloseTo(EFFECT_SHORT_CAP_MS);
  });

  it("both scale with animations.speed, after the cap", () => {
    expect(signatureTiming(long, "full", 2)!.ms).toBe(SIGNATURE_CAP_MS * 2);
    expect(signatureTiming(long, "short", 0.5)!.ms).toBe(EFFECT_SHORT_CAP_MS * 0.5);
    // A nonsense speed is speed 1.
    expect(signatureTiming(long, "full", Number.NaN)!.ms).toBe(SIGNATURE_CAP_MS);
  });

  it("still (Off, reduced motion) plays no scene", () => {
    expect(signatureTiming(long, "still", 1)).toBeNull();
    expect(signatureCueTiming({ kind: "destroy", delayMs: 600 }, "still", 1)).toBeNull();
  });

  it("a card's own cue inside a scene ends within the cap", () => {
    for (const delayMs of [0, 650, 1400, 3000]) {
      const full = signatureCueTiming({ kind: "dies", delayMs }, "full", 1)!;
      expect(full.delayMs + full.ms).toBeLessThanOrEqual(SIGNATURE_CAP_MS + 0.001);
      const short = signatureCueTiming({ kind: "destroy", delayMs }, "short", 1)!;
      expect(short.delayMs + short.ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS + 0.001);
    }
    const slow = signatureCueTiming({ kind: "destroy", delayMs: 650 }, "full", 2)!;
    const fast = signatureCueTiming({ kind: "destroy", delayMs: 650 }, "full", 1)!;
    expect(slow.delayMs).toBeCloseTo(fast.delayMs * 2);
  });
});

// ---- planning ----

let seq = 0;
const entry = (kind: LogEvent["kind"], extra: Partial<LogEvent> = {}): LogEvent => ({
  seq: ++seq,
  kind,
  seat: 0,
  text: "",
  ...extra,
});
const dies = (id: string, source: string, batch = 7): LogEvent =>
  entry("zone", {
    card_id: id,
    old_zone: "battlefield",
    new_zone: "graveyard",
    cause: "sba",
    batch,
    source,
  });

const scenes: Record<string, SceneID> = {
  act: "blasphemous_act",
  deluge: "toxic_deluge",
  triumph: "bitter_triumph",
  bolas: "nicol_bolas",
};
const ctx = () => ({
  sceneOf: (id: string) => scenes[id] ?? null,
});

describe("planSignatures", () => {
  it("a spell's resolve names its batch: every permanent with the spell as its source", () => {
    const log = [
      entry("resolve", { card_id: "act", batch: 7 }),
      dies("a", "act"),
      dies("b", "act"),
      dies("x", "other"),
      entry("zone", { card_id: "act", old_zone: "stack", new_zone: "graveyard" }),
    ];
    const [sig, ...rest] = planSignatures(log, ctx());
    expect(rest).toHaveLength(0);
    expect(sig).toMatchObject({
      scene: "blasphemous_act",
      cardID: "act",
      took: ["a", "b"],
      lands: null,
    });
  });

  it("an ability's resolve is not the spell's", () => {
    expect(
      planSignatures([entry("resolve", { card_id: "act", label: "Something" })], ctx()),
    ).toEqual([]);
  });

  it("a scene that needs a card plays nothing when its batch took none", () => {
    expect(planSignatures([entry("resolve", { card_id: "triumph" })], ctx())).toEqual([]);
    const [sig] = planSignatures(
      [entry("resolve", { card_id: "triumph" }), dies("a", "triumph")],
      ctx(),
    );
    expect(sig.took).toEqual(["a"]);
    // A wipe plays even over an empty board.
    expect(planSignatures([entry("resolve", { card_id: "act" })], ctx())).toHaveLength(1);
  });

  it("casting an enter scene's card plays nothing (#4416: only the entering version)", () => {
    expect(planSignatures([entry("cast", { card_id: "bolas" })], ctx())).toEqual([]);
  });

  it("an enter scene plays when the card enters the battlefield, cast or not", () => {
    for (const from of ["stack", "graveyard", "exile", "hand"]) {
      const enter = entry("zone", { card_id: "bolas", old_zone: from, new_zone: "battlefield" });
      const [up, ...rest] = planSignatures([enter], ctx());
      expect(rest).toEqual([]);
      expect(up).toMatchObject({
        scene: "nicol_bolas",
        lands: "battlefield",
        replacesCue: String(enter.seq),
      });
    }
    // Cast and resolved in one frame: once, on the battlefield.
    const both = planSignatures(
      [
        entry("cast", { card_id: "bolas" }),
        entry("zone", { card_id: "bolas", old_zone: "stack", new_zone: "battlefield" }),
      ],
      ctx(),
    );
    expect(both.map((s) => s.lands)).toEqual(["battlefield"]);
    // Moving about the battlefield is not entering it.
    expect(
      planSignatures(
        [entry("zone", { card_id: "bolas", old_zone: "battlefield", new_zone: "battlefield" })],
        ctx(),
      ),
    ).toEqual([]);
  });

  it("a card with no signature, or an entry with no card, is no signature", () => {
    expect(
      planSignatures([entry("resolve", { card_id: "bears" }), entry("resolve", {})], ctx()),
    ).toEqual([]);
  });
});

const cue = (id: string, cardID: string, kind: EffectCue["kind"] = "dies"): EffectCue => ({
  id,
  seq: Number(id),
  kind,
  cardID,
  stackItemID: null,
  batch: 7,
  weight: "light",
  delayMs: 250,
});

describe("withSignatures", () => {
  const plan: EffectPlan = {
    cues: [cue("2", "a"), cue("3", "b"), cue("9", "z", "destroy")],
    wipes: [{ batch: 7, source: "act", count: 3, tone: "dark", seq: 2 }],
  };
  const sig = planSignatures(
    [entry("resolve", { card_id: "act", batch: 7 }), dies("a", "act"), dies("b", "act")],
    ctx(),
  );

  it("a scene that has not loaded changes nothing: the generic cues play", () => {
    const sp = withSignatures(plan, sig, () => false, "full");
    expect(sp.sigs).toEqual([]);
    expect(sp.plan).toBe(plan);
  });

  it("with motion, it takes the sweep and puts its cards' cues on its clock", () => {
    const sp = withSignatures(plan, sig, () => true, "full");
    expect(sp.sigs).toHaveLength(1);
    expect(sp.plan.wipes).toEqual([]);
    expect(sp.silencedWipes).toEqual(new Set([7]));
    const [a, b, z] = sp.plan.cues;
    expect(a).toMatchObject({
      id: "2",
      weight: "light",
      delayMs: SCENES.blasphemous_act.cardsAtMs,
    });
    expect(b.delayMs).toBeGreaterThan(a.delayMs);
    // A card the spell did not take keeps its own cue as it was.
    expect(z).toBe(plan.cues[2]);
    expect(sp.inScene).toEqual(new Set(["2", "3"]));
  });

  it("still keeps the generic still cues, but the signature's sound replaces the wipe's", () => {
    const sp = withSignatures(plan, sig, () => true, "still");
    expect(sp.plan).toBe(plan);
    expect(sp.inScene.size).toBe(0);
    expect(sp.silencedWipes).toEqual(new Set([7]));
  });

  it("a scene that draws its cards' exits (Toxic Deluge) leaves their cues out, with motion", () => {
    expect(SCENES.toxic_deluge.ownsCards).toBe(true);
    const deluge = planSignatures(
      [entry("resolve", { card_id: "deluge", batch: 7 }), dies("a", "deluge"), dies("b", "deluge")],
      ctx(),
    );
    const plan2: EffectPlan = { ...plan, wipes: [{ ...plan.wipes[0], source: "deluge" }] };
    const sp = withSignatures(plan2, deluge, () => true, "full");
    expect(sp.plan.cues.map((c) => c.id)).toEqual(["9"]);
    expect(sp.plan.wipes).toEqual([]);
    expect(sp.inScene.size).toBe(0);
    // Still keeps every still cue: the badge says what happened.
    const still = withSignatures(plan2, deluge, () => true, "still");
    expect(still.plan.cues.map((c) => c.id)).toEqual(["2", "3", "9"]);
  });

  it("only Toxic Deluge leaves a tint on what it did not kill", () => {
    const tinted = Object.entries(SCENES)
      .filter(([, spec]) => spec.tint)
      .map(([id, spec]) => [id, spec.tint]);
    expect(tinted).toEqual([["toxic_deluge", "deluge"]]);
  });

  it("a one-card batch is one sound: the signature's", () => {
    const one = planSignatures(
      [entry("resolve", { card_id: "triumph" }), dies("z", "triumph")],
      ctx(),
    );
    const sp = withSignatures(plan, one, () => true, "full");
    expect(sp.silencedCues).toEqual(new Set(["9"]));
    expect(sp.plan.cues.find((c) => c.id === "9")).toMatchObject({ weight: "light" });
  });

  it("an arrival replaces the planeswalker's ring, and in still only its sound", () => {
    const enter = entry("zone", {
      card_id: "bolas",
      old_zone: "graveyard",
      new_zone: "battlefield",
    });
    const pw: EffectPlan = {
      cues: [{ ...cue(String(enter.seq), "bolas", "pw_enter"), weight: "full" }],
      wipes: [],
    };
    const found = planSignatures([enter], ctx());
    const moving = withSignatures(pw, found, () => true, "short");
    expect(moving.plan.cues).toEqual([]);
    expect(moving.silencedCues).toEqual(new Set([String(enter.seq)]));
    const still = withSignatures(pw, found, () => true, "still");
    expect(still.plan.cues).toHaveLength(1);
    expect(still.silencedCues).toEqual(new Set([String(enter.seq)]));
  });
});

// ---- lazy loading ----

describe("SignatureLoader", () => {
  const mod = { play: vi.fn() };

  it("loads a scene the first time its card is seen, once, and preloads its sound", async () => {
    const load = vi.fn(() => Promise.resolve(mod));
    const preloadSound = vi.fn();
    const l = new SignatureLoader({ load, preloadSound });
    l.notice([{ name: "Grizzly Bears" }, { name: "" }]);
    expect(load).not.toHaveBeenCalled();
    expect(l.status("blasphemous_act")).toBe("idle");

    l.notice([{ name: "Blasphemous Act" }]);
    expect(load).toHaveBeenCalledWith("blasphemous_act");
    expect(preloadSound).toHaveBeenCalledWith("sig_blasphemous_act");
    // Not ready until the import resolves: nothing waits on it.
    expect(l.status("blasphemous_act")).toBe("loading");
    expect(l.ready("blasphemous_act")).toBeNull();
    l.notice([{ name: "Blasphemous Act" }, { name: "Blasphemous Act" }]);
    await Promise.resolve();
    expect(l.ready("blasphemous_act")).toBe(mod);
    l.notice([{ name: "Blasphemous Act" }]);
    expect(load).toHaveBeenCalledTimes(1);
    expect(l.ready("toxic_deluge")).toBeNull();
  });

  it("a failed load is tried again on a later sighting, a few times at most", async () => {
    const load = vi.fn(() => Promise.reject(new Error("offline")));
    const l = new SignatureLoader({ load });
    for (let i = 0; i < SCENE_LOAD_ATTEMPTS + 3; i++) {
      l.notice([{ name: "Toxic Deluge" }]);
      await Promise.resolve();
      await Promise.resolve();
    }
    expect(load).toHaveBeenCalledTimes(SCENE_LOAD_ATTEMPTS);
    expect(l.status("toxic_deluge")).toBe("failed");
    expect(l.ready("toxic_deluge")).toBeNull();
  });

  it("each scene's import is a module with a play function", async () => {
    for (const spec of Object.values(SCENES)) {
      const m = await spec.load();
      expect(typeof m.play).toBe("function");
    }
  });
});
