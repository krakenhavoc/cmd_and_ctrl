// boardShot — #4116: a picture of the table, taken in the reporter's own
// browser, for the bug report form (ADR 0017 §13).
//
// The board is DOM and images, not a canvas, so there is nothing to
// copy pixels from. html-to-image (MIT, no dependencies) clones the
// element, inlines its images, fonts and computed styles into an SVG
// and rasterises that. Every card image is `/cards/<id>/image`, served
// by this origin, so nothing taints the canvas.
//
// What it captures is what the reporter's screen already shows, because
// it reads their DOM: their hand, face-down cards as backs, nothing the
// server did not send them. No new information leaves the client that
// the reporter could not screenshot themselves, and the form shows the
// picture before it is sent (it is an ordinary attachment, with the
// same remove button, size caps and storage as one they upload).
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
      filter: (node) => !skipsBoardShot(node),
    });
  } catch (e) {
    throw new BoardShotError(
      `couldn't take a picture of the board (${e instanceof Error ? e.message : "unknown error"})`,
    );
  }
  const blob = await encodeUnderCap(
    (type, quality) => new Promise((resolve) => canvas.toBlob(resolve, type, quality)),
  );
  if (!blob) throw new BoardShotError("the picture of the board is too large to attach");
  return new File([blob], boardShotName(blob.type), { type: blob.type });
}
