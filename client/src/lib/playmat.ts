// playmat.ts — ADR 0128. A signed-in person's playmat is one image the
// server stores and every player at the table sees behind THEIR
// battlefield, as with a paper mat.
//
// Pure logic only: which seats draw a mat under the per-device setting,
// and how a wire URL becomes an <img> source. No DOM, no network. The
// calls are in api.ts, the board layer in PlayerPanel.svelte and the
// Settings section in PlaymatSettings.svelte.

import { currentSession } from "./session";
import type { PlaymatsMode } from "./playmatMode";
import type { PlaymatCrop, PlaymatSuggestion } from "./api";

export {
  DEFAULT_PLAYMATS_MODE,
  PLAYMATS_MODES,
  isPlaymatsMode,
  type PlaymatsMode,
} from "./playmatMode";

// The only shape the server mints. A wire value that is not exactly
// this is not loaded, whatever it says: the board must never be a way
// to make a browser contact a third-party host, which is the whole
// reason the server stores the bytes (ADR 0128 §3).
const PLAYMAT_PATH = /^\/playmats\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

/** isPlaymatPath reports whether url is a playmat route this server minted. */
export function isPlaymatPath(url: unknown): url is string {
  return typeof url === "string" && PLAYMAT_PATH.test(url);
}

/**
 * playmatSrc turns a wire playmat_url into the src an <img> loads, or
 * null for anything that is not a playmat path.
 *
 * The route is session-gated and an <img> cannot set an Authorization
 * header, so the token rides as ?token=, as avatarURL's does: the
 * session cookie alone is not reliable (a Secure-flag mismatch, a
 * cleared cookie with a live localStorage session).
 */
export function playmatSrc(url: string | undefined): string | null {
  if (!isPlaymatPath(url)) return null;
  const token = currentSession()?.token;
  return token ? `${url}?token=${encodeURIComponent(token)}` : url;
}

/**
 * playmatShownFor is the one rule for whether a seat's mat is drawn:
 * the setting, then whether the seat is the viewer's own. It returns
 * the wire URL, or null for none.
 */
export function playmatShownFor(
  mode: PlaymatsMode,
  seat: { playmat_url?: string },
  isSelf: boolean,
): string | null {
  if (mode === "off") return null;
  if (mode === "mine" && !isSelf) return null;
  return isPlaymatPath(seat.playmat_url) ? seat.playmat_url : null;
}

/** The server's cap for a pasted or uploaded image, for the form's hint. */
export const PLAYMAT_MAX_BYTES = 10 * 1024 * 1024;

// ---- the best-size suggestion (ADR 0128 §11) ----
//
// The server decides what fits and what the crop is; the client only
// lets the person slide the crop window along the axis being cropped and
// says where it ended up. Everything here is arithmetic on the numbers
// the server sent.

/** The crop's origin, in the stored image's own pixels. */
export interface CropOrigin {
  x: number;
  y: number;
}

/**
 * cropAxis is the one axis the crop window can slide along: "x" for a
 * wide image (the sides are trimmed), "y" for a tall one, null when the
 * crop is the whole image and there is nothing to choose.
 */
export function cropAxis(imgW: number, imgH: number, crop: PlaymatCrop): "x" | "y" | null {
  if (crop.width < imgW) return "x";
  if (crop.height < imgH) return "y";
  return null;
}

/**
 * clampCropOrigin keeps the window inside the image and on its axis: the
 * other axis stays where the server centred it (it has no room to move).
 */
export function clampCropOrigin(
  imgW: number,
  imgH: number,
  crop: PlaymatCrop,
  at: CropOrigin,
): CropOrigin {
  const axis = cropAxis(imgW, imgH, crop);
  const clamp = (v: number, hi: number) => Math.round(Math.min(Math.max(v, 0), Math.max(hi, 0)));
  return {
    x: axis === "x" ? clamp(at.x, imgW - crop.width) : crop.x,
    y: axis === "y" ? clamp(at.y, imgH - crop.height) : crop.y,
  };
}

/**
 * nudgeCrop moves the window one keyboard step along its axis. dir is
 * -1 for up/left and +1 for down/right; the step is a fraction of the
 * room it has to move, at least one pixel. "home" and "end" jump to an
 * end.
 */
export function nudgeCrop(
  imgW: number,
  imgH: number,
  crop: PlaymatCrop,
  at: CropOrigin,
  move: -1 | 1 | "home" | "end",
  fraction = 0.05,
): CropOrigin {
  const axis = cropAxis(imgW, imgH, crop);
  if (axis === null) return { x: crop.x, y: crop.y };
  const room = axis === "x" ? imgW - crop.width : imgH - crop.height;
  const cur = axis === "x" ? at.x : at.y;
  const next =
    move === "home"
      ? 0
      : move === "end"
        ? room
        : cur + move * Math.max(1, Math.round(room * fraction));
  return clampCropOrigin(
    imgW,
    imgH,
    crop,
    axis === "x" ? { x: next, y: at.y } : { x: at.x, y: next },
  );
}

/**
 * dragCrop turns a pointer's travel, in CSS pixels over the shown image
 * (boxW x boxH), into a new origin in image pixels, from where the drag
 * started.
 */
export function dragCrop(
  imgW: number,
  imgH: number,
  crop: PlaymatCrop,
  start: CropOrigin,
  dxPx: number,
  dyPx: number,
  boxW: number,
  boxH: number,
): CropOrigin {
  if (boxW <= 0 || boxH <= 0) return start;
  return clampCropOrigin(imgW, imgH, crop, {
    x: start.x + (dxPx * imgW) / boxW,
    y: start.y + (dyPx * imgH) / boxH,
  });
}

/** cropBoxStyle is the window's place over the shown image, in percent. */
export function cropBoxStyle(
  imgW: number,
  imgH: number,
  crop: PlaymatCrop,
  at: CropOrigin,
): { left: string; top: string; width: string; height: string } {
  const pct = (n: number, of: number) => `${((n / of) * 100).toFixed(3)}%`;
  return {
    left: pct(at.x, imgW),
    top: pct(at.y, imgH),
    width: pct(crop.width, imgW),
    height: pct(crop.height, imgH),
  };
}

/** The prompt's first line, as Ian wrote it. */
export function fitPromptText(imgW: number, imgH: number, idealW: number, idealH: number): string {
  return `This image is ${imgW}\u00d7${imgH}. Playmats look best at ${idealW}\u00d7${idealH} (the shape of a paper playmat).`;
}

/** What accepting will leave behind, and a warning when it will be small. */
export function fitResultText(s: PlaymatSuggestion): string {
  const size = `${s.target_width}\u00d7${s.target_height}`;
  return s.smaller
    ? `Fitting keeps ${size}, the most this image has at that shape. It is smaller than the best size, so it may look soft at the table.`
    : `Fitting keeps the highlighted part and saves it at ${size}.`;
}
