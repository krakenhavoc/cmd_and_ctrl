// @vitest-environment jsdom
//
// autoAnswersSettings.render.test.ts — ADR 0127 §6: Settings → Gameplay
// → Automatic answers. One row per rule, Ask / Always / Never, Forget,
// and what the section says when there is none.

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { flushSync } from "svelte";
import { get } from "svelte/store";

import AutoAnswersSettings from "./components/AutoAnswersSettings.svelte";
import { L } from "./labels";
import { resetSettings, settings, updateSettings } from "./settings";
import { cleanup, click, render } from "./test/render.svelte";

beforeEach(() => resetSettings());
afterEach(() => cleanup());

function seed(): void {
  updateSettings("gameplay", "autoAnswers", [
    { key: "tax", card: "Rhystic Study", prompt: "Rhystic Study — pay {1}?", answer: "never" },
    {
      key: "sphinx",
      card: "Consecrated Sphinx",
      prompt: "Consecrated Sphinx — draw two cards?",
      answer: "always",
    },
  ]);
}

describe("Automatic answers in Settings", () => {
  it("says how to add one when there is none", () => {
    const r = render(AutoAnswersSettings, {});
    const section = r.container.querySelector(`[aria-label="${L.automaticAnswers}"]`);
    expect(section).not.toBeNull();
    expect(section?.textContent).toContain(`Tick ${L.rememberThisAnswer} on a prompt`);
    expect(r.container.querySelectorAll("li")).toHaveLength(0);
  });

  it("lists the rules with their card, question and answer", () => {
    seed();
    const r = render(AutoAnswersSettings, {});
    const rows = r.container.querySelectorAll("li");
    expect(rows).toHaveLength(2);
    expect(rows[0].textContent).toContain("Rhystic Study");
    expect(rows[0].textContent).toContain("pay {1}?");
    expect((rows[0].querySelector("select") as HTMLSelectElement).value).toBe("never");
    expect((rows[1].querySelector("select") as HTMLSelectElement).value).toBe("always");
  });

  it("changes an answer, and Ask removes the rule", () => {
    seed();
    const r = render(AutoAnswersSettings, {});
    const select = r.container.querySelector("li select") as HTMLSelectElement;
    select.value = "always";
    select.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    expect(get(settings).gameplay.autoAnswers.find((x) => x.key === "tax")?.answer).toBe("always");
    const again = r.container.querySelector("li select") as HTMLSelectElement;
    again.value = "ask";
    again.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    expect(get(settings).gameplay.autoAnswers.map((x) => x.key)).toEqual(["sphinx"]);
  });

  it("Forget removes the rule", () => {
    seed();
    const r = render(AutoAnswersSettings, {});
    const forget = [...r.container.querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === L.forgetAutoAnswer,
    );
    expect(forget).toBeDefined();
    click(forget!);
    expect(get(settings).gameplay.autoAnswers.map((x) => x.key)).toEqual(["sphinx"]);
  });
});
