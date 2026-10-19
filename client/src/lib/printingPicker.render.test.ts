// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";
import PrintingPicker from "./components/decks/PrintingPicker.svelte";
import { render, click, cleanup } from "./test/render.svelte";
import { L } from "./labels";
import {
  fetchDeckPrintings,
  fetchCardPrintings,
  saveDeckPrintings,
  type CardPrinting,
} from "./api";

vi.mock("./api", () => ({
  fetchDeckPrintings: vi.fn(),
  fetchCardPrintings: vi.fn(),
  saveDeckPrintings: vi.fn(),
}));
const first: CardPrinting = { id: "first", name: "Sol Ring", set: "c21", collector_number: "1" };
const second: CardPrinting = {
  id: "second",
  name: "Sol Ring",
  flavor_name: "Alternate Ring",
  set: "sld",
  collector_number: "2",
};
const deck = { id: "deck", name: "Squall", commanders: [], card_count: 100, updated_at: "" };

beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(fetchDeckPrintings).mockResolvedValue({
    revision: "source-version",
    cards: [
      {
        entry: 4,
        name: "Sol Ring",
        count: 1,
        board: "Mainboard",
        printing_id: first.id,
        set: first.set,
        collector_number: first.collector_number,
      },
    ],
  });
  vi.mocked(fetchCardPrintings).mockResolvedValue([first, second]);
  vi.mocked(saveDeckPrintings).mockResolvedValue(deck);
});
afterEach(cleanup);
async function ready(): Promise<void> {
  await vi.waitFor(() => {
    flushSync();
    expect(document.querySelectorAll(".version")).toHaveLength(2);
  });
}
function button(container: HTMLElement, text: string): HTMLButtonElement {
  const found = Array.from(container.querySelectorAll("button")).find((b) =>
    b.textContent?.includes(text),
  );
  expect(found).toBeDefined();
  return found!;
}

describe("saved-deck printing picker", () => {
  it("previews and saves only after explicit save", async () => {
    const onsaved = vi.fn();
    const { container } = render(PrintingPicker, { deck, onclose: vi.fn(), onsaved });
    await ready();
    expect(container.querySelectorAll("img")).toHaveLength(2);
    click(button(container, "Alternate Ring"));
    expect(button(container, "Alternate Ring").getAttribute("aria-pressed")).toBe("true");
    expect(saveDeckPrintings).not.toHaveBeenCalled();
    click(button(container, L.savePrintings));
    await vi.waitFor(() => expect(onsaved).toHaveBeenCalledWith(deck));
    expect(saveDeckPrintings).toHaveBeenCalledWith("deck", "source-version", [
      { entry: 4, printing_id: "second" },
    ]);
  });
  it("cancels without writing the draft", async () => {
    const onclose = vi.fn();
    const { container } = render(PrintingPicker, { deck, onclose, onsaved: vi.fn() });
    await ready();
    click(button(container, "Alternate Ring"));
    click(button(container, L.cancelPrintings));
    expect(onclose).toHaveBeenCalledOnce();
    expect(saveDeckPrintings).not.toHaveBeenCalled();
  });
  it("keeps choices when a stale save fails", async () => {
    vi.mocked(saveDeckPrintings).mockRejectedValue(
      new Error("This deck changed. Reopen the printing picker and try again."),
    );
    const onsaved = vi.fn();
    const { container } = render(PrintingPicker, { deck, onclose: vi.fn(), onsaved });
    await ready();
    click(button(container, "Alternate Ring"));
    click(button(container, L.savePrintings));
    await vi.waitFor(() => {
      flushSync();
      expect(container.querySelector("[role=alert]")?.textContent).toContain("Reopen");
    });
    expect(onsaved).not.toHaveBeenCalled();
    expect(button(container, L.savePrintings).disabled).toBe(false);
  });
  it("ignores late options for the previously selected card", async () => {
    vi.mocked(fetchDeckPrintings).mockResolvedValue({
      revision: "r",
      cards: [
        {
          entry: 0,
          name: "Sol Ring",
          count: 1,
          board: "Mainboard",
          printing_id: "first",
          set: "c21",
          collector_number: "1",
        },
        {
          entry: 1,
          name: "Plains",
          count: 10,
          board: "Mainboard",
          printing_id: "plains",
          set: "one",
          collector_number: "1",
        },
      ],
    });
    let finish: (value: CardPrinting[]) => void = () => {};
    vi.mocked(fetchCardPrintings).mockImplementation((id) =>
      id === "first"
        ? new Promise((resolve) => {
            finish = resolve;
          })
        : Promise.resolve([{ ...first, id: "plains", name: "Plains" }]),
    );
    const { container } = render(PrintingPicker, { deck, onclose: vi.fn(), onsaved: vi.fn() });
    await vi.waitFor(() => expect(fetchCardPrintings).toHaveBeenCalledWith("first"));
    flushSync();
    const select = container.querySelector("select")!;
    select.value = "1";
    select.dispatchEvent(new Event("change", { bubbles: true }));
    await vi.waitFor(() => {
      flushSync();
      expect(container.querySelector(".version-name")?.textContent).toBe("Plains");
    });
    finish([first, second]);
    await Promise.resolve();
    flushSync();
    expect(container.querySelector(".version-name")?.textContent).toBe("Plains");
  });
});
