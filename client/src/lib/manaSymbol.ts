// manaSymbol.ts — the shapes and colours behind ManaSymbol.svelte
// (#1438), kept out of the component so a test can check them and so
// there is one table of symbol colours rather than one per component.
//
// Every glyph is drawn on a 32×32 box around a disc of radius 15 and
// filled with fill-rule evenodd, which is how the skull's eyes and the
// colourless diamond's middle are cut out.

export interface ManaSymbolMeta {
  /** The colour's name, for tooltips: "Green". */
  name: string;
  /** Disc colour. */
  fill: string;
  /** SVG path for the glyph; empty for a text symbol. */
  glyph: string;
}

// One disc of radius r around (cx, cy), as a path.
function disc(cx: number, cy: number, r: number): string {
  return `M${cx + r} ${cy}A${r} ${r} 0 1 0 ${cx - r} ${cy}A${r} ${r} 0 1 0 ${cx + r} ${cy}Z`;
}

// The sun: a disc and eight rays. Computed rather than written out so
// the rays are exactly evenly spaced.
function sun(): string {
  const parts = [disc(16, 16, 5.2)];
  const round = (n: number): number => Math.round(n * 100) / 100;
  for (let k = 0; k < 8; k++) {
    const a = (k * Math.PI) / 4;
    const spread = (11 * Math.PI) / 180;
    const tip = [16 + 12.5 * Math.cos(a), 16 + 12.5 * Math.sin(a)];
    const left = [16 + 7 * Math.cos(a - spread), 16 + 7 * Math.sin(a - spread)];
    const right = [16 + 7 * Math.cos(a + spread), 16 + 7 * Math.sin(a + spread)];
    parts.push(
      `M${round(tip[0])} ${round(tip[1])}L${round(left[0])} ${round(left[1])}L${round(right[0])} ${round(right[1])}Z`,
    );
  }
  return parts.join("");
}

const DROP = "M16 5.5C16 5.5 8.5 14.5 8.5 19.5A7.5 7.5 0 0 0 23.5 19.5C23.5 14.5 16 5.5 16 5.5Z";

const SKULL = [
  // Cranium and jaw as one outline, so evenodd does not cut the overlap.
  "M10 17.5A8.2 8.2 0 1 1 22 17.5L20.5 19.5L20.5 24Q20.5 25 19.5 25L12.5 25Q11.5 25 11.5 24L11.5 19.5Z",
  disc(12.8, 14.5, 2.4),
  disc(19.2, 14.5, 2.4),
  "M16 17.2L17.1 19.4L14.9 19.4Z",
  "M14.3 21.8h1v2.4h-1Z",
  "M16.7 21.8h1v2.4h-1Z",
].join("");

const FLAME = [
  "M16 4.5C17.5 9 23.5 11.5 23.5 18.5A7.5 7.5 0 0 1 8.5 18.5C8.5 14.5 10.5 12 12.8 10.2C12.6 13 13.6 14.8 15.2 15.6C14.2 11.8 14.6 8 16 4.5Z",
  // The hollow at the heart of the flame.
  "M16 17C17 19 19.5 20 19.5 22.3A3.5 3.5 0 0 1 12.5 22.3C12.5 20.5 14.5 19.5 16 17Z",
].join("");

const TREE =
  "M16 4.5C21 4.5 23.5 8.5 22.8 12C25.5 13 26 17.5 23 19C21 20.2 18.5 19.5 17.5 19L17.5 25Q17.5 26 16.5 26L15.5 26Q14.5 26 14.5 25L14.5 19C13.5 19.5 11 20.2 9 19C6 17.5 6.5 13 9.2 12C8.5 8.5 11 4.5 16 4.5Z";

// Colourless: a hollow diamond.
const DIAMOND = "M16 5.5L26 16L16 26.5L6 16ZM16 10L21.7 16L16 22L10.3 16Z";

// Energy (ADR 0129 §8, CR 107.14): a lightning bolt, on the energy
// counter's own colour (counterTypes.ts).
const BOLT = "M18.5 4.5L8.5 18H15L13 27.5L23.5 13.5H17L18.5 4.5Z";

export const MANA_SYMBOL_META: Readonly<Record<string, ManaSymbolMeta>> = {
  W: { name: "White", fill: "#f8f6d8", glyph: sun() },
  U: { name: "Blue", fill: "#c1d7e9", glyph: DROP },
  B: { name: "Black", fill: "#bab1ab", glyph: SKULL },
  R: { name: "Red", fill: "#e49977", glyph: FLAME },
  G: { name: "Green", fill: "#a3c095", glyph: TREE },
  C: { name: "Colorless", fill: "#cbc2bf", glyph: DIAMOND },
  // Not mana: {E} is one energy counter (CR 107.14). Drawn with the
  // mana symbols because a cost prints it among them.
  E: { name: "Energy", fill: "#e0c050", glyph: BOLT },
};

/** A generic, X, hybrid or snow symbol: grey disc, text on top. */
export const GENERIC_SYMBOL_FILL = "#cac5c0";

export function manaSymbolMeta(symbol: string): ManaSymbolMeta {
  const key = symbol.toUpperCase();
  return MANA_SYMBOL_META[key] ?? { name: key, fill: GENERIC_SYMBOL_FILL, glyph: "" };
}

/** manaSymbols splits brace notation into symbols: "{1}{U}" → ["1", "U"]. */
export function manaSymbols(cost: string): string[] {
  return Array.from(cost.matchAll(/\{([^}]+)\}/g), (m) => m[1]);
}

/** What a symbol is, for drawing and for speaking. */
export type PipKind =
  | "generic"
  | "colour"
  | "colourless"
  | "x"
  | "hybrid"
  | "phyrexian"
  | "snow"
  | "energy";

export interface ManaPip {
  /** The symbol with braces stripped, upper-cased: "U", "2", "W/U", "U/P". */
  symbol: string;
  kind: PipKind;
}

const COLOUR_LETTERS = ["W", "U", "B", "R", "G"];

function pipKind(sym: string): PipKind {
  if (/^\d+$/.test(sym)) return "generic";
  if (sym === "X" || sym === "Y" || sym === "Z") return "x";
  if (sym === "C") return "colourless";
  if (sym === "S") return "snow";
  if (sym === "E") return "energy";
  if (COLOUR_LETTERS.includes(sym)) return "colour";
  if (/\/P$/.test(sym)) return "phyrexian";
  if (sym.includes("/")) return "hybrid";
  return "generic";
}

/**
 * pipRun turns symbols into what is drawn: every generic number is
 * summed into ONE grey circle (`{1}{1}{U}` becomes 2, U), as a person
 * reads a cost; {C}, X, hybrid, Phyrexian and snow stay one pip each.
 * The generic total takes the place of the first generic symbol. A bare
 * `{0}` is a "0" circle; a zero beside other pips is dropped.
 */
export function pipRun(symbols: readonly string[]): ManaPip[] {
  const out: ManaPip[] = [];
  let total = 0;
  let at = -1;
  for (const raw of symbols) {
    const symbol = raw.toUpperCase();
    if (/^\d+$/.test(symbol)) {
      total += Number(symbol);
      if (at < 0) {
        at = out.length;
        out.push({ symbol: "0", kind: "generic" });
      }
      continue;
    }
    out.push({ symbol, kind: pipKind(symbol) });
  }
  if (at >= 0) {
    if (total === 0 && out.length > 1) out.splice(at, 1);
    else out[at] = { symbol: String(total), kind: "generic" };
  }
  return out;
}

/** costPips is pipRun over a brace-notation cost: "{2}{U}" gives 2, U. */
export function costPips(cost: string): ManaPip[] {
  return pipRun(manaSymbols(cost));
}

function colourName(letter: string): string {
  return (MANA_SYMBOL_META[letter]?.name ?? letter).toLowerCase();
}

/** pipWords says one pip aloud: "2 generic", "blue", "blue or white". */
export function pipWords(p: ManaPip): string {
  const half = (s: string): string => (/^\d+$/.test(s) ? `${s} generic` : colourName(s));
  switch (p.kind) {
    case "generic":
      return `${p.symbol} generic`;
    case "colour":
      return colourName(p.symbol);
    case "colourless":
      return "colourless";
    case "x":
      return p.symbol;
    case "snow":
      return "snow";
    case "energy":
      return "energy";
    case "phyrexian":
      return `${colourName(p.symbol.split("/")[0])} or 2 life`;
    default:
      return p.symbol.split("/").map(half).join(" or ");
  }
}

/**
 * costWords is a cost as a fragment for an accessible name: "{2}{U}"
 * becomes "2 generic and 1 blue"; repeated colours are counted.
 */
export function costWords(cost: string): string {
  const counts = new Map<string, number>();
  for (const p of costPips(cost)) {
    const w = pipWords(p);
    counts.set(w, (counts.get(w) ?? 0) + 1);
  }
  const parts = Array.from(counts, ([w, n]) =>
    /^\d+ generic$/.test(w) || w === "X" ? w : `${n} ${w}`,
  );
  if (parts.length <= 1) return parts[0] ?? "";
  return `${parts.slice(0, -1).join(", ")} and ${parts[parts.length - 1]}`;
}
