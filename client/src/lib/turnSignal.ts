// turnSignal.ts — #3342: tell a viewer in a background tab that the
// table is waiting on them, with the tab title and a dot on the favicon.
//
// "Waiting on you" is not defined here: it is the union of the signals
// the board already reads (priority.ts, timing.ts and the pending-choice
// and discard queues the modals are driven by).

import { applyFavicon, rootFaviconColours } from "./favicon";
import type { GameView } from "./protocol";
import { blockDeclarationPending, owesBlockDecision } from "./priority";
import { hasPriority } from "./timing";

/** True when the live game is waiting on this viewer's seat for something. */
export function waitingOnViewer(
  view: GameView | null | undefined,
  viewerID: string | null,
): boolean {
  if (!view || !viewerID || view.state !== "active") return false;
  const seat = view.seats?.find((s) => s.id === viewerID);
  if (!seat || seat.eliminated) return false;
  if (view.mulligans_open) return seat.mulligan_turn === true;
  return (
    hasPriority(view, viewerID) ||
    owesBlockDecision(view, viewerID) ||
    blockDeclarationPending(view, viewerID) ||
    (view.discard_pending?.[viewerID] ?? 0) > 0 ||
    !!view.pending_choices?.some((c) => c.chooser === viewerID)
  );
}

/** Whether the tab should be signalling: waiting on the viewer AND not looking. */
export function shouldSignal(
  view: GameView | null | undefined,
  viewerID: string | null,
  hidden: boolean,
): boolean {
  return hidden && waitingOnViewer(view, viewerID);
}

export const SIGNAL_PREFIX = "● Your turn — ";

/**
 * Put the signal on the document and return the function that takes it
 * off again, restoring exactly the title and icon the page had.
 */
export function showTurnSignal(doc: Document): () => void {
  const title = doc.title;
  const icon = doc.querySelector('link[rel="icon"]');
  const href = icon?.getAttribute("href") ?? null;
  doc.title = SIGNAL_PREFIX + title;
  applyFavicon(doc, rootFaviconColours(doc), true);
  return () => {
    doc.title = title;
    if (href !== null) icon?.setAttribute("href", href);
  };
}
