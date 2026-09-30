// counterTypes.ts mirrors the server's named counter registry from
// server/internal/game/counter_types.go. The server-side map is the
// source of truth — keep these strings byte-equal to its constants.
//
// Each entry carries an optional pinned colour + glyph for the
// pip-overlay UI. Counter names not in the map render as text-only
// pips with a neutral colour.

// Card-level counter type identifiers.
export const COUNTER_PLUS_ONE = "+1/+1";
export const COUNTER_MINUS_ONE = "-1/-1";
export const COUNTER_LOYALTY = "loyalty";
export const COUNTER_DEFENSE = "defense";
export const COUNTER_CHARGE = "charge";
export const COUNTER_STORAGE = "storage";
export const COUNTER_STUN = "stun";
export const COUNTER_SHIELD = "shield";
export const COUNTER_LORE = "lore";
export const COUNTER_AGE = "age";

// Keyword counter kinds (CR 122.1b, ADR 0101). Mirrors
// keywordCounterKinds in server/internal/game/keyword_counters.go: the
// kind IS the keyword it grants, and the server adds that keyword to the
// card's `abilities`, so the badge row shows it beside this pip.
export const KEYWORD_COUNTER_KINDS: readonly string[] = [
  "flying",
  "first strike",
  "double strike",
  "deathtouch",
  "haste",
  "hexproof",
  "indestructible",
  "lifelink",
  "menace",
  "reach",
  "shadow",
  "trample",
  "vigilance",
];

// Player-level counter type identifiers.
export const COUNTER_POISON = "poison";
export const COUNTER_ENERGY = "energy";
export const COUNTER_EXPERIENCE = "experience";
export const COUNTER_RAD = "rad";

// PoisonLethal mirrors the server's PoisonLethal const — the
// poison-counter total at which a player loses (CR 704.5c).
export const POISON_LETHAL = 10;

export interface CounterStyle {
  // Short label for the pip chip (typically 2-3 chars).
  abbr: string;
  // CSS colour for the pip background.
  color: string;
  // Optional emoji / icon glyph for the player-counter panel.
  glyph?: string;
}

// KEYWORD_COUNTER_COLOR is the pip colour every keyword counter shares.
const KEYWORD_COUNTER_COLOR = "#4f7fbf";

// COUNTER_STYLES is the iconography pin map. Unknown counter names
// fall back to neutral styling (see counterStyle() below).
export const COUNTER_STYLES: Record<string, CounterStyle> = {
  [COUNTER_PLUS_ONE]: { abbr: "+1", color: "#5fbf67", glyph: "▲" },
  [COUNTER_MINUS_ONE]: { abbr: "-1", color: "#c2536b", glyph: "▼" },
  [COUNTER_LOYALTY]: { abbr: "L", color: "#7a5fbf", glyph: "★" },
  [COUNTER_DEFENSE]: { abbr: "D", color: "#3f6e8a", glyph: "🛡" },
  [COUNTER_CHARGE]: { abbr: "C", color: "#bf9a5f", glyph: "⚡" },
  [COUNTER_STORAGE]: { abbr: "St", color: "#8a7f5f", glyph: "🔋" },
  [COUNTER_STUN]: { abbr: "S", color: "#a5a5a5", glyph: "💫" },
  [COUNTER_SHIELD]: { abbr: "Sh", color: "#5f9abf", glyph: "🛡" },
  [COUNTER_LORE]: { abbr: "Lo", color: "#bfa55f", glyph: "📜" },
  [COUNTER_AGE]: { abbr: "Age", color: "#8a7f6e", glyph: "⌛" },
  [COUNTER_POISON]: { abbr: "Poi", color: "#7ab85f", glyph: "🟢" },
  [COUNTER_ENERGY]: { abbr: "E", color: "#e0c050", glyph: "⚡" },
  [COUNTER_EXPERIENCE]: { abbr: "Exp", color: "#d0a050", glyph: "⭐" },
  [COUNTER_RAD]: { abbr: "Rad", color: "#a8c050", glyph: "☢" },
  // Keyword counters (ADR 0101): one colour for the family, and an
  // abbreviation that names the keyword.
  flying: { abbr: "Fly", color: KEYWORD_COUNTER_COLOR },
  "first strike": { abbr: "1st", color: KEYWORD_COUNTER_COLOR },
  "double strike": { abbr: "2x", color: KEYWORD_COUNTER_COLOR },
  deathtouch: { abbr: "DT", color: KEYWORD_COUNTER_COLOR },
  haste: { abbr: "Hst", color: KEYWORD_COUNTER_COLOR },
  hexproof: { abbr: "Hex", color: KEYWORD_COUNTER_COLOR },
  indestructible: { abbr: "Ind", color: KEYWORD_COUNTER_COLOR },
  lifelink: { abbr: "LL", color: KEYWORD_COUNTER_COLOR },
  menace: { abbr: "Men", color: KEYWORD_COUNTER_COLOR },
  reach: { abbr: "Rch", color: KEYWORD_COUNTER_COLOR },
  shadow: { abbr: "Shd", color: KEYWORD_COUNTER_COLOR },
  trample: { abbr: "Trm", color: KEYWORD_COUNTER_COLOR },
  vigilance: { abbr: "Vig", color: KEYWORD_COUNTER_COLOR },
};

const NEUTRAL_STYLE: CounterStyle = { abbr: "?", color: "#6c7a99" };

// PT_COUNTER matches every power/toughness counter name the server's
// game.ParsePTCounter accepts: "[+-]N/[+-]M" (#1664, CR 122.1a).
const PT_COUNTER = /^([+-])(\d{1,3})\/([+-])(\d{1,3})$/;

// counterStyle resolves a counter name to its pin style, falling
// back to a neutral style with the first 3 characters as the abbr.
//
// #1664: a P/T counter other than +1/+1 and -1/-1 (Contagion's -2/-1,
// Dwarven Armorer's +1/+0) shows its WHOLE name — three characters
// would draw "+1/" for both +1/+0 and +1/+2 — in the +1/+1 green when
// it adds more than it takes and the -1/-1 red when it takes more.
export function counterStyle(name: string): CounterStyle {
  const known = COUNTER_STYLES[name];
  if (known) return known;
  const pt = PT_COUNTER.exec(name);
  if (pt) {
    const power = (pt[1] === "-" ? -1 : 1) * Number(pt[2]);
    const toughness = (pt[3] === "-" ? -1 : 1) * Number(pt[4]);
    const base = COUNTER_STYLES[power + toughness < 0 ? COUNTER_MINUS_ONE : COUNTER_PLUS_ONE];
    return { abbr: name, color: base?.color ?? NEUTRAL_STYLE.color };
  }
  return { ...NEUTRAL_STYLE, abbr: name.slice(0, 3) };
}

// PINNED_PLAYER_COUNTERS is the ordered list of counters that the
// player-counter panel (PlayerIdentity) renders as always-visible
// rows. Other counters appear only when non-zero.
export const PINNED_PLAYER_COUNTERS: readonly string[] = [
  COUNTER_POISON,
  COUNTER_ENERGY,
  COUNTER_EXPERIENCE,
  COUNTER_RAD,
];
