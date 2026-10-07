// @vitest-environment jsdom
//
// settingsNarrow.render.test.ts — #2519. At phone width the Settings
// modal's 180px tab column becomes a scrolling strip above the content.
// The collapse itself is CSS (jsdom has no layout), so this pins what a
// test can: the nav stays real buttons in DOM order, the active tab is
// announced with aria-current and follows a click, and the stylesheet
// carries the narrow rule while the wide rule is untouched.

import { readFileSync } from "node:fs";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import Settings from "./components/Settings.svelte";
import { L } from "./labels";
import { defaultSettings, openSettings, settings } from "./settings";
import { _resetForTests as resetModals } from "./modalLayers";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const TABS = [
  "Audio",
  "Animations",
  "Display",
  "Playmat",
  "Gameplay",
  "Shortcuts",
  "Accessibility",
  "Advanced",
];

beforeEach(() => {
  settings.set(defaultSettings());
  resetModals();
});
afterEach(() => cleanup());

const navButtons = () => [
  ...document.querySelectorAll<HTMLButtonElement>(`nav[aria-label="${L.settingsSections}"] button`),
];

describe("Settings tab navigation", () => {
  it("is one button per tab, in order, with exactly one announced as current", () => {
    render(Settings as never, {} as never);
    openSettings("audio");
    flushSync();
    expect(navButtons().map((b) => b.textContent?.trim())).toEqual(TABS);
    const current = navButtons().filter((b) => b.getAttribute("aria-current") === "page");
    expect(current.map((b) => b.textContent?.trim())).toEqual(["Audio"]);
  });

  it("switching tabs moves the current marker and the content", () => {
    render(Settings as never, {} as never);
    openSettings("audio");
    flushSync();
    const display = navButtons().find((b) => b.textContent?.trim() === "Display")!;
    click(display);
    flushSync();
    expect(display.getAttribute("aria-current")).toBe("page");
    expect(navButtons().filter((b) => b.hasAttribute("aria-current"))).toHaveLength(1);
    expect(document.querySelector("section h3")?.textContent).toBe("Display");
  });
});

describe("Settings stylesheet", () => {
  const src = readFileSync("src/lib/components/Settings.svelte", "utf8");
  const style = src.slice(src.indexOf("<style>"));

  it("keeps the wide 180px column outside any media query", () => {
    const wide = style.slice(0, style.indexOf("@media (max-width: 599px)"));
    expect(wide).toMatch(/nav \{[^}]*flex-direction: column;[^}]*width: 180px;/s);
  });

  it("turns the nav into a horizontally scrolling strip below 600px", () => {
    const narrow = style.slice(style.indexOf("@media (max-width: 599px)"));
    expect(narrow).toMatch(/\.body \{\s*flex-direction: column;/);
    expect(narrow).toMatch(/nav \{[^}]*flex-direction: row;[^}]*overflow-x: auto;/s);
  });
});
