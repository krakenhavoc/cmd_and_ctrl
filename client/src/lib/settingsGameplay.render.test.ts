// @vitest-environment jsdom
//
// settingsGameplay.render.test.ts — ADR 0143 §3. The Gameplay tab is
// five short sections (Passing priority, Reading time, Prompts, Mana,
// Table) and a collapsed Advanced section holding everything a typical
// player never changes.

import { readFileSync } from "node:fs";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import Settings from "./components/Settings.svelte";
import { L } from "./labels";
import { defaultSettings, openSettings, settings } from "./settings";
import { _resetForTests as resetModals } from "./modalLayers";
import { cleanup, flushSync, render } from "./test/render.svelte";
import { setCurrentTablePace } from "./tableSettings";

const ADVANCED_OPEN_KEY = "cmdctrl.settings.gameplayAdvancedOpen";

beforeEach(() => {
  localStorage.clear();
  settings.set(defaultSettings());
  resetModals();
});
afterEach(() => cleanup());

function openGameplay(): HTMLElement {
  render(Settings as never, {} as never);
  openSettings("gameplay");
  flushSync();
  const h3 = [...document.querySelectorAll("h3")].find((h) => h.textContent === "Gameplay");
  if (!h3?.parentElement) throw new Error("no Gameplay tab");
  return h3.parentElement;
}

const advanced = (root: HTMLElement) => {
  const d = root.querySelector<HTMLDetailsElement>("details.gp-advanced");
  if (!d) throw new Error("no Advanced section");
  return d;
};

// The text of every label (and legend) in a part of the page.
const labelsIn = (el: Element) =>
  [...el.querySelectorAll("label, legend")].map((l) => l.textContent?.replace(/\s+/g, " ").trim());

describe("Settings → Gameplay: the sections (ADR 0143 §3.1)", () => {
  it("has the five sections in order, then Advanced", () => {
    const root = openGameplay();
    const headings = [...root.querySelectorAll(".gp-heading")].map((h) => h.textContent?.trim());
    expect(headings).toEqual([
      "Passing priority",
      "Reading time",
      "Prompts",
      "Mana",
      "Table",
      "Advanced",
    ]);
  });

  it("puts each everyday control in its section", () => {
    const root = openGameplay();
    const section = (id: string) => {
      const s = root.querySelector(`section[aria-labelledby="${id}"]`);
      if (!s) throw new Error(`no section ${id}`);
      return s;
    };
    const passing = section("gp-passing");
    expect(passing.querySelectorAll('input[type="radio"][name="pass-mode"]')).toHaveLength(3);
    expect(labelsIn(passing)).toEqual(
      expect.arrayContaining(["Auto-pass", "Smart", "Careful", "Manual", "Stop at these steps"]),
    );
    expect(passing.querySelector("table.step-stops-table")).not.toBeNull();

    // ADR 0143 §2.6: no personal control; the table's pace sets it.
    expect(section("gp-reading").querySelector("select, input")).toBeNull();

    const prompts = section("gp-prompts");
    expect(labelsIn(prompts).some((l) => l?.startsWith("Order my triggers"))).toBe(true);
    expect(prompts.querySelector(`[aria-label="${L.automaticAnswers}"]`)).not.toBeNull();

    expect(labelsIn(section("gp-mana"))).toEqual(["Charge mana costs"]);
    expect(labelsIn(section("gp-table"))).toEqual([
      "Highlight what you can do right now",
      "Ask before leaving a game in progress",
    ]);
  });

  it("uses the ADR's copy for the pass modes", () => {
    const root = openGameplay();
    const passing = root.querySelector('section[aria-labelledby="gp-passing"]')!;
    const text = passing.textContent?.replace(/\s+/g, " ") ?? "";
    expect(text).toContain(
      "Stops where you can do something, and whenever you can respond to an opponent.",
    );
    expect(text).toContain(
      "You always get a chance to respond to an opponent's spell, an attack, and an opponent's end step, ticked or not.",
    );
  });
});

describe("Settings → Gameplay: Reading time (ADR 0143 §2.6, §3.2)", () => {
  afterEach(() => setCurrentTablePace(null));

  const readingText = (root: HTMLElement) =>
    root
      .querySelector('section[aria-labelledby="gp-reading"]')
      ?.textContent?.replace(/\s+/g, " ")
      .trim() ?? "";

  it("names this table's pace and its hold during a game", () => {
    for (const [pace, name, secs] of [
      ["fast", "Fast", "0"],
      ["normal", "Normal", "2"],
      ["slow", "Slow", "3"],
    ] as const) {
      setCurrentTablePace(pace);
      const text = readingText(openGameplay());
      expect(text).toContain(
        `This table's pace is ${name}: other players' spells stay on the stack for ${secs} s before they resolve.`,
      );
      expect(text).toContain("The host sets it in Table settings.");
      expect(text).toContain("Click wait on the countdown to keep priority and respond.");
      cleanup();
      resetModals();
    }
  });

  it("gives all three paces outside a game", () => {
    const text = readingText(openGameplay());
    expect(text).toContain("0 s (Fast), 2 s (Normal) or 3 s (Slow)");
    expect(text).not.toContain("This table's pace");
  });
});

describe("Settings → Gameplay: Advanced (ADR 0143 §3.1)", () => {
  it("starts collapsed", () => {
    const root = openGameplay();
    expect(advanced(root).open).toBe(false);
  });

  it("holds the controls a typical player never changes, and nothing else on the page does", () => {
    const root = openGameplay();
    const adv = labelsIn(advanced(root));
    for (const want of [
      "Skip a ticked step when I have nothing to do there",
      "What counts as a response",
      "Counterspells",
      "Instants and flash",
      "Targeted and protective abilities",
      "Value abilities (Mind Stone, fetch lands, cycling)",
      "Special actions (foretell, suspend, face-up)",
      "Pass my own spells and triggers straight away",
      "Bluffing",
      "Represent a counterspell",
      "Represent an instant",
      "Shortest pause (ms)",
      "Longest pause (ms)",
      "Manual card controls on right-click",
      "Show bot reasoning",
    ]) {
      expect(
        adv.some((l) => l?.startsWith(want)),
        want,
      ).toBe(true);
    }
    // None of those is outside Advanced.
    const outside = [...root.querySelectorAll("section.gp-section")].flatMap(labelsIn);
    for (const l of outside) {
      expect(l).not.toMatch(/Skip a ticked step|Counterspells|Bluffing|right-click|bot reasoning/);
    }
  });

  it("remembers being opened on this device", () => {
    const root = openGameplay();
    const d = advanced(root);
    d.open = true;
    d.dispatchEvent(new Event("toggle"));
    flushSync();
    expect(localStorage.getItem(ADVANCED_OPEN_KEY)).toBe("1");
    cleanup();
    resetModals();
    expect(advanced(openGameplay()).open).toBe(true);
  });

  it("forgets it when closed again", () => {
    localStorage.setItem(ADVANCED_OPEN_KEY, "1");
    const d = advanced(openGameplay());
    expect(d.open).toBe(true);
    d.open = false;
    d.dispatchEvent(new Event("toggle"));
    flushSync();
    expect(localStorage.getItem(ADVANCED_OPEN_KEY)).toBeNull();
  });

  it("disables bluffing outside Smart, and says why", () => {
    settings.update((s) => ({ ...s, gameplay: { ...s.gameplay, passMode: "careful" } }));
    const root = openGameplay();
    const bluff = [...advanced(root).querySelectorAll("fieldset")].find(
      (f) => f.querySelector("legend")?.textContent?.trim() === "Bluffing",
    )!;
    expect(bluff.disabled).toBe(true);
    expect(bluff.textContent).toContain("Bluffing needs Smart auto-pass: in Careful");
  });

  it("never disables the response categories", () => {
    for (const passMode of ["smart", "careful", "manual"] as const) {
      settings.update((s) => ({ ...s, gameplay: { ...s.gameplay, passMode } }));
      const root = openGameplay();
      const responses = [...advanced(root).querySelectorAll("fieldset")].find(
        (f) => f.querySelector("legend")?.textContent?.trim() === "What counts as a response",
      )!;
      expect(responses.disabled, passMode).toBe(false);
      cleanup();
      resetModals();
    }
  });
});

// Tablet is the smallest screen the table is played on. jsdom has no
// layout, so this pins what can be pinned: nothing in the new sections
// or around the stops table asks for a fixed or minimum width, so the
// panel's own width (768 px and up) decides, and the stops table's two
// columns stay side by side (ADR 0143 §3.1).
describe("Settings → Gameplay at tablet width", () => {
  const src = readFileSync("src/lib/components/Settings.svelte", "utf8");
  const style = src.slice(src.indexOf("<style>"));
  const rules = (selector: string) =>
    [...style.matchAll(/([^{}]+)\{([^}]*)\}/g)]
      .filter(([, sel]) => sel.includes(selector) && !sel.includes("input"))
      .map(([, , body]) => body);

  it("sets no fixed or minimum width on the sections or the stops table", () => {
    for (const sel of [
      ".gp-section",
      ".gp-heading",
      ".gp-advanced",
      ".pass-mode-choices",
      ".step-stops-table",
    ]) {
      expect(rules(sel).length, sel).toBeGreaterThan(0);
      for (const body of rules(sel)) {
        expect(body, sel).not.toMatch(/(^|[\s;])(min-)?width\s*:/);
      }
    }
  });

  it("lets the pass-mode choices wrap", () => {
    expect(rules(".pass-mode-choices").join(";")).toMatch(/flex-wrap:\s*wrap/);
  });
});
