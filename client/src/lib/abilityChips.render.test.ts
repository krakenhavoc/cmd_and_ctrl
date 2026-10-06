// @vitest-environment jsdom
//
// #2219: an art tile shows one chip per kind of non-keyword ability the
// server listed in `ability_rows` — ⚡ triggered, ◆ static, ↻ activated —
// with a count, after the keyword chips. Hovering or focusing a chip
// lists that kind's labels; a press pins the list. A full card, an
// uncatalogued card and a card with no rows show no chips.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import { abilityChips, abilityChipDescription } from "./abilityChips";
import { L } from "./labels";
import { updateSettings, defaultSettings } from "./settings";
import type { CardView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const ID = "11111111-1111-1111-1111-111111111111";
const titan = (extra: Partial<CardView> = {}): CardView => ({
  instance_id: "titan",
  name: "Inferno Titan",
  owner: "me",
  controller: "me",
  type_line: "Creature — Giant",
  power: 6,
  toughness: 6,
  scryfall_id: ID,
  known_by_you: true,
  abilities: ["flying"],
  ability_rows: [
    { kind: "activated", label: "{R}: Inferno Titan gets +1/+0 until end of turn." },
    { kind: "triggered", label: "3 damage divided among one, two, or three targets" },
    { kind: "triggered", label: "Whenever it attacks, 3 damage divided" },
    { kind: "static", label: "Power/toughness effect" },
  ],
  ...extra,
});

function setArt(battlefield: boolean) {
  updateSettings("display", "battlefieldArt", battlefield);
  flushSync();
}

function row(cards: CardView[]) {
  return render(
    BattlefieldRow as never,
    {
      label: "creatures",
      cards,
      viewerID: "me",
      attachmentsByHost: {},
      onCardClick: () => {},
    } as never,
  );
}

const chipButtons = (c: HTMLElement) => [...c.querySelectorAll<HTMLButtonElement>(".ability-chip")];
const list = () => document.body.querySelector<HTMLElement>(".ability-list");

// jsdom has no PointerEvent: a MouseEvent carrying the pointer type the
// chip reads stands in for a mouse resting on it.
function mouse(el: Element, type: "pointerenter" | "pointerleave"): void {
  const ev = new MouseEvent(type);
  Object.defineProperty(ev, "pointerType", { value: "mouse" });
  el.dispatchEvent(ev);
  flushSync();
}

beforeEach(() => {
  localStorage.clear();
});
afterEach(() => {
  setArt(defaultSettings().display.battlefieldArt);
  cleanup();
});

describe("abilityChips", () => {
  it("groups rows by kind in chip order and drops kinds with none", () => {
    const chips = abilityChips(titan().ability_rows);
    expect(chips.map((c) => [c.kind, c.count])).toEqual([
      ["triggered", 2],
      ["static", 1],
      ["activated", 1],
    ]);
    expect(abilityChipDescription(chips[0])).toBe(
      "3 damage divided among one, two, or three targets; Whenever it attacks, 3 damage divided",
    );
    expect(abilityChips(undefined)).toEqual([]);
    expect(abilityChips([{ kind: "mystery" as never, label: "x" }])).toEqual([]);
  });

  it("names each chip by count and kind", () => {
    expect(L.abilityChip(2, "triggered")).toBe("2 triggered abilities");
    expect(L.abilityChip(1, "static")).toBe("1 static ability");
  });
});

describe("ability chips on an art tile (#2219)", () => {
  it("draws one chip per kind, with its count, after the keyword chips", () => {
    setArt(true);
    const { container } = row([titan()]);
    const chips = chipButtons(container);
    expect(chips.map((b) => b.getAttribute("aria-label"))).toEqual([
      "2 triggered abilities",
      "1 static ability",
      "1 activated ability",
    ]);
    expect(chips.map((b) => b.textContent?.trim())).toEqual(["2", "1", "1"]);
    // In the keyword row, after the keyword badge.
    const kwRow = container.querySelector(".keyword-row");
    expect(kwRow).not.toBeNull();
    const kids = [...(kwRow as HTMLElement).children];
    const flying = kids.findIndex((k) => k.getAttribute("aria-label") === "Flying");
    const firstChip = kids.findIndex((k) => k.classList.contains("ability-chip"));
    expect(flying).toBeGreaterThanOrEqual(0);
    expect(firstChip).toBeGreaterThan(flying);
    // A screen reader hears every label without opening anything.
    const desc = document.getElementById(chips[0].getAttribute("aria-describedby") ?? "");
    expect(desc?.textContent).toContain("3 damage divided among one, two, or three targets");
  });

  it("lists a kind's labels on hover and on keyboard focus, and closes after", () => {
    setArt(true);
    const { container } = row([titan()]);
    const [trig, stat] = chipButtons(container);
    expect(list()).toBeNull();

    mouse(trig, "pointerenter");
    expect(list()?.dataset.abilityList).toBe("triggered");
    expect([...(list()?.querySelectorAll("li") ?? [])].map((li) => li.textContent)).toEqual([
      "3 damage divided among one, two, or three targets",
      "Whenever it attacks, 3 damage divided",
    ]);
    expect(trig.getAttribute("aria-expanded")).toBe("true");
    mouse(trig, "pointerleave");
    expect(list()).toBeNull();

    // jsdom has no :focus-visible, so a focus counts as keyboard focus.
    stat.focus();
    flushSync();
    expect(list()?.dataset.abilityList).toBe("static");
    expect(list()?.textContent).toContain("Power/toughness effect");
    stat.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(list()).toBeNull();
    stat.blur();
    flushSync();
  });

  it("pins the list on a press without pressing the card, and unpins on a second", () => {
    setArt(true);
    let cardClicks = 0;
    const { container } = render(
      BattlefieldRow as never,
      {
        label: "creatures",
        cards: [titan()],
        viewerID: "me",
        attachmentsByHost: {},
        onCardClick: () => {
          cardClicks++;
        },
      } as never,
    );
    const act = chipButtons(container)[2];
    act.click();
    flushSync();
    expect(list()?.dataset.abilityList).toBe("activated");
    expect(cardClicks).toBe(0);
    act.click();
    flushSync();
    expect(list()).toBeNull();
    expect(cardClicks).toBe(0);
  });

  it("shows no chips on a full card, or for a card with no rows", () => {
    setArt(false);
    let r = row([titan()]);
    expect(chipButtons(r.container)).toHaveLength(0);
    cleanup();

    setArt(true);
    // An uncatalogued card: the server sends no rows (it keeps its
    // `unimplemented` mark for the hover panel and the stack).
    r = row([titan({ ability_rows: undefined, unimplemented: true })]);
    expect(chipButtons(r.container)).toHaveLength(0);
    // The keyword chip is still there.
    expect(r.container.querySelector('[aria-label="Flying"]')).not.toBeNull();
  });
});
