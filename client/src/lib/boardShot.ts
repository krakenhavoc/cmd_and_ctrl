// boardShot — #4116: a picture of the table, taken in the reporter's own
// browser, for the bug report form (ADR 0017 §13).
//
// The board is DOM and images, not a canvas, so there is nothing to
// copy pixels from. html-to-image (MIT, no dependencies) clones the
// element, inlines its images, fonts and computed styles into an SVG
// and rasterises that. Every card image is `/cards/<id>/image`, served
// by this origin, so nothing taints the canvas.
//
// It reads the reporter's own DOM, so left alone it would hold what only
// they may see: their hand, a face-down card of theirs they may look at, a
// library top they may peek at, the cards they own outside the game, and
// whatever an open prompt lists. A report is a public GitHub issue and the
// game may be live, so none of that goes in (ADR 0017 section 13): before
// the picture is taken each such element gets a neutral "hidden" box of the
// same size laid over it, and the filter drops everything inside it from
// the cloned source, so no card image, name or text of it is ever part of
// what is rendered. The layout stays readable, and what is public (the
// battlefield, the stack, graveyards, exile, the commander zone, life,
// counters, the log) is as the reporter sees it. The form shows the picture
// before it is sent; it is an ordinary attachment, with the same remove
// button, size caps and storage as one they upload.
//
// The library is imported on demand, so it stays out of the bundle
// that loads with the table.

import { BUG_MAX_IMAGE_BYTES, BUG_MAX_TOTAL_IMAGE_BYTES } from "./bugReport";

/** The widest picture we make, in px. A 4K table is scaled down to this. */
export const BOARD_SHOT_MAX_WIDTH = 1920;

/**
 * The largest picture we make: under the per-image cap, and 30% of the
 * per-report cap, which leaves room for a pasted screenshot or two.
 */
export const BOARD_SHOT_MAX_BYTES = Math.min(
  BUG_MAX_IMAGE_BYTES,
  Math.floor(BUG_MAX_TOTAL_IMAGE_BYTES * 0.3),
);

/** An element carrying this attribute (and what is inside it) is left out of the picture. */
export const BOARD_SHOT_SKIP_ATTR = "data-board-shot-skip";

/** The encodings tried in order until one fits under the cap. */
export const BOARD_SHOT_ENCODINGS: ReadonlyArray<{ type: string; quality?: number }> = [
  { type: "image/png" },
  { type: "image/jpeg", quality: 0.85 },
  { type: "image/jpeg", quality: 0.6 },
  { type: "image/jpeg", quality: 0.4 },
];

/** True for a node the picture leaves out: the form itself, and anything marked. */
export function skipsBoardShot(node: Node): boolean {
  return node.nodeType === 1 && (node as Element).hasAttribute(BOARD_SHOT_SKIP_ATTR);
}

/**
 * An element (a card, an overlay) that shows something only its viewer may
 * see, marked in its component. The picture covers it with "hidden".
 */
export const BOARD_SHOT_PRIVATE_ATTR = "data-board-shot-private";
/** Set on an element while it is covered. */
const HIDE_ATTR = "data-board-shot-hide";
/** Set on the box that covers it. */
export const BOARD_SHOT_PLACEHOLDER_ATTR = "data-board-shot-placeholder";

/**
 * What is covered. A whitelist of what is public would be safer still but
 * the board is too varied; this is the list of what is private, and a new
 * private surface marks itself with BOARD_SHOT_PRIVATE_ATTR.
 *
 * - your hand (an opponent's is card backs and stays);
 * - a face-down or foretold card you may look at (Card sets the attribute);
 * - the top of a library, on the pile and on the play peek;
 * - the sideboard pile, the cards you own outside the game;
 * - any open prompt or dialog, which may list cards only you can see.
 */
export const BOARD_SHOT_PRIVATE_SELECTOR = [
  ".hand:not(.opponent) .hand-slot",
  `[${BOARD_SHOT_PRIVATE_ATTR}]`,
  '[data-pile="library"] .thumb',
  '[data-pile="sideboard"]',
  "[data-library-top-peek] .peek-card",
  `[role="dialog"]:not([${BOARD_SHOT_SKIP_ATTR}])`,
  `[role="alertdialog"]:not([${BOARD_SHOT_SKIP_ATTR}])`,
].join(", ");

/**
 * Cover every private element under `root` with a "hidden" box of its own
 * size. Returns the function that takes the boxes off again; call it when
 * the capture is done, whatever happened. Only the outermost of nested
 * private elements is covered.
 */
export function maskPrivate(root: HTMLElement): () => void {
  const all = Array.from(root.querySelectorAll<HTMLElement>(BOARD_SHOT_PRIVATE_SELECTOR));
  const outer = all.filter((el) => !all.some((o) => o !== el && o.contains(el)));
  const undo: Array<() => void> = [];
  for (const el of outer) {
    const doc = el.ownerDocument;
    const box = doc.createElement("span");
    box.setAttribute(BOARD_SHOT_PLACEHOLDER_ATTR, "");
    box.textContent = "hidden";
    Object.assign(box.style, {
      position: "absolute",
      inset: "0",
      boxSizing: "border-box",
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      background: "#232a36",
      border: "1px dashed #4a5568",
      borderRadius: "inherit",
      color: "#8791a5",
      font: "600 10px/1 monospace",
      letterSpacing: "0.08em",
      textTransform: "uppercase",
      zIndex: "1000",
      pointerEvents: "none",
    });
    const wasStatic = (doc.defaultView ?? window).getComputedStyle(el).position === "static";
    const before = el.style.position;
    if (wasStatic) el.style.position = "relative";
    el.setAttribute(HIDE_ATTR, "");
    el.appendChild(box);
    undo.push(() => {
      box.remove();
      el.removeAttribute(HIDE_ATTR);
      if (wasStatic) el.style.position = before;
    });
  }
  return () => undo.forEach((f) => f());
}

/**
 * True for a node inside a covered element: everything but the element
 * itself and its "hidden" box. Dropping these from the clone is what keeps
 * a private card's image, name and text out of the picture's source.
 */
export function hiddenFromBoardShot(node: Node): boolean {
  const el = node.nodeType === 1 ? (node as Element) : node.parentElement;
  if (!el) return false;
  if (el.closest(`[${BOARD_SHOT_PLACEHOLDER_ATTR}]`)) return false;
  const host = el.closest(`[${HIDE_ATTR}]`);
  return host !== null && host !== node;
}

/** The filter html-to-image is given: out are the form and what is covered. */
export function boardShotFilter(node: Node): boolean {
  return !skipsBoardShot(node) && !hiddenFromBoardShot(node);
}

/** The first encoding whose bytes fit under `cap`, or null when none does. */
export async function encodeUnderCap(
  toBlob: (type: string, quality?: number) => Promise<Blob | null>,
  cap: number = BOARD_SHOT_MAX_BYTES,
): Promise<Blob | null> {
  for (const enc of BOARD_SHOT_ENCODINGS) {
    const blob = await toBlob(enc.type, enc.quality);
    if (blob && blob.type === enc.type && blob.size <= cap) return blob;
  }
  return null;
}

/** Why a picture could not be made, in words for the reporter. */
export class BoardShotError extends Error {}

/** The file name a captured board is attached under. */
export function boardShotName(type: string, at: Date = new Date()): string {
  const stamp = at.toISOString().slice(0, 19).replace(/[-:]/g, "").replace("T", "-");
  return `board-${stamp}.${type === "image/png" ? "png" : "jpg"}`;
}

/**
 * Take the picture of `root` as a File, ready to attach. The form is
 * left out of it. Throws a BoardShotError with a reason fit to show.
 */
export async function captureBoardShot(root: HTMLElement): Promise<File> {
  let canvas: HTMLCanvasElement;
  const unmask = maskPrivate(root);
  try {
    const { toCanvas } = await import("html-to-image");
    const width = root.offsetWidth || root.getBoundingClientRect().width;
    const pixelRatio = Math.min(1, BOARD_SHOT_MAX_WIDTH / Math.max(1, width));
    canvas = await toCanvas(root, {
      pixelRatio,
      cacheBust: false,
      // A card image that fails to load is a hole in the picture, not
      // a failed picture.
      imagePlaceholder:
        "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7",
      backgroundColor: getComputedStyle(document.body).backgroundColor || undefined,
      filter: boardShotFilter,
    });
  } catch (e) {
    throw new BoardShotError(
      `couldn't take a picture of the board (${e instanceof Error ? e.message : "unknown error"})`,
    );
  } finally {
    unmask();
  }
  const blob = await encodeUnderCap(
    (type, quality) => new Promise((resolve) => canvas.toBlob(resolve, type, quality)),
  );
  if (!blob) throw new BoardShotError("the picture of the board is too large to attach");
  return new File([blob], boardShotName(blob.type), { type: blob.type });
}
