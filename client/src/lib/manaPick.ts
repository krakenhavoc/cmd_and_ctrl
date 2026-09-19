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
//
// #986 extends the same contract to a choose_color prompt, whose
// buttons are now ordered by the card's declared `color_purpose`
// (legal.OrderColorOptionsLocked, the function that also orders a bot
// seat's answers). So the client still does not sort — what it does
// with the purpose is SAY IT: five identical buttons cannot tell a
// player whether they are naming the colour their Coldsteel Heart will
// produce or the colour their Wash Out is about to bounce, and those
// are opposite answers.

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

/**
 * Whether a colour prompt can be answered at all: it has at least one
 * button.
 *
 * The server never queues an empty one — CR 903.4f (#844) means a
 * "commander's color identity" source with no identity adds no mana and
 * prompts for nothing, rather than opening a picker with no colours in
 * it — so this is a floor, not a workflow. A prompt nobody can answer
 * must not open a modal that blocks the board.
 */
export function colorPromptAnswerable(choice: {
  kind?: string;
  color_options?: readonly string[];
}): boolean {
  if (choice.kind !== "mana_pick" && choice.kind !== "choose_color") return true;
  return (choice.color_options?.length ?? 0) > 0;
}

/**
 * ColorPurpose mirrors `game.ColorPurpose` — what a choose_color
 * prompt's card will DO with the answer. A prompt from a card nobody
 * has annotated carries no purpose at all, and an unknown string from a
 * newer server is treated the same way.
 */
export type ColorPurpose = "mana" | "benefit" | "harm" | "filter" | "protect";

/** The wording a colour prompt uses: a fallback title and the line under it. */
export interface ColorPromptCopy {
  /** Used when the prompt sends no `reason` of its own. */
  title: string;
  /** Always shown: what the colour is for. */
  hint: string;
}

// The neutral copy, and the one a prompt with no declared purpose
// keeps. It is deliberately vague because it has to be: the same
// prompt comes from a permanent entering (the colour is remembered) and
// from a spell resolving (it is used once), and without a purpose the
// view does not say which.
const UNDECLARED_COLOR_COPY: ColorPromptCopy = {
  title: "Choose a color",
  hint: "Pick exactly one color. The card's text says how it is used.",
};

// One entry per game.ColorPurpose. The hints say what HAPPENS to the
// colour rather than naming a card, because the purpose is a shape of
// question and not a card — the same five lines have to serve Wash Out
// and everything else that ever punishes a colour.
const COLOR_PROMPT_COPY: Record<ColorPurpose, ColorPromptCopy> = {
  mana: {
    title: "Choose a color to produce",
    hint: "This produces mana of the color you choose.",
  },
  benefit: {
    title: "Choose a color to favour",
    hint: "The color you choose is the one that is helped — or the one that survives.",
  },
  harm: {
    title: "Choose a color to hit",
    hint: "Everything of the color you choose is hit, including your own permanents.",
  },
  filter: {
    title: "Choose a color to look for",
    hint: "The color you choose decides which cards the effect acts on. Nothing of yours is at stake.",
  },
  protect: {
    title: "Choose a color to be protected from",
    hint: "The color you choose is the one you are defended against.",
  },
};

/**
 * colorPromptCopy words a choose_color prompt from the purpose the
 * card declared (#780, #986).
 *
 * An absent or unrecognised purpose falls back to the neutral copy
 * rather than guessing: a card the catalog has not annotated must read
 * as vague, not as wrong. Never call it for a `mana_pick` — that prompt
 * has its own copy and is not asking this question.
 */
export function colorPromptCopy(purpose: string | undefined): ColorPromptCopy {
  if (!purpose) return UNDECLARED_COLOR_COPY;
  return COLOR_PROMPT_COPY[purpose as ColorPurpose] ?? UNDECLARED_COLOR_COPY;
}
