// @vitest-environment jsdom
//
// #4116: the report form's "include a picture of the board" box. It is
// offered only on a table (when the host names an element to
// photograph), starts unticked, takes the picture only when ticked, puts
// it among the screenshots where it can be removed, and unticking or
// removing it takes it out. Nothing is captured unless asked.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const captureBoardShot = vi.fn();

vi.mock("./boardShot", async (importOriginal) => {
  const orig = await importOriginal<typeof import("./boardShot")>();
  return { ...orig, captureBoardShot: (...a: unknown[]) => captureBoardShot(...a) };
});

import BugReportModal from "./components/BugReportModal.svelte";
import { BoardShotError } from "./boardShot";
import { _resetForTests as resetDraft } from "./bugReportDraft.svelte";
import { _resetForTests as resetModalLayers } from "./modalLayers";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

const shotFile = (name = "board-20261019-123456.png") =>
  new File(["png"], name, { type: "image/png" });

const box = (): HTMLInputElement | null =>
  document.querySelector<HTMLInputElement>('.board-shot input[type="checkbox"]');
const thumbs = (): string[] =>
  Array.from(document.querySelectorAll<HTMLImageElement>(".shots img")).map(
    (i) => i.getAttribute("alt") ?? "",
  );

function mountForm(extra: Record<string, unknown> = {}): void {
  render(
    BugReportModal as never,
    {
      gameID: "5462880a-a13c-4704-9aab-2a733172b302",
      view: null,
      seq: 3,
      connection: "connected",
      wsLog: [],
      attachments: true,
      captureTarget: () => document.body,
      onclose: () => {},
      onminimize: () => {},
      ...extra,
    } as never,
  );
}

beforeEach(() => {
  let n = 0;
  URL.createObjectURL = vi.fn(() => `blob:test/${++n}`);
  URL.revokeObjectURL = vi.fn();
  captureBoardShot.mockReset();
  resetDraft();
  resetModalLayers();
});
afterEach(() => {
  cleanup();
  resetDraft();
});

describe("the board picture box on the report form", () => {
  it("is offered on a table, unticked, and takes no picture until ticked", async () => {
    mountForm();
    await settle();
    expect(box()).not.toBeNull();
    expect(box()!.checked).toBe(false);
    expect(captureBoardShot).not.toHaveBeenCalled();
  });

  it("is not offered with no board to photograph, or with nowhere to store it", async () => {
    mountForm({ captureTarget: undefined });
    await settle();
    expect(box()).toBeNull();
    cleanup();
    resetDraft();
    mountForm({ attachments: false });
    await settle();
    expect(box()).toBeNull();
  });

  it("ticking takes the picture of the table and shows it among the screenshots", async () => {
    captureBoardShot.mockResolvedValue(shotFile());
    mountForm();
    await settle();

    click(box()!);
    await settle();

    expect(captureBoardShot).toHaveBeenCalledTimes(1);
    expect(captureBoardShot).toHaveBeenCalledWith(document.body);
    expect(box()!.checked).toBe(true);
    expect(thumbs()).toEqual(["board-20261019-123456.png"]);
  });

  it("unticking takes the picture back out", async () => {
    captureBoardShot.mockResolvedValue(shotFile());
    mountForm();
    await settle();
    click(box()!);
    await settle();

    click(box()!);
    await settle();

    expect(box()!.checked).toBe(false);
    expect(thumbs()).toEqual([]);
  });

  it("removing its thumbnail unticks the box", async () => {
    captureBoardShot.mockResolvedValue(shotFile());
    mountForm();
    await settle();
    click(box()!);
    await settle();

    click(document.querySelector<HTMLButtonElement>(".shots .remove")!);
    await settle();

    expect(box()!.checked).toBe(false);
    expect(thumbs()).toEqual([]);
  });

  it("says why when the picture cannot be made, and leaves the box unticked", async () => {
    captureBoardShot.mockRejectedValue(new BoardShotError("couldn't take a picture (boom)"));
    mountForm();
    await settle();

    click(box()!);
    await settle();

    expect(document.querySelector('[role="alert"]')?.textContent).toContain("boom");
    expect(box()!.checked).toBe(false);
    expect(thumbs()).toEqual([]);
  });

  it("keeps a screenshot the reporter pasted alongside it", async () => {
    captureBoardShot.mockResolvedValue(shotFile());
    mountForm();
    await settle();
    const input = document.querySelector<HTMLInputElement>('input[type="file"]')!;
    Object.defineProperty(input, "files", {
      value: [new File(["x"], "mine.png", { type: "image/png" })],
      configurable: true,
    });
    input.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();

    click(box()!);
    await settle();

    expect(thumbs().sort()).toEqual(["board-20261019-123456.png", "mine.png"]);
  });

  it("refuses to take a fifth image, and says so", async () => {
    captureBoardShot.mockResolvedValue(shotFile());
    mountForm();
    await settle();
    const input = document.querySelector<HTMLInputElement>('input[type="file"]')!;
    const four = ["a", "b", "c", "d"].map(
      (n) => new File(["x"], `${n}.png`, { type: "image/png" }),
    );
    Object.defineProperty(input, "files", { value: four, configurable: true });
    input.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();

    click(box()!);
    await settle();

    expect(captureBoardShot).not.toHaveBeenCalled();
    expect(document.querySelector('[role="alert"]')?.textContent).toMatch(/at most 4/);
    expect(box()!.checked).toBe(false);
  });
});
