// @vitest-environment jsdom
//
// modelKeySettings.render.test.ts — ADR 0182: Settings › AI opponents.
// The saved key is described by provider, model and last four only; the
// key field is a password field that starts empty; a custom provider
// asks for its address; and the form says CMD & CTRL.

import { afterEach, describe, expect, it } from "vitest";

import ModelKeySettings from "./components/ModelKeySettings.svelte";
import { DEFAULT_LIMITS, MODEL_KEY_OFF, modelKeyState, type ModelKeyState } from "./modelKey";
import { cleanup, render } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  modelKeyState.set(MODEL_KEY_OFF);
});

const state = (key: ModelKeyState["key"]): ModelKeyState => ({
  available: true,
  key,
  providers: [
    { id: "anthropic", label: "Anthropic", transport: "anthropic", model_hint: "claude-haiku-4-5" },
    { id: "custom", label: "Other (OpenAI-compatible)", transport: "openai" },
  ],
  limits: DEFAULT_LIMITS,
});

describe("ModelKeySettings", () => {
  it("offers to add a key, with an empty password field", () => {
    modelKeyState.set(state(null));
    const { container } = render(ModelKeySettings, {});
    expect(container.textContent).toContain("CMD & CTRL");
    const key = container.querySelector<HTMLInputElement>('input[type="password"]');
    expect(key).not.toBeNull();
    expect(key?.value).toBe("");
    expect(container.textContent).toContain("Add key");
    expect(container.querySelector('input[type="url"]')).toBeNull();
  });

  it("describes a saved key without the key and shows its last error", () => {
    modelKeyState.set(
      state({
        provider: "anthropic",
        provider_label: "Anthropic",
        model: "claude-haiku-4-5",
        last4: "cdef",
        games_per_day: 3,
        calls_per_game: 40,
        games_today: 1,
        created_at: "2026-10-14T00:00:00Z",
        updated_at: "2026-10-14T00:00:00Z",
        last_error: "The provider refused the key. Check it, or paste a new one.",
      }),
    );
    const { container } = render(ModelKeySettings, {});
    expect(container.textContent).toContain("Anthropic · claude-haiku-4-5 · key ending cdef");
    expect(container.textContent).toContain("1 of");
    expect(container.textContent).toContain("The provider refused the key.");
    expect(container.textContent).toContain("Test key");
    expect(container.textContent).toContain("Remove key");
    const key = container.querySelector<HTMLInputElement>('input[type="password"]');
    expect(key?.value).toBe("");
    expect(key?.placeholder).toContain("cdef");
  });

  it("asks a custom provider for its address", () => {
    modelKeyState.set(
      state({
        provider: "custom",
        provider_label: "Other (OpenAI-compatible)",
        base_url: "https://api.example.com/v1",
        model: "m",
        last4: "abcd",
        games_per_day: 3,
        calls_per_game: 40,
        games_today: 0,
        created_at: "2026-10-14T00:00:00Z",
        updated_at: "2026-10-14T00:00:00Z",
      }),
    );
    const { container } = render(ModelKeySettings, {});
    const url = container.querySelector<HTMLInputElement>('input[type="url"]');
    expect(url?.value).toBe("https://api.example.com/v1");
  });
});
