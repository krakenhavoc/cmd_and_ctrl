// @vitest-environment jsdom
//
// ADR 0159 §6: the voice room panel in the table settings modal.

import { afterEach, describe, expect, it, vi } from "vitest";

import VoiceRoom from "./components/VoiceRoom.svelte";
import type { VoiceView } from "./protocol";
import { cleanup, render } from "./test/render.svelte";

const URL = "https://discord.com/channels/1/2";
const INVITE = "https://discord.gg/abc123";

function mount(voice: VoiceView, over: Record<string, unknown> = {}) {
  const ontoggle = vi.fn();
  const onrefresh = vi.fn();
  const { container } = render(
    VoiceRoom as never,
    {
      voice,
      canManage: false,
      ended: false,
      inviteURL: INVITE,
      ontoggle,
      onrefresh,
      ...over,
    } as never,
  );
  return { container, ontoggle, onrefresh };
}

const links = (c: HTMLElement) => [...c.querySelectorAll("a")].map((a) => a.getAttribute("href"));
const buttons = (c: HTMLElement) =>
  [...c.querySelectorAll("button")].map((b) => b.textContent?.trim());

describe("VoiceRoom", () => {
  afterEach(() => cleanup());

  it("gives a granted player the room's link", () => {
    const { container } = mount({ enabled: true, status: "ready", you: "granted", url: URL });
    expect(links(container)).toEqual([URL]);
    expect(container.querySelector("input[type=checkbox]")).toBeNull();
  });

  it("sends a player outside the server to the invite, and lets them check again", () => {
    const { container, onrefresh } = mount({
      enabled: true,
      status: "ready",
      you: "not_in_server",
    });
    expect(links(container)).toEqual([INVITE]);
    const check = [...container.querySelectorAll("button")].find((b) =>
      b.textContent?.includes("check again"),
    );
    check?.click();
    expect(onrefresh).toHaveBeenCalledTimes(1);
  });

  it("offers a reopen for an idle room, but not to a guest", () => {
    expect(buttons(mount({ enabled: true, status: "idle", you: "granted" }).container)).toContain(
      "Reopen voice room",
    );
    cleanup();
    expect(
      buttons(mount({ enabled: true, status: "idle", you: "no_discord" }).container),
    ).not.toContain("Reopen voice room");
  });

  it("gives the host the switch, and an ended table can only turn it off", () => {
    const { container, ontoggle } = mount(
      { enabled: false, status: "off", you: "granted" },
      { canManage: true },
    );
    const box = container.querySelector("input[type=checkbox]") as HTMLInputElement;
    expect(box.checked).toBe(false);
    box.click();
    expect(ontoggle).toHaveBeenCalledWith(true);
    cleanup();

    const ended = mount(
      { enabled: false, status: "off" },
      { canManage: true, ended: true },
    ).container;
    expect((ended.querySelector("input[type=checkbox]") as HTMLInputElement).disabled).toBe(true);
    cleanup();
    const endedOn = mount(
      { enabled: true, status: "ready" },
      { canManage: true, ended: true },
    ).container;
    expect((endedOn.querySelector("input[type=checkbox]") as HTMLInputElement).disabled).toBe(
      false,
    );
  });

  it("shows the server's error", () => {
    const { container } = mount(
      { enabled: true, status: "ready", you: "granted" },
      { error: "nope" },
    );
    expect(container.querySelector("[role=alert]")?.textContent).toBe("nope");
  });
});
