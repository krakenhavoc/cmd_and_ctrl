// manaPick builds the colour buttons ChoicePromptModal renders for a
// mana_pick or choose_color prompt, and the per-colour rows a mana
// ability's menu offers (manaAbilityEntries).
//
// The server decides both which colours are legal and their order. An
// "any color" source (Birds of Paradise, Treasure) offers all five
// colours with the chooser's commander colour identity first, so a
// mono-green deck sees G first. Only a card whose printed text says "in
// your commander's color identity" (Command Tower, Arcane Signet) is
// narrowed. The client neither sorts nor filters the list, because
// re-sorting into WUBRG would bury the identity colours the server put
// first.

export interface ColorMeta {
  label: string;
  fill: string;
}

export const COLOR_META: Record<string, ColorMeta> = {
  W: { label: "White", fill: "#f4ead5" },
  U: { label: "Blue", fill: "#aad4ff" },
  B: { label: "Black", fill: "#2b2b3d" },
  R: { label: "Red", fill: "#ff9a85" },
  G: { label: "Green", fill: "#92c493" },
  C: { label: "Colorless", fill: "#c6cfdd" },
};

export interface ColorButton extends ColorMeta {
  color: string;
  /** Tokens this answer adds: 1 unless the pick carries color_amounts (#742). */
  amount: number;
}

/**
 * One button per offered colour, in the order the server sent them.
 * A colour missing from `amounts` adds one mana.
 */
export function colorButtons(
  options: readonly string[] | undefined,
  amounts?: Readonly<Record<string, number>>,
): ColorButton[] {
  return (options ?? []).map((color) => ({
    color,
    ...(COLOR_META[color] ?? { label: color, fill: "#ccc" }),
    amount: amounts?.[color] ?? 1,
  }));
}

/** The fields of a ManaAbilityView the menu rows are built from. */
export interface ManaAbilityColorShape {
  index: number;
  label?: string;
  produced?: string;
  color_options?: string[];
  one_click_color?: string;
}

export interface ManaAbilityEntry {
  /** Stable key: the ability index, plus the colour for an explicit row. */
  key: string;
  label: string;
  /** Sent as `color` on activate_mana_ability; absent on the bare activation. */
  color?: string;
}

/**
 * The menu rows for one mana ability (owner decision 2026-09-17).
 *
 * A bare activation of a source with exactly one commander-identity
 * colour on offer makes that colour in one click: a Scrubland in a
 * mono-white deck gives {W}. The server says so by sending
 * `color_options` (every colour, identity first) and `one_click_color`.
 * The first row is then the bare activation, labelled with the colour
 * it makes, and every other colour gets its own row that names it, so
 * an off-identity colour is one step away. A source whose bare
 * activation prompts (two or more identity colours, or none) carries
 * neither field and keeps its single row; the prompt lists every
 * colour.
 *
 * A row that names a colour on a source with several multi-colour
 * slots (a filter land's {W|U}{W|U}) fixes the first slot and prompts
 * for the rest, hence "first".
 */
export function manaAbilityEntries(a: ManaAbilityColorShape): ManaAbilityEntry[] {
  const base = a.label || a.produced || "add mana";
  const colors = a.color_options ?? [];
  if (colors.length < 2) return [{ key: `${a.index}`, label: base }];
  const oneClick = a.one_click_color;
  const slots = (a.produced?.match(/\{/g) ?? []).length;
  const first = slots > 1 ? " first" : "";
  const rows: ManaAbilityEntry[] = [
    { key: `${a.index}`, label: oneClick ? `${base} → {${oneClick}}` : base },
  ];
  for (const color of colors) {
    if (color === oneClick) continue;
    rows.push({ key: `${a.index}-${color}`, label: `${base} → {${color}}${first}`, color });
  }
  return rows;
}
