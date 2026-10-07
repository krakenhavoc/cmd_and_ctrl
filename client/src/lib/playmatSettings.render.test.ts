// @vitest-environment jsdom
//
// playmatSettings.render.test.ts — ADR 0128 §11. The Settings "Playmat"
// tab: the per-device choice for everyone, and for an account up to
// three slot cards (Use / Stop using, Replace, Fit to best size,
// Remove), the add / replace panel, the best-size prompt (accept, keep,
// drag, keyboard) and the full-slots replace flow.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

const api = vi.hoisted(() => ({
  fetchMyPlaymats: vi.fn(),
  uploadMyPlaymat: vi.fn(),
  linkMyPlaymat: vi.fn(),
  removeMyPlaymat: vi.fn(),
  setMyPlaymatWash: vi.fn(),
  activateMyPlaymat: vi.fn(),
  fitMyPlaymat: vi.fn(),
}));
vi.mock("./api", async (orig) => ({ ...((await orig()) as object), ...api }));

import PlaymatSettings from "./components/PlaymatSettings.svelte";
import type { MyPlaymats, PlaymatSlot } from "./api";
import { resetSettings } from "./settings";
import { LobbyApiError, session, type Session } from "./session";
import { cleanup, click, render } from "./test/render.svelte";

const URL1 = "/playmats/11111111-1111-4111-8111-111111111111";
const URL2 = "/playmats/22222222-2222-4222-8222-222222222222";
const URL3 = "/playmats/33333333-3333-4333-8333-333333333333";

beforeEach(() => {
  resetSettings();
  session.set(null);
  for (const f of Object.values(api)) f.mockReset();
});

afterEach(() => {
  cleanup();
  session.set(null);
});

const person = (): Session =>
  ({
    token: "tok",
    expiresAt: "2999-01-01T00:00:00Z",
    principal: { role: "identified", user_id: "u1" },
  }) as unknown as Session;

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) await new Promise((r) => setTimeout(r, 0));
  flushSync();
}

// A slot the server calls the right shape, and one it suggests a fit for
// (a 3000 x 2000 image: the crop is 3000 x 1750, centred at y = 125).
const fits = (slot: number, url: string): PlaymatSlot => ({
  slot,
  url,
  width: 2400,
  height: 1400,
  fits: true,
});
const tall = (slot: number, url: string): PlaymatSlot => ({
  slot,
  url,
  width: 3000,
  height: 2000,
  fits: false,
  suggestion: {
    target_width: 2400,
    target_height: 1400,
    crop: { x: 0, y: 125, width: 3000, height: 1750 },
    smaller: false,
  },
});
const account = (slots: PlaymatSlot[], extra: Partial<MyPlaymats> = {}): MyPlaymats => ({
  enabled: true,
  max_slots: 3,
  ideal_width: 2400,
  ideal_height: 1400,
  slots,
  wash: 58,
  ...extra,
});

const radio = (c: HTMLElement, v: string) =>
  c.querySelector<HTMLInputElement>(`input[name="playmats-mode"][value="${v}"]`)!;
const byLabel = (c: ParentNode, label: string) =>
  c.querySelector<HTMLButtonElement>(`button[aria-label="${label}"]`);
const button = (c: ParentNode, text: string) =>
  [...c.querySelectorAll<HTMLButtonElement>("button")].find((b) => b.textContent?.trim() === text);
const card = (c: ParentNode, n: number) =>
  c.querySelector<HTMLElement>(`[aria-label="Playmat ${n}"]`)!;
const text = (el: ParentNode | null) => el?.textContent?.replace(/\s+/g, " ").trim() ?? "";

async function mount(a: MyPlaymats) {
  session.set(person());
  api.fetchMyPlaymats.mockResolvedValue(a);
  const r = render(PlaymatSettings as never, {} as never);
  await settle();
  return r;
}

async function chooseFile(c: HTMLElement, file: File): Promise<void> {
  const input = c.querySelector<HTMLInputElement>('input[type="file"]')!;
  Object.defineProperty(input, "files", { value: [file], configurable: true });
  input.dispatchEvent(new Event("change", { bubbles: true }));
  await settle();
}

const png = () => new File([new Uint8Array(10)], "mat.png", { type: "image/png" });

describe("who sees the account half", () => {
  it("always offers the per-device choice, defaulting to everyone's, and saves it", async () => {
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    expect(radio(r.container, "all").checked).toBe(true);
    click(radio(r.container, "off"));
    const stored = JSON.parse(localStorage.getItem("cmdctrl.settings.v1") ?? "{}");
    expect(stored.display.playmats).toBe("off");
    expect(radio(r.container, "off").checked).toBe(true);
  });

  it("hides the slots from a guest, and does not ask the server", async () => {
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    expect(api.fetchMyPlaymats).not.toHaveBeenCalled();
    expect(r.container.querySelector(".slots")).toBeNull();
    expect(r.container.textContent).toContain("Sign in with Discord");
  });

  it("hides them when the server says 403 (no account to own a playmat)", async () => {
    session.set(person());
    api.fetchMyPlaymats.mockRejectedValue(new LobbyApiError(403, "not signed in as a person"));
    const r = render(PlaymatSettings as never, {} as never);
    await settle();
    expect(r.container.querySelector(".slots")).toBeNull();
    expect(radio(r.container, "all")).not.toBeNull();
  });

  it("hides them when the server has nowhere to store one", async () => {
    const r = await mount({ enabled: false });
    expect(r.container.querySelector(".slots")).toBeNull();
  });
});

describe("the three slot cards", () => {
  it("shows three cards: saved ones with their thumbnail, empty ones with Add a playmat", async () => {
    const r = await mount(account([fits(1, URL1), tall(3, URL3)], { active: 1 }));
    const cards = r.container.querySelectorAll('[role="listitem"]');
    expect(cards).toHaveLength(3);
    expect(card(r.container, 1).querySelector("img")!.getAttribute("src")).toBe(
      `${URL1}?token=tok`,
    );
    expect(card(r.container, 3).querySelector("img")!.getAttribute("src")).toBe(
      `${URL3}?token=tok`,
    );
    expect(card(r.container, 2).querySelector("img")).toBeNull();
    expect(byLabel(card(r.container, 2), "Add a playmat to slot 2")).not.toBeNull();
    expect(text(card(r.container, 1))).toContain("2400×1400");
  });

  it("marks the active one Using, and offers Stop using there and Use elsewhere", async () => {
    const r = await mount(account([fits(1, URL1), fits(2, URL2)], { active: 2 }));
    expect(text(card(r.container, 2))).toContain("Using");
    expect(text(card(r.container, 1))).not.toContain("Using");
    expect(byLabel(card(r.container, 2), "Stop using playmat 2")).not.toBeNull();
    expect(byLabel(card(r.container, 1), "Use playmat 1")).not.toBeNull();
  });

  it("offers Fit to best size only on a slot that is not the best shape", async () => {
    const r = await mount(account([fits(1, URL1), tall(2, URL2)], { active: 1 }));
    expect(byLabel(card(r.container, 1), "Fit playmat 1 to the best size")).toBeNull();
    expect(byLabel(card(r.container, 2), "Fit playmat 2 to the best size")).not.toBeNull();
    expect(text(card(r.container, 2))).toContain("not the best shape");
  });

  it("Use activates that slot, and Stop using shows none", async () => {
    const r = await mount(account([fits(1, URL1), fits(2, URL2)], { active: 1 }));
    api.activateMyPlaymat.mockResolvedValue(account([fits(1, URL1), fits(2, URL2)], { active: 2 }));
    click(byLabel(card(r.container, 2), "Use playmat 2")!);
    await settle();
    expect(api.activateMyPlaymat).toHaveBeenLastCalledWith(2);
    expect(text(card(r.container, 2))).toContain("Using");
    api.activateMyPlaymat.mockResolvedValue(account([fits(1, URL1), fits(2, URL2)]));
    click(byLabel(card(r.container, 2), "Stop using playmat 2")!);
    await settle();
    expect(api.activateMyPlaymat).toHaveBeenLastCalledWith(null);
    expect(r.container.querySelector(".badge")).toBeNull();
  });

  it("Remove asks first, and only the confirmation removes", async () => {
    const r = await mount(account([fits(1, URL1), fits(2, URL2)], { active: 1 }));
    click(byLabel(card(r.container, 2), "Remove playmat 2")!);
    expect(text(card(r.container, 2).querySelector(".confirm"))).toContain("Remove playmat 2?");
    expect(api.removeMyPlaymat).not.toHaveBeenCalled();
    click(button(card(r.container, 2), "Cancel")!);
    expect(card(r.container, 2).querySelector(".confirm")).toBeNull();
    expect(api.removeMyPlaymat).not.toHaveBeenCalled();

    click(byLabel(card(r.container, 2), "Remove playmat 2")!);
    api.removeMyPlaymat.mockResolvedValue(account([fits(1, URL1)], { active: 1 }));
    click(button(card(r.container, 2), "Yes, remove it")!);
    await settle();
    expect(api.removeMyPlaymat).toHaveBeenCalledWith(2);
    expect(card(r.container, 2).querySelector("img")).toBeNull();
  });

  it("warns that removing the one on the table takes it off the table", async () => {
    const r = await mount(account([fits(1, URL1)], { active: 1 }));
    click(byLabel(card(r.container, 1), "Remove playmat 1")!);
    expect(text(card(r.container, 1).querySelector(".confirm"))).toContain("on your table");
  });

  it("shows no broken image when a thumbnail will not load", async () => {
    const r = await mount(account([fits(1, URL1)], { active: 1 }));
    card(r.container, 1).querySelector("img")!.dispatchEvent(new Event("error"));
    flushSync();
    expect(card(r.container, 1).querySelector("img")).toBeNull();
    expect(text(card(r.container, 1))).toContain("Image unavailable");
  });
});

describe("adding, replacing and the best-size prompt", () => {
  it("Add a playmat on an empty slot opens upload-or-link for that slot, and uploads into it", async () => {
    const r = await mount(account([fits(1, URL1)], { active: 1 }));
    click(byLabel(card(r.container, 2), "Add a playmat to slot 2")!);
    expect(
      r.container.querySelector('[aria-label="Add a playmat to slot 2"][role="group"]'),
    ).not.toBeNull();
    api.uploadMyPlaymat.mockResolvedValue(
      account([fits(1, URL1), fits(2, URL2)], { active: 1, slot: 2 }),
    );
    const file = png();
    await chooseFile(r.container, file);
    expect(api.uploadMyPlaymat).toHaveBeenCalledWith(2, file);
    expect(card(r.container, 2).querySelector("img")).not.toBeNull();
    // It fits, so no prompt; the panel closes.
    expect(r.container.querySelector('[aria-label="Fit playmat to the best size"]')).toBeNull();
    expect(
      r.container.querySelector('[role="group"][aria-label^="Add a playmat to slot"]'),
    ).toBeNull();
  });

  it("links into the chosen slot, clears the field, and sends the link once", async () => {
    const r = await mount(account([]));
    click(byLabel(card(r.container, 1), "Add a playmat to slot 1")!);
    api.linkMyPlaymat.mockResolvedValue(account([fits(1, URL1)], { active: 1, slot: 1 }));
    const use = button(r.container, "Use this image")!;
    expect(use.disabled).toBe(true);
    const input = r.container.querySelector<HTMLInputElement>('input[type="url"]')!;
    input.value = " https://example.com/mat.jpg ";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    r.container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(api.linkMyPlaymat).toHaveBeenCalledTimes(1);
    expect(api.linkMyPlaymat).toHaveBeenCalledWith(1, "https://example.com/mat.jpg");
    // The preview is OUR url, not the pasted one.
    expect(r.container.querySelector(".preview img")!.getAttribute("src")).toContain("/playmats/");
    expect(r.container.querySelector(".slots")!.innerHTML).not.toContain("example.com");
  });

  it("shows the server's message when it refuses a link or a file, and keeps what was typed", async () => {
    const r = await mount(account([]));
    click(byLabel(card(r.container, 1), "Add a playmat to slot 1")!);
    api.linkMyPlaymat.mockRejectedValue(
      new LobbyApiError(422, "could not fetch that link: only https links are accepted"),
    );
    const input = r.container.querySelector<HTMLInputElement>('input[type="url"]')!;
    input.value = "http://example.com/a.png";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    r.container
      .querySelector("form")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(r.container.querySelector('[role="alert"]')!.textContent).toContain("only https links");
    expect(input.value).toBe("http://example.com/a.png");

    api.uploadMyPlaymat.mockRejectedValue(
      new LobbyApiError(415, "that file is not a PNG, JPEG or WebP image"),
    );
    await chooseFile(r.container, png());
    expect(r.container.querySelector('[role="alert"]')!.textContent).toContain(
      "not a PNG, JPEG or WebP",
    );
  });

  it("refuses an oversized file before sending it", async () => {
    const r = await mount(account([]));
    click(byLabel(card(r.container, 1), "Add a playmat to slot 1")!);
    const big = new File([new Uint8Array(1)], "big.png", { type: "image/png" });
    Object.defineProperty(big, "size", { value: 11 * 1024 * 1024 });
    await chooseFile(r.container, big);
    expect(api.uploadMyPlaymat).not.toHaveBeenCalled();
    expect(r.container.querySelector('[role="alert"]')!.textContent).toContain("10 MB");
  });

  it("Replace on a saved slot says what is lost, and uploads into that slot", async () => {
    const r = await mount(account([fits(1, URL1), fits(2, URL2)], { active: 1 }));
    click(byLabel(card(r.container, 2), "Replace playmat 2")!);
    expect(
      text(r.container.querySelector('[role="group"][aria-label^="Add a playmat"]')),
    ).toContain("Replace playmat 2. The image it holds now is deleted.");
    api.uploadMyPlaymat.mockResolvedValue(
      account([fits(1, URL1), fits(2, URL3)], { active: 1, slot: 2 }),
    );
    await chooseFile(r.container, png());
    expect(api.uploadMyPlaymat).toHaveBeenCalledWith(2, expect.any(File));
  });

  describe("with all three slots full", () => {
    const full = () => account([fits(1, URL1), fits(2, URL2), fits(3, URL3)], { active: 1 });

    it("Add a playmat asks which one to replace, then replaces that one", async () => {
      const r = await mount(full());
      expect(byLabel(r.container, "Add a playmat to slot 1")).toBeNull(); // no empty card
      click(button(r.container, "Add a playmat")!);
      const chooser = r.container.querySelector(
        '[role="group"][aria-label="Choose a slot to replace"]',
      )!;
      expect(text(chooser)).toContain("All 3 slots are full");
      expect(
        r.container.querySelector('[role="group"][aria-label^="Add a playmat to slot"]'),
      ).toBeNull();
      click(button(chooser, "Replace playmat 3")!);
      expect(
        r.container.querySelector('[role="group"][aria-label="Choose a slot to replace"]'),
      ).toBeNull();
      expect(
        text(r.container.querySelector('[role="group"][aria-label^="Add a playmat"]')),
      ).toContain("Replace playmat 3");
      api.uploadMyPlaymat.mockResolvedValue({ ...full(), slot: 3 });
      await chooseFile(r.container, png());
      expect(api.uploadMyPlaymat).toHaveBeenCalledWith(3, expect.any(File));
      expect(api.uploadMyPlaymat).toHaveBeenCalledTimes(1);
    });

    it("Cancel closes the chooser without sending anything", async () => {
      const r = await mount(full());
      click(button(r.container, "Add a playmat")!);
      click(
        button(r.container.querySelector('[aria-label="Choose a slot to replace"]')!, "Cancel")!,
      );
      expect(r.container.querySelector('[aria-label="Choose a slot to replace"]')).toBeNull();
      expect(api.uploadMyPlaymat).not.toHaveBeenCalled();
    });
  });

  it("Add a playmat with room goes straight to the first free slot", async () => {
    const r = await mount(account([fits(1, URL1), fits(3, URL3)], { active: 1 }));
    click(button(r.container, "Add a playmat")!);
    expect(
      r.container.querySelector('[role="group"][aria-label="Add a playmat to slot 2"]'),
    ).not.toBeNull();
  });
});

describe("the best-size prompt", () => {
  const promptOf = (c: ParentNode) =>
    c.querySelector<HTMLElement>('section[aria-label="Fit playmat to the best size"]');
  const windowOf = (c: ParentNode) => c.querySelector<HTMLElement>('[role="slider"]')!;
  // The browser normalises "6.250%" to "6.25%", so compare as numbers.
  const pct = (v: string) => parseFloat(v);

  // Upload a wrong-shaped image into slot 2 and land on the prompt.
  async function withPrompt() {
    const r = await mount(account([fits(1, URL1)], { active: 1 }));
    click(byLabel(card(r.container, 2), "Add a playmat to slot 2")!);
    api.uploadMyPlaymat.mockResolvedValue(
      account([fits(1, URL1), tall(2, URL2)], { active: 1, slot: 2 }),
    );
    await chooseFile(r.container, png());
    return r;
  }

  it("opens right after saving a wrong-shaped image, with the sizes in the text", async () => {
    const r = await withPrompt();
    const p = promptOf(r.container)!;
    expect(p).not.toBeNull();
    expect(text(p)).toContain(
      "This image is 3000×2000. Playmats look best at 2400×1400 (the shape of a paper playmat).",
    );
    expect(button(p, "Fit to best size")).toBeDefined();
    expect(button(p, "Keep as is")).toBeDefined();
    expect(p.querySelector("img")!.getAttribute("src")).toBe(`${URL2}?token=tok`);
    // The kept area is outlined at the server's centred default.
    const w = windowOf(p);
    expect(pct(w.style.top)).toBeCloseTo(6.25, 2);
    expect(pct(w.style.height)).toBeCloseTo(87.5, 2);
    expect(pct(w.style.left)).toBeCloseTo(0.0, 2);
    expect(pct(w.style.width)).toBeCloseTo(100.0, 2);
    expect(w.getAttribute("aria-label")).toContain("up and down");
  });

  it("does not open for an image that fits", async () => {
    const r = await mount(account([]));
    click(byLabel(card(r.container, 1), "Add a playmat to slot 1")!);
    api.uploadMyPlaymat.mockResolvedValue(account([fits(1, URL1)], { active: 1, slot: 1 }));
    await chooseFile(r.container, png());
    expect(promptOf(r.container)).toBeNull();
  });

  it("Fit to best size sends the centred origin, and closes", async () => {
    const r = await withPrompt();
    api.fitMyPlaymat.mockResolvedValue(
      account([fits(1, URL1), fits(2, URL3)], { active: 1, slot: 2 }),
    );
    click(button(promptOf(r.container)!, "Fit to best size")!);
    await settle();
    expect(api.fitMyPlaymat).toHaveBeenCalledWith(2, 0, 125);
    expect(promptOf(r.container)).toBeNull();
    expect(byLabel(card(r.container, 2), "Fit playmat 2 to the best size")).toBeNull();
  });

  it("Keep as is closes it, leaves the image alone, and the slot offers the fit later", async () => {
    const r = await withPrompt();
    click(button(promptOf(r.container)!, "Keep as is")!);
    expect(promptOf(r.container)).toBeNull();
    expect(api.fitMyPlaymat).not.toHaveBeenCalled();
    const later = byLabel(card(r.container, 2), "Fit playmat 2 to the best size")!;
    expect(later).not.toBeNull();
    click(later);
    expect(promptOf(r.container)).not.toBeNull();
  });

  it("dragging the window moves it along the cropped axis, and the fit uses where it ended", async () => {
    const r = await withPrompt();
    const p = promptOf(r.container)!;
    const frame = p.querySelector<HTMLElement>(".frame")!;
    frame.getBoundingClientRect = () => ({
      width: 300,
      height: 200,
      x: 0,
      y: 0,
      top: 0,
      left: 0,
      right: 300,
      bottom: 200,
      toJSON: () => ({}),
    });
    const w = windowOf(p);
    const ptr = (type: string, x: number, y: number) =>
      w.dispatchEvent(new MouseEvent(type, { bubbles: true, clientX: x, clientY: y }));
    ptr("pointerdown", 50, 50);
    // 10 px of a 200 px frame is 100 px of a 2000 px image: 125 + 100 = 225.
    ptr("pointermove", 80, 60); // the sideways travel is ignored
    ptr("pointerup", 80, 60);
    flushSync();
    expect(pct(w.style.top)).toBeCloseTo(11.25, 2);
    expect(pct(w.style.left)).toBeCloseTo(0.0, 2);
    // It cannot be dragged out of the image: the room is 250 px.
    ptr("pointerdown", 0, 0);
    ptr("pointermove", 0, 500);
    ptr("pointerup", 0, 500);
    flushSync();
    expect(pct(w.style.top)).toBeCloseTo(12.5, 2);
    api.fitMyPlaymat.mockResolvedValue(
      account([fits(1, URL1), fits(2, URL3)], { active: 1, slot: 2 }),
    );
    click(button(p, "Fit to best size")!);
    await settle();
    expect(api.fitMyPlaymat).toHaveBeenCalledWith(2, 0, 250);
  });

  it("arrow keys move the window, Home and End jump, Enter accepts", async () => {
    const r = await withPrompt();
    const w = windowOf(promptOf(r.container)!);
    const key = (k: string, init: KeyboardEventInit = {}) => {
      w.dispatchEvent(
        new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true, ...init }),
      );
      flushSync();
    };
    expect(w.getAttribute("tabindex")).toBe("0");
    // 5% of the 250 px of room is 13 px (rounded).
    key("ArrowDown");
    expect(pct(w.style.top)).toBeCloseTo(((125 + 13) / 2000) * 100, 2);
    key("ArrowUp");
    key("ArrowUp");
    expect(pct(w.style.top)).toBeCloseTo(((125 - 13) / 2000) * 100, 2);
    key("Home");
    expect(pct(w.style.top)).toBeCloseTo(0.0, 2);
    key("End");
    expect(pct(w.style.top)).toBeCloseTo(12.5, 2);
    key("ArrowLeft"); // left and up both step back along the one axis a tall image has
    expect(pct(w.style.left)).toBeCloseTo(0.0, 2);
    api.fitMyPlaymat.mockResolvedValue(
      account([fits(1, URL1), fits(2, URL3)], { active: 1, slot: 2 }),
    );
    key("Enter");
    await settle();
    expect(api.fitMyPlaymat).toHaveBeenCalledWith(2, 0, 237);
  });

  it("Escape keeps the image as it is, from the window and from a button", async () => {
    const r = await withPrompt();
    windowOf(promptOf(r.container)!).dispatchEvent(
      new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true }),
    );
    flushSync();
    expect(promptOf(r.container)).toBeNull();
    expect(api.fitMyPlaymat).not.toHaveBeenCalled();

    click(byLabel(card(r.container, 2), "Fit playmat 2 to the best size")!);
    button(promptOf(r.container)!, "Keep as is")!.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true }),
    );
    flushSync();
    expect(promptOf(r.container)).toBeNull();
    expect(api.fitMyPlaymat).not.toHaveBeenCalled();
  });

  it("warns, and says what it will keep, when the image is too small to reach the best size", async () => {
    const small: PlaymatSlot = {
      slot: 1,
      url: URL1,
      width: 800,
      height: 800,
      fits: false,
      suggestion: {
        target_width: 800,
        target_height: 467,
        crop: { x: 0, y: 166, width: 800, height: 467 },
        smaller: true,
      },
    };
    const r = await mount(account([]));
    click(byLabel(card(r.container, 1), "Add a playmat to slot 1")!);
    api.uploadMyPlaymat.mockResolvedValue(account([small], { active: 1, slot: 1 }));
    await chooseFile(r.container, png());
    const t = text(promptOf(r.container));
    expect(t).toContain("This image is 800×800.");
    expect(t).toContain("800×467");
    expect(t).toContain("may look soft");
  });

  it("shows the server's message when a fit is refused, and stays open", async () => {
    const r = await withPrompt();
    api.fitMyPlaymat.mockRejectedValue(
      new LobbyApiError(400, "that crop does not lie inside the image"),
    );
    click(button(promptOf(r.container)!, "Fit to best size")!);
    await settle();
    expect(r.container.querySelector('[role="alert"]')!.textContent).toContain(
      "does not lie inside",
    );
    expect(promptOf(r.container)).not.toBeNull();
  });
});

// ADR 0128 §10: the owner sets how dark their playmat is, one value for
// the account, whichever slot is in use.
describe("the owner-set wash and the preview of the playmat in use", () => {
  const washOf = (el: HTMLElement | null) => el!.style.getPropertyValue("--playmat-wash");

  it("previews the playmat in use under the wash, and says so when none is", async () => {
    const r = await mount(account([fits(1, URL1), fits(2, URL2)], { active: 2, wash: 70 }));
    expect(r.container.querySelector<HTMLImageElement>(".preview img")!.getAttribute("src")).toBe(
      `${URL2}?token=tok`,
    );
    expect(washOf(r.container.querySelector<HTMLElement>(".preview"))).toBe("70%");
    api.activateMyPlaymat.mockResolvedValue(account([fits(1, URL1), fits(2, URL2)], { wash: 70 }));
    click(byLabel(card(r.container, 2), "Stop using playmat 2")!);
    await settle();
    expect(r.container.querySelector(".preview")).toBeNull();
    expect(r.container.textContent).toContain("No playmat is on your table");
  });

  it("starts from the account's wash and saves once the slider rests", async () => {
    vi.useFakeTimers();
    try {
      session.set(person());
      api.fetchMyPlaymats.mockResolvedValue(account([fits(1, URL1)], { active: 1, wash: 70 }));
      api.setMyPlaymatWash.mockResolvedValue(account([fits(1, URL1)], { active: 1, wash: 45 }));
      const r = render(PlaymatSettings as never, {} as never);
      await vi.advanceTimersByTimeAsync(10);
      flushSync();
      const slider = r.container.querySelector<HTMLInputElement>(
        'input[aria-label="Playmat darkness"]',
      )!;
      expect(slider.value).toBe("70");
      expect(washOf(r.container.querySelector<HTMLElement>(".preview"))).toBe("70%");
      for (const v of ["60", "50", "45"]) {
        slider.value = v;
        slider.dispatchEvent(new Event("input", { bubbles: true }));
      }
      flushSync();
      expect(washOf(r.container.querySelector<HTMLElement>(".preview"))).toBe("45%");
      expect(api.setMyPlaymatWash).not.toHaveBeenCalled();
      await vi.advanceTimersByTimeAsync(450);
      expect(api.setMyPlaymatWash).toHaveBeenCalledTimes(1);
      expect(api.setMyPlaymatWash).toHaveBeenCalledWith(45);
    } finally {
      vi.useRealTimers();
    }
  });
});
