// manaPick builds the colour buttons ChoicePromptModal renders for a
// mana_pick or choose_color prompt.
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
