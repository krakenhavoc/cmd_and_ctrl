// scenarios.ts — one scenario per table-effect cue (#4868, ADR 0209).
//
// A scenario is the table just before a moment (`before`), the table the
// frame shows after it (`after`), and the log entries the server sends for
// it (`log`). The harness draws `before`, lets the layers measure it, then
// lands `after` with `log` as one new frame, exactly as Game.svelte does.
//
// To add one: copy the scenario nearest to your cue, change the cards and
// the entries, and list the cue it shows in `shows`. scenarios.test.ts
// fails while any EffectKind has no scenario, or a scenario's log does not
// plan the cue it says it shows. See README.md.

import type { EffectKind } from "../../src/lib/tableEffects";
import type { Entry, FxCard, FxOut } from "./table";

/**
 * What a scenario shows: a TableEffects cue kind, a board wipe, or a cue
 * of one of the other layers the board mounts beside it (EliminationFx,
 * ExtortScales, DiceLayer).
 */
export type CueName =
  | EffectKind
  | "board_wipe"
  | "decked_out"
  | "eliminated"
  | "extort"
  | "dice_roll"
  | "coin_flip";

export interface Scenario {
  /** Its URL hash: `#<id>` plays it on load, `#<id>@<ms>` also pauses it at <ms>. */
  id: string;
  /** Its button's text, in sentence case. */
  title: string;
  /** The heading it is listed under. */
  group: string;
  /** The issue that built the cue. */
  issue: number;
  shows: CueName[];
  before: FxCard[];
  after: FxCard[];
  /** Seats that have left the game in `after`. */
  out?: FxOut[];
  log: Entry[];
}

// ---- Cards used by several scenarios ----

const ogre: FxCard = { id: "ogre", name: "Hoarding Ogre", pt: "3/3" };
const elf: FxCard = { id: "elf", name: "Llanowar Elves", pt: "1/1" };
const angel: FxCard = { id: "angel", name: "Serra Angel", pt: "4/4" };
const jace: FxCard = {
  id: "jace",
  name: "Jace, the Mind Sculptor",
  type: "Legendary Planeswalker — Jace",
};
const bolt: FxCard = {
  id: "bolt",
  name: "Lightning Bolt",
  type: "Instant",
  zone: "stack",
  item: "item-1",
};
const cmdr: FxCard = {
  id: "cmdr",
  name: "Meren of Clan Nel Toth",
  type: "Legendary Creature — Human Shaman",
  commander: true,
};
const werewolf = (id: string, face: number): FxCard => ({
  id,
  name: id === "w1" ? "Tovolar's Huntmaster" : "Kessig Naturalist",
  type: "Creature — Human Werewolf",
  face,
});

/** The same card, moved to another zone (and seat). */
const to = (c: FxCard, zone: FxCard["zone"], extra: Partial<FxCard> = {}): FxCard => ({
  ...c,
  zone,
  ...extra,
});

/** One card leaving the battlefield, as a `zone` entry. */
const leaves = (id: string, newZone: string, extra: Partial<Entry> = {}): Entry => ({
  kind: "zone",
  seat: -1,
  card_id: id,
  old_zone: "battlefield",
  new_zone: newZone,
  batch: 1,
  ...extra,
});

const wiped: FxCard[] = [
  { id: "w-a", name: "Hoarding Ogre", seat: 1 },
  { id: "w-b", name: "Llanowar Elves", seat: 1 },
  { id: "w-c", name: "Serra Angel", seat: 2 },
  { id: "w-d", name: "Grizzly Bears", seat: 2 },
  { id: "w-e", name: "Sakura-Tribe Elder", seat: 3 },
];

const milled = (n: number): FxCard[] =>
  Array.from({ length: n }, (_, i) => ({
    id: `m${i + 1}`,
    name: ["Opt", "Island", "Counterspell", "Brainstorm", "Ponder", "Preordain", "Forest"][i % 7],
    type: "Instant",
    zone: "graveyard",
  }));

export const SCENARIOS: Scenario[] = [
  // ---- A permanent leaves the battlefield (ADR 0209 §3, #4396, #4416) ----
  {
    id: "destroy",
    title: "Destroy",
    group: "Leaving the battlefield",
    issue: 4396,
    shows: ["destroy"],
    before: [ogre, elf],
    after: [to(ogre, "graveyard"), elf],
    log: [leaves("ogre", "graveyard", { cause: "destroy" })],
  },
  {
    id: "exile",
    title: "Exile",
    group: "Leaving the battlefield",
    issue: 4396,
    shows: ["exile"],
    before: [ogre, elf],
    after: [to(ogre, "exile"), elf],
    log: [leaves("ogre", "exile", { cause: "exile_effect" })],
  },
  {
    id: "exile_from_destroy",
    title: "Destroy replaced by exile",
    group: "Leaving the battlefield",
    issue: 4416,
    shows: ["exile"],
    before: [ogre, elf],
    after: [to(ogre, "exile"), elf],
    log: [leaves("ogre", "exile", { cause: "destroy" })],
  },
  {
    id: "dies",
    title: "Dies (state-based action)",
    group: "Leaving the battlefield",
    issue: 4396,
    shows: ["dies"],
    before: [ogre, elf],
    after: [to(ogre, "graveyard"), elf],
    log: [leaves("ogre", "graveyard", { cause: "sba" })],
  },
  {
    id: "sacrifice",
    title: "Sacrifice",
    group: "Leaving the battlefield",
    issue: 4416,
    shows: ["sacrifice"],
    before: [ogre, elf],
    after: [to(ogre, "graveyard"), elf],
    log: [{ kind: "sacrifice", seat: 1, card_id: "ogre", batch: 1 }],
  },
  {
    id: "crack",
    title: "Fetch land and Treasure cracked",
    group: "Leaving the battlefield",
    issue: 4427,
    shows: ["crack"],
    before: [
      { id: "wilds", name: "Evolving Wilds", type: "Land" },
      { id: "treasure", name: "Treasure", type: "Token Artifact — Treasure", token: true },
      elf,
    ],
    after: [{ id: "wilds", name: "Evolving Wilds", type: "Land", zone: "graveyard" }, elf],
    log: [
      { kind: "sacrifice", seat: 1, card_id: "wilds", cause: "cost", source: "wilds", batch: 1 },
      {
        kind: "sacrifice",
        seat: 1,
        card_id: "treasure",
        cause: "cost",
        source: "treasure",
        batch: 2,
      },
    ],
  },
  {
    id: "bounce",
    title: "Bounce",
    group: "Leaving the battlefield",
    issue: 4396,
    shows: ["bounce"],
    before: [ogre, elf],
    after: [to(ogre, "hand"), elf],
    log: [leaves("ogre", "hand", { cause: "bounce" })],
  },
  {
    id: "pw_enter",
    title: "Planeswalker enters",
    group: "Leaving the battlefield",
    issue: 4396,
    shows: ["pw_enter"],
    before: [to(jace, "stack"), elf],
    after: [jace, elf],
    log: [{ kind: "zone", card_id: "jace", old_zone: "stack", new_zone: "battlefield", batch: 1 }],
  },
  {
    id: "pw_leave",
    title: "Planeswalker destroyed",
    group: "Leaving the battlefield",
    issue: 4416,
    shows: ["pw_leave"],
    before: [jace, elf],
    after: [to(jace, "graveyard"), elf],
    log: [leaves("jace", "graveyard", { cause: "destroy" })],
  },
  {
    id: "wipe_destroy",
    title: "Board wipe (Damnation)",
    group: "Leaving the battlefield",
    issue: 4416,
    shows: ["board_wipe", "destroy"],
    before: wiped,
    after: wiped.map((c) => to(c, "graveyard")),
    log: wiped.map((c) => leaves(c.id, "graveyard", { cause: "destroy", source: "damnation" })),
  },
  {
    id: "wipe_exile",
    title: "Exile wipe (Farewell)",
    group: "Leaving the battlefield",
    issue: 4416,
    shows: ["board_wipe", "exile"],
    before: wiped,
    after: wiped.map((c) => to(c, "exile")),
    log: wiped.map((c) => leaves(c.id, "exile", { cause: "exile_effect", source: "farewell" })),
  },

  // ---- The stack (ADR 0209 §3, #4450) ----
  {
    id: "counter",
    title: "Counterspell",
    group: "The stack",
    issue: 4416,
    shows: ["counter"],
    before: [bolt, elf],
    after: [to(bolt, "graveyard"), elf],
    log: [{ kind: "counter", seat: 2, target: "bolt", stack_item_id: "item-1", batch: 1 }],
  },
  {
    id: "fizzle",
    title: "Fizzle",
    group: "The stack",
    issue: 4450,
    shows: ["fizzle"],
    before: [bolt, elf],
    after: [to(bolt, "graveyard"), elf],
    log: [
      {
        kind: "fizzle",
        seat: 1,
        card_id: "bolt",
        stack_item_id: "item-1",
        batch: 1,
        text: "Lightning Bolt fizzled: no legal targets",
      },
    ],
  },

  // ---- S80: a cue for every common moment (#4399) ----
  {
    id: "token_create",
    title: "Tokens created",
    group: "Common moments (S80)",
    issue: 4441,
    shows: ["token_create"],
    before: [ogre],
    after: [
      ogre,
      ...["t1", "t2", "t3"].map((id) => ({
        id,
        name: "Goblin",
        type: "Token Creature — Goblin",
        token: true,
        pt: "1/1",
      })),
    ],
    log: ["t1", "t2", "t3"].map((id) => ({ kind: "token", seat: 1, card_id: id, batch: 1 })),
  },
  {
    id: "counters",
    title: "Counters put on",
    group: "Common moments (S80)",
    issue: 4442,
    shows: ["counters"],
    before: [
      ogre,
      { id: "chalice", name: "Everflowing Chalice", type: "Artifact", counter: "charge" },
      angel,
    ],
    after: [
      { ...ogre, pt: "5/5" },
      { id: "chalice", name: "Everflowing Chalice", type: "Artifact", counter: "charge" },
      { ...angel, pt: "3/3" },
    ],
    log: [
      { kind: "counters", card_id: "ogre", label: "+1/+1", amount: 2, delta: 2, batch: 1 },
      { kind: "counters", card_id: "angel", label: "-1/-1", amount: 1, delta: 1, batch: 2 },
      { kind: "counters", card_id: "chalice", label: "charge", amount: 1, delta: 1, batch: 3 },
    ],
  },
  {
    id: "commander_cast",
    title: "Commander cast",
    group: "Common moments (S80)",
    issue: 4443,
    shows: ["commander_cast"],
    before: [to(cmdr, "command"), elf],
    after: [to(cmdr, "stack"), elf],
    log: [{ kind: "cast", seat: 1, card_id: "cmdr", old_zone: "command", batch: 1 }],
  },
  {
    id: "commander_return",
    title: "Commander goes home from the graveyard",
    group: "Common moments (S80)",
    issue: 4443,
    shows: ["commander_return"],
    before: [to(cmdr, "graveyard"), elf],
    after: [to(cmdr, "command"), elf],
    log: [
      {
        kind: "zone",
        seat: 1,
        card_id: "cmdr",
        old_zone: "graveyard",
        new_zone: "command",
        batch: 1,
      },
    ],
  },
  {
    id: "commander_return_field",
    title: "Commander bounced home",
    group: "Common moments (S80)",
    issue: 4443,
    shows: ["commander_return"],
    before: [cmdr, elf],
    after: [to(cmdr, "command"), elf],
    log: [leaves("cmdr", "command", { cause: "bounce" })],
  },
  {
    id: "attach_equip",
    title: "Equipment attached",
    group: "Common moments (S80)",
    issue: 4444,
    shows: ["attach_equip"],
    before: [ogre, { id: "sword", name: "Sword of Fire and Ice", type: "Artifact — Equipment" }],
    after: [
      ogre,
      {
        id: "sword",
        name: "Sword of Fire and Ice",
        type: "Artifact — Equipment",
        attachedTo: "ogre",
      },
    ],
    log: [
      { kind: "attach", seat: 1, card_id: "sword", target: "ogre", label: "Equipment", batch: 1 },
    ],
  },
  {
    id: "attach_aura",
    title: "Aura enchants",
    group: "Common moments (S80)",
    issue: 4444,
    shows: ["attach_aura"],
    before: [ogre, { id: "aura", name: "Rancor", type: "Enchantment — Aura", zone: "stack" }],
    after: [ogre, { id: "aura", name: "Rancor", type: "Enchantment — Aura", attachedTo: "ogre" }],
    log: [{ kind: "attach", seat: 1, card_id: "aura", target: "ogre", label: "Aura", batch: 1 }],
  },
  {
    id: "mill",
    title: "Mill three",
    group: "Common moments (S80)",
    issue: 4445,
    shows: ["mill"],
    before: [elf],
    after: [elf, ...milled(3)],
    log: [{ kind: "mill", seat: 1, amount: 3, card_ids: ["m1", "m2", "m3"], batch: 1 }],
  },
  {
    id: "mill_sweep",
    title: "Mill seven (one sweep)",
    group: "Common moments (S80)",
    issue: 4445,
    shows: ["mill"],
    before: [elf],
    after: [elf, ...milled(7)],
    log: [
      {
        kind: "mill",
        seat: 1,
        amount: 7,
        card_ids: milled(7).map((c) => c.id),
        batch: 1,
      },
    ],
  },
  {
    id: "discard",
    title: "Opponent discards two",
    group: "Common moments (S80)",
    issue: 4445,
    shows: ["discard"],
    before: [
      elf,
      { id: "d1", name: "Fact or Fiction", type: "Instant", zone: "hand" },
      { id: "d2", name: "Mystic Remora", type: "Enchantment", zone: "hand" },
      { id: "d3", name: "Island", type: "Basic Land — Island", zone: "hand" },
    ],
    after: [
      elf,
      { id: "d1", name: "Fact or Fiction", type: "Instant", zone: "graveyard" },
      { id: "d2", name: "Mystic Remora", type: "Enchantment", zone: "graveyard" },
      { id: "d3", name: "Island", type: "Basic Land — Island", zone: "hand" },
    ],
    log: ["d1", "d2"].map((id) => ({
      kind: "discard",
      seat: 1,
      card_id: id,
      old_zone: "hand",
      new_zone: "graveyard",
      cause: "effect",
      batch: 1,
    })),
  },
  {
    id: "discard_own",
    title: "You discard one",
    group: "Common moments (S80)",
    issue: 4445,
    shows: ["discard"],
    before: [
      elf,
      { id: "h1", name: "Sol Ring", type: "Artifact", seat: 0, zone: "hand" },
      { id: "h2", name: "Cultivate", type: "Sorcery", seat: 0, zone: "hand" },
      { id: "h3", name: "Forest", type: "Basic Land — Forest", seat: 0, zone: "hand" },
    ],
    after: [
      elf,
      { id: "h1", name: "Sol Ring", type: "Artifact", seat: 0, zone: "hand" },
      { id: "h2", name: "Cultivate", type: "Sorcery", seat: 0, zone: "graveyard" },
      { id: "h3", name: "Forest", type: "Basic Land — Forest", seat: 0, zone: "hand" },
    ],
    log: [
      {
        kind: "discard",
        seat: 0,
        card_id: "h2",
        old_zone: "hand",
        new_zone: "graveyard",
        cause: "effect",
        batch: 1,
      },
    ],
  },
  {
    id: "reanimate",
    title: "Reanimate",
    group: "Common moments (S80)",
    issue: 4446,
    shows: ["reanimate"],
    before: [elf, { id: "griselbrand", name: "Griselbrand", zone: "graveyard" }],
    after: [elf, { id: "griselbrand", name: "Griselbrand" }],
    log: [
      {
        kind: "zone",
        seat: 1,
        card_id: "griselbrand",
        old_zone: "graveyard",
        new_zone: "battlefield",
        batch: 1,
      },
    ],
  },
  {
    id: "regrowth",
    title: "Regrowth to a hand",
    group: "Common moments (S80)",
    issue: 4446,
    shows: ["reanimate"],
    before: [elf, { id: "regrown", name: "Demonic Tutor", type: "Sorcery", zone: "graveyard" }],
    after: [elf, { id: "regrown", name: "Demonic Tutor", type: "Sorcery", zone: "hand" }],
    log: [
      {
        kind: "zone",
        seat: 1,
        card_id: "regrown",
        old_zone: "graveyard",
        new_zone: "hand",
        batch: 1,
      },
    ],
  },
  {
    id: "library_search",
    title: "Library searched",
    group: "Common moments (S80)",
    issue: 4447,
    shows: ["library_search"],
    before: [elf],
    after: [elf],
    log: [
      {
        kind: "search",
        seat: 1,
        target_seat: 1,
        library_seat: 1,
        label: "search_shuffled",
        batch: 1,
        text: "Bea searched their library and shuffled it",
      },
    ],
  },
  {
    id: "transform",
    title: "Transform",
    group: "Common moments (S80)",
    issue: 4448,
    shows: ["transform"],
    before: [werewolf("w1", 0), elf],
    after: [werewolf("w1", 1), elf],
    log: [
      {
        kind: "transform",
        seat: 1,
        card_id: "w1",
        label: "Tovolar's Huntmaster (front)",
        batch: 1,
      },
    ],
  },
  {
    id: "day_night",
    title: "Night falls",
    group: "Common moments (S80)",
    issue: 4448,
    shows: ["day_night", "transform"],
    before: [werewolf("w1", 0), werewolf("w2", 0), elf],
    after: [werewolf("w1", 1), werewolf("w2", 1), elf],
    log: [
      { kind: "day_night", label: "night", batch: 1, text: "It becomes night" },
      {
        kind: "transform",
        seat: 1,
        card_id: "w1",
        label: "Tovolar's Huntmaster (front)",
        batch: 1,
      },
      { kind: "transform", seat: 1, card_id: "w2", label: "Kessig Naturalist (front)", batch: 1 },
    ],
  },

  // ---- Less common moments and players (#4450, #4449, #4544, #4879) ----
  {
    id: "drain_table",
    title: "Table-wide drain",
    group: "Players",
    issue: 4450,
    shows: ["drain_table"],
    before: [{ id: "konrad", name: "Syr Konrad, the Grim", pt: "5/4" }],
    after: [{ id: "konrad", name: "Syr Konrad, the Grim", pt: "5/4" }],
    log: [0, 2, 3].map((seat) => ({
      kind: "life",
      seat,
      card_id: "konrad",
      amount: -1,
      batch: 1,
    })),
  },
  {
    id: "monarch",
    title: "Becomes the monarch",
    group: "Players",
    issue: 4450,
    shows: ["monarch"],
    before: [elf],
    after: [elf],
    log: [{ kind: "monarch", seat: 2, batch: 1 }],
  },
  {
    id: "initiative",
    title: "Takes the initiative",
    group: "Players",
    issue: 4450,
    shows: ["initiative"],
    before: [elf],
    after: [elf],
    log: [{ kind: "initiative", seat: 3, batch: 1 }],
  },
  {
    id: "extort",
    title: "Extort",
    group: "Players",
    issue: 4544,
    shows: ["extort"],
    before: [{ id: "ghast", name: "Crypt Ghast", pt: "2/2" }],
    after: [{ id: "ghast", name: "Crypt Ghast", pt: "2/2" }],
    log: [
      { kind: "resolve", seat: 1, card_id: "ghast", label: "Crypt Ghast — extort", batch: 1 },
      ...[0, 2, 3].map((seat) => ({
        kind: "life" as const,
        seat,
        card_id: "ghast",
        amount: -1,
        batch: 1,
      })),
      { kind: "life", seat: 1, card_id: "ghast", amount: 3, batch: 1 },
    ],
  },
  {
    id: "dice_roll",
    title: "Die rolled",
    group: "Players",
    issue: 4450,
    shows: ["dice_roll"],
    before: [ogre],
    after: [ogre],
    log: [
      {
        kind: "roll",
        seat: 1,
        card_id: "ogre",
        sides: 20,
        results: [14],
        text: "Bea rolled a d20 for Hoarding Ogre: 14",
      },
    ],
  },
  {
    id: "coin_flip",
    title: "Coin flipped",
    group: "Players",
    issue: 4450,
    shows: ["coin_flip"],
    before: [elf],
    after: [elf],
    log: [
      { kind: "flip", seat: 2, faces: ["heads"], call: "heads", text: "Cal called heads: heads" },
    ],
  },
  {
    id: "eliminated",
    title: "Player eliminated",
    group: "Players",
    issue: 4449,
    shows: ["eliminated"],
    before: [elf],
    after: [elf],
    out: [{ seat: 3, cause: "life" }],
    log: [{ kind: "eliminated", seat: 3, cause: "life" }],
  },
  {
    id: "decked_out",
    title: "Player decked out",
    group: "Players",
    issue: 4879,
    shows: ["decked_out"],
    before: [
      elf,
      ...["k1", "k2", "k3", "k4"].map((id) => ({ id, seat: 2 as const, zone: "hand" as const })),
    ],
    after: [elf],
    out: [{ seat: 2, cause: "empty_draw" }],
    log: [{ kind: "eliminated", seat: 2, cause: "empty_draw" }],
  },
];
