// skins.ts — the player-selectable skins and the custom accent colour.
//
// A skin is a block of colour tokens in app.css keyed on
// :root[data-theme="<id>"]; the bare :root block is the default skin, so
// first paint is right before settings load. display.theme stores the
// id (a synced setting, so a skin follows the player to every device).
// display.accent optionally overrides the skin's accent with one colour;
// accentVars derives the rest of the accent family from it.

export interface SkinInfo {
  id: string;
  name: string;
  blurb: string;
  // Preview chips for the picker: page, panel, text, accent.
  swatches: readonly [string, string, string, string];
}

export const SKINS = [
  {
    id: "warroom",
    name: "War room",
    blurb: "Navy steel with a signal-orange accent.",
    swatches: ["#070d14", "#121e2c", "#e8eff7", "#ff7a1a"],
  },
  {
    id: "arcade",
    name: "Arcade night",
    blurb: "Deep indigo with a hot-pink accent.",
    swatches: ["#0c0a1a", "#1a1736", "#f0edff", "#ff4f9a"],
  },
  {
    id: "prism",
    name: "Prism",
    blurb: "Cool graphite with an electric-violet accent.",
    swatches: ["#0b0c10", "#181b22", "#eef0f5", "#8f6bff"],
  },
  {
    id: "classic",
    name: "Classic",
    blurb: "Warm near-black with the original gold.",
    swatches: ["#0b0a09", "#1a1816", "#f1ece2", "#d9b45c"],
  },
  {
    id: "light",
    name: "Light",
    blurb: "Warm paper with a deep-gold accent.",
    swatches: ["#f3efe6", "#ffffff", "#1c1814", "#9a7a2c"],
  },
  {
    id: "high-contrast",
    name: "High contrast",
    blurb: "Pure black, white text and a bright yellow accent.",
    swatches: ["#000000", "#0a0a0a", "#ffffff", "#ffd400"],
  },
] as const satisfies readonly SkinInfo[];

export type Skin = (typeof SKINS)[number]["id"];

// The skins the Settings picker offers. Light is a valid skin (a
// stored "light" is kept) but stays out of the picker until the table's
// on-art chips and empty piles have light-skin styles.
export const PICKER_SKINS = SKINS.filter((s) => s.id !== "light");

export const DEFAULT_SKIN: Skin = "warroom";

export function isSkin(v: unknown): v is Skin {
  return typeof v === "string" && SKINS.some((s) => s.id === v);
}

const HEX = /^#[0-9a-f]{6}$/i;

// normalizeAccent returns a lower-case #rrggbb, or "" (use the skin's
// own accent) for anything else.
export function normalizeAccent(v: unknown): string {
  return typeof v === "string" && HEX.test(v.trim()) ? v.trim().toLowerCase() : "";
}

// The accent tokens a custom accent overrides. Everything else in the
// accent family (--accent-line, --accent-glow, the --gold aliases) is
// already derived from --accent in app.css.
export const ACCENT_VARS = ["--accent", "--accent-strong", "--accent-soft", "--accent-fg"] as const;

function luminance(hex: string): number {
  const ch = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255);
  const [r, g, b] = ch.map((c) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

// accentFg picks near-black or white for text on an accent fill,
// whichever has the higher contrast ratio.
export function accentFg(hex: string): string {
  const l = luminance(hex);
  const onDark = (l + 0.05) / (0.012 + 0.05); // vs #111
  const onLight = 1.05 / (l + 0.05); // vs #fff
  return onDark >= onLight ? "#111111" : "#ffffff";
}

// accentVars maps a custom accent to the token values to set on :root.
export function accentVars(hex: string): Record<(typeof ACCENT_VARS)[number], string> {
  return {
    "--accent": hex,
    "--accent-strong": `color-mix(in oklab, ${hex} 65%, white)`,
    "--accent-soft": `color-mix(in srgb, ${hex} 16%, transparent)`,
    "--accent-fg": accentFg(hex),
  };
}
