// hintContexts.ts — the contexts the hint tests and the label guard read
// every hint's anchor over (ADR 0125 §2.3 rule 3, §8), and the fixture
// hints PR 4 ships in place of real ones.

import { collectHints } from "../hints";
import { HINT_PLACES, type Hint, type HintContext, type HintPlace } from "../hints/hint";
import type { TableMoment } from "../hints/tableMoment";
import type { GameView, StackItemView } from "../protocol";
import { BOT, ME, board, elves, forest } from "./tutorialBoards";

/** The fixture hints under lib/test/hints/, which never ship. */
export const FIXTURE_HINTS: readonly Hint[] = collectHints(
  import.meta.glob("./hints/*.hint.ts", { eager: true }),
);

/** moment is a quiet table moment, long settled, with `over` applied. */
export function moment(over: Partial<TableMoment> = {}): TableMoment {
  return {
    since: 0,
    dockRequest: false,
    dialog: false,
    gesture: false,
    viewerHasPriority: false,
    stackTopController: null,
    viewerID: ME,
    coachVisible: false,
    ...over,
  };
}

const PAGE_PLACES = HINT_PLACES.filter(
  (p): p is Exclude<HintPlace, "site" | "settings" | "table"> =>
    p !== "site" && p !== "settings" && p !== "table",
);

/** A page context for each kind of visitor. */
function pageContexts(place: HintPlace): HintContext[] {
  const out: HintContext[] = [];
  const who = [
    { signedIn: false, adminMode: false, adminToken: false },
    { signedIn: true, adminMode: false, adminToken: false },
    { signedIn: true, adminMode: true, adminToken: false },
    { signedIn: false, adminMode: false, adminToken: true },
  ];
  for (const w of who) {
    for (const hasEndedGame of [null, false, true]) {
      out.push({
        place,
        route: place,
        ...w,
        hasEndedGame,
        moment: null,
        view: null,
        viewerID: null,
      });
    }
  }
  return out;
}

function item(controller: string, i: number): StackItemView {
  return {
    id: `item-${i}`,
    kind: "spell",
    controller,
    owner: controller,
    source_card_id: `card-${i}`,
  };
}

/** Table views: steps, both seats active, an empty stack and one from each side. */
function tableViews(): GameView[] {
  const out: GameView[] = [];
  for (const step of ["upkeep", "precombat_main", "declare_attackers", "end"]) {
    for (const active of [0, 1]) {
      for (const mine of [[], [forest(), elves()]]) {
        for (const top of [null, ME, BOT]) {
          const v = board({ step, active, mine, stack: top ? [elves()] : [] });
          if (top) v.stack_items = [item(top, 1)];
          out.push(v);
        }
      }
    }
  }
  return out;
}

/**
 * hintContexts is every context a hint's anchor is read over: its own
 * place (every page for a "site" hint), for each kind of visitor, and at
 * the table over the tutorial's boards.
 */
export function hintContexts(h: Hint): HintContext[] {
  if (h.place === "table") {
    return tableViews().flatMap((view) =>
      [moment(), moment({ viewerHasPriority: true })].map((m) => ({
        ...pageContexts("table")[1],
        route: "game",
        moment: { ...m, stackTopController: view.stack_items?.at(-1)?.controller ?? null },
        view,
        viewerID: ME,
      })),
    );
  }
  const places = h.place === "site" ? PAGE_PLACES : [h.place];
  return places.flatMap(pageContexts);
}
