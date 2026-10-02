// @vitest-environment jsdom
//
// ADR 0110 §4 owner answer 6, the rendered half: after a sign-in
// applied the account's settings over different ones, the toast says
// so and offers this browser's back. The sync logic is mocked; its own
// tests are in settingsSync.test.ts.

import { afterEach, describe, expect, it, vi } from "vitest";

const { toast, keepBrowserSettings, dismissSettingsSyncToast } = await vi.hoisted(async () => {
  const { writable } = await import("svelte/store");
  return {
    toast: writable<{ browser: Record<string, Record<string, unknown>> } | null>(null),
    keepBrowserSettings: vi.fn(),
    dismissSettingsSyncToast: vi.fn(),
  };
});

vi.mock("./settingsSync", () => ({
  settingsSyncToast: toast,
  keepBrowserSettings: () => keepBrowserSettings(),
  dismissSettingsSyncToast: () => dismissSettingsSyncToast(),
}));

import SettingsSyncToast from "./components/SettingsSyncToast.svelte";
import { click, cleanup, render } from "./test/render.svelte";
import { flushSync } from "svelte";

afterEach(() => {
  cleanup();
  toast.set(null);
  vi.clearAllMocks();
});

describe("the account settings toast", () => {
  it("is absent until a sign-in applied different settings", () => {
    const { container } = render(SettingsSyncToast as never, {});
    expect(container.querySelector('[role="status"]')).toBeNull();
  });

  it("says the account's settings are in use, and keeps this browser's on request", () => {
    const { container } = render(SettingsSyncToast as never, {});
    toast.set({ browser: { display: { theme: "light" } } });
    flushSync();
    const el = container.querySelector('[role="status"]');
    expect(el?.textContent).toContain("Using your account's settings");
    const keep = Array.from(container.querySelectorAll("button")).find((b) =>
      b.textContent?.includes("Keep this browser's instead"),
    );
    expect(keep).toBeDefined();
    click(keep!);
    expect(keepBrowserSettings).toHaveBeenCalledOnce();

    const dismiss = container.querySelector<HTMLButtonElement>(
      'button[aria-label="keep the account\'s settings"]',
    );
    click(dismiss!);
    expect(dismissSettingsSyncToast).toHaveBeenCalledOnce();
  });
});
