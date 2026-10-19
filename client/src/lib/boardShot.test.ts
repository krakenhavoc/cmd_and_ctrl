// @vitest-environment jsdom
//
// #4116: the picture of the board for the bug report. html-to-image does
// the rasterising and needs real layout, so it is checked in the browser
// (tests-e2e/tests/bug-report-board-shot-4116.spec.ts). What is pinned
// here is everything that is decision: what is left out, which encoding
// fits under the cap, and the name.

import { describe, expect, it, vi } from "vitest";

import { BUG_MAX_IMAGE_BYTES, BUG_MAX_TOTAL_IMAGE_BYTES } from "./bugReport";
import {
  BOARD_SHOT_ENCODINGS,
  BOARD_SHOT_PLACEHOLDER_ATTR,
  BOARD_SHOT_MAX_BYTES,
  BOARD_SHOT_SKIP_ATTR,
  BoardShotError,
  boardShotName,
  boardShotFilter,
  captureBoardShot,
  encodeUnderCap,
  maskPrivate,
  skipsBoardShot,
} from "./boardShot";

describe("skipsBoardShot", () => {
  it("leaves out the marked form and nothing else", () => {
    const form = document.createElement("div");
    form.setAttribute(BOARD_SHOT_SKIP_ATTR, "");
    expect(skipsBoardShot(form)).toBe(true);
    expect(skipsBoardShot(document.createElement("div"))).toBe(false);
    // Text and comment nodes come through the filter too.
    expect(skipsBoardShot(document.createTextNode("x"))).toBe(false);
  });

  it("is the attribute the report form carries", async () => {
    const src = await import("./components/BugReportModal.svelte?raw");
    expect(src.default).toContain(BOARD_SHOT_SKIP_ATTR);
  });
});

describe("the size cap", () => {
  it("is under the per-image cap and leaves room in the per-report cap", () => {
    expect(BOARD_SHOT_MAX_BYTES).toBeLessThanOrEqual(BUG_MAX_IMAGE_BYTES);
    expect(BOARD_SHOT_MAX_BYTES).toBeLessThan(BUG_MAX_TOTAL_IMAGE_BYTES / 2);
  });
});

describe("encodeUnderCap", () => {
  const blobOf = (type: string, size: number) => new Blob([new Uint8Array(size)], { type });

  it("takes the PNG when it fits", async () => {
    const out = await encodeUnderCap(async (type) => blobOf(type, 100), 1000);
    expect(out?.type).toBe("image/png");
  });

  it("falls to a lossier JPEG, in order, until one fits", async () => {
    const tried: Array<[string, number | undefined]> = [];
    const out = await encodeUnderCap(async (type, quality) => {
      tried.push([type, quality]);
      return blobOf(type, type === "image/png" ? 5000 : quality === 0.85 ? 2000 : 500);
    }, 1000);
    expect(out?.type).toBe("image/jpeg");
    expect(out?.size).toBe(500);
    expect(tried).toEqual([
      ["image/png", undefined],
      ["image/jpeg", 0.85],
      ["image/jpeg", 0.6],
    ]);
  });

  it("gives up when nothing fits, and when the browser cannot encode", async () => {
    expect(await encodeUnderCap(async (type) => blobOf(type, 5000), 1000)).toBeNull();
    expect(await encodeUnderCap(async () => null, 1000)).toBeNull();
  });

  it("does not accept a type the browser fell back to", async () => {
    // toBlob answers PNG when it does not know the asked type; that is
    // not the JPEG that was asked for and must not pass as one.
    const asked: string[] = [];
    const out = await encodeUnderCap(async (type) => {
      asked.push(type);
      return blobOf("image/png", type === "image/png" ? 5000 : 10);
    }, 1000);
    expect(out).toBeNull();
    expect(asked).toHaveLength(BOARD_SHOT_ENCODINGS.length);
  });
});

describe("boardShotName", () => {
  it("is stamped, and ends in the extension the type has", () => {
    const at = new Date("2026-10-19T12:34:56Z");
    expect(boardShotName("image/png", at)).toBe("board-20261019-123456.png");
    expect(boardShotName("image/jpeg", at)).toBe("board-20261019-123456.jpg");
  });
});

describe("captureBoardShot", () => {
  it("says so, in words, when the picture cannot be made", async () => {
    vi.resetModules();
    vi.doMock("html-to-image", () => ({
      toCanvas: async () => {
        throw new Error("boom");
      },
    }));
    const { captureBoardShot: capture, BoardShotError: Err } = await import("./boardShot");
    await expect(capture(document.createElement("div"))).rejects.toBeInstanceOf(Err);
    await expect(capture(document.createElement("div"))).rejects.toThrow(/boom/);
    vi.doUnmock("html-to-image");
  });

  it("returns an attachable file when the canvas encodes", async () => {
    vi.resetModules();
    const canvas = {
      toBlob: (cb: (b: Blob | null) => void, type: string) =>
        cb(new Blob(["x"], { type: type || "image/png" })),
    };
    let asked: Record<string, unknown> | null = null;
    vi.doMock("html-to-image", () => ({
      toCanvas: async (_root: unknown, opts: Record<string, unknown>) => {
        asked = opts;
        return canvas;
      },
    }));
    const { captureBoardShot: capture } = await import("./boardShot");
    const file = await capture(document.createElement("div"));
    expect(file).toBeInstanceOf(File);
    expect(file.type).toBe("image/png");
    expect(file.name).toMatch(/^board-\d{8}-\d{6}\.png$/);
    // The form is filtered out of the picture.
    const filter = (asked as unknown as { filter: (n: Node) => boolean }).filter;
    const form = document.createElement("div");
    form.setAttribute(BOARD_SHOT_SKIP_ATTR, "");
    expect(filter(form)).toBe(false);
    expect(filter(document.createElement("div"))).toBe(true);
    vi.doUnmock("html-to-image");
  });
});

// Keeps the static import used, so a rename of the symbol fails here.
void captureBoardShot;
void BoardShotError;

// ADR 0017 section 13: what only the reporter may see is covered, not
// pictured. The filter is what html-to-image is given, so what it drops is
// what the picture's source lacks.
describe("covering what is private", () => {
  const img = (src: string) => `<img src="${src}" alt="x">`;
  function table(): HTMLElement {
    const root = document.createElement("div");
    root.innerHTML = `
      <div class="hand"><div class="hand-slot" id="mine">${img("/cards/hand1/image")}<span>Sol Ring</span></div></div>
      <div class="hand opponent"><div class="hand-slot" id="theirs">${img("/cards/back/image")}</div></div>
      <div class="pile" data-pile="library"><span class="thumb" id="libthumb">${img("/cards/top/image")}</span></div>
      <div data-pile="sideboard" id="side">${img("/cards/side/image")}</div>
      <div data-library-top-peek><div class="peek-card" id="peek">${img("/cards/peek/image")}</div><button class="pile-action">cast</button></div>
      <div role="dialog" aria-label="pick" id="prompt">${img("/cards/choice/image")}<span>Black Lotus</span></div>
      <div class="card" id="morph" data-board-shot-private>${img("/cards/morph/image")}</div>
      <div class="card" id="public">${img("/cards/public/image")}</div>
      <div role="dialog" id="form" data-board-shot-skip>${img("/cards/form/image")}</div>`;
    document.body.appendChild(root);
    return root;
  }
  // Every node html-to-image would keep: the filter is asked of each one.
  function kept(root: HTMLElement): { imgs: string[]; text: string } {
    const imgs: string[] = [];
    let text = "";
    const walk = (n: Node): void => {
      if (n !== root && !boardShotFilter(n)) return;
      if (n instanceof HTMLImageElement) imgs.push(n.getAttribute("src") ?? "");
      if (n.nodeType === 3) text += n.textContent;
      n.childNodes.forEach(walk);
    };
    walk(root);
    return { imgs, text };
  }

  it("keeps public cards and drops every private image and name from the source", () => {
    const root = table();
    const unmask = maskPrivate(root);
    const { imgs, text } = kept(root);
    expect(imgs.sort()).toEqual(["/cards/back/image", "/cards/public/image"]);
    expect(text).not.toContain("Sol Ring");
    expect(text).not.toContain("Black Lotus");
    // The peek's cast button is not a card: it stays.
    expect(text).toContain("cast");
    unmask();
    root.remove();
  });

  it("puts a 'hidden' box in each covered element, so the layout stays", () => {
    const root = table();
    const unmask = maskPrivate(root);
    for (const id of ["mine", "libthumb", "side", "peek", "prompt", "morph"]) {
      const host = root.querySelector(`#${id}`)!;
      const box = host.querySelector(`[${BOARD_SHOT_PLACEHOLDER_ATTR}]`);
      expect(box?.textContent, id).toBe("hidden");
    }
    // Not an opponent's hand (backs), nor a public card.
    expect(root.querySelector(`#theirs [${BOARD_SHOT_PLACEHOLDER_ATTR}]`)).toBeNull();
    expect(root.querySelector(`#public [${BOARD_SHOT_PLACEHOLDER_ATTR}]`)).toBeNull();
    unmask();
    root.remove();
  });

  it("takes the boxes off again and restores what it touched", () => {
    const root = table();
    const before = root.innerHTML;
    maskPrivate(root)();
    expect(root.innerHTML).toBe(before);
    root.remove();
  });

  it("covers only the outermost of nested private elements", () => {
    const root = document.createElement("div");
    root.innerHTML = `<div role="dialog" id="outer"><div class="card" data-board-shot-private id="inner">${img("/cards/x/image")}</div></div>`;
    document.body.appendChild(root);
    const unmask = maskPrivate(root);
    expect(root.querySelectorAll(`[${BOARD_SHOT_PLACEHOLDER_ATTR}]`)).toHaveLength(1);
    expect(root.querySelector(`#outer > [${BOARD_SHOT_PLACEHOLDER_ATTR}]`)).not.toBeNull();
    unmask();
    root.remove();
  });
});
