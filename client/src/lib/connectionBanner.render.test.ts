// @vitest-environment jsdom
//
// #519, the rendered half: the board admitting it is stale.
//
// The unit tests next door pin the copy. This pins the thing the copy
// was useless without — that it actually reaches the DOM, in each of
// the three states a player meets, with a live region a screen reader
// will read and a retry a mouse can press.

import { describe, it, expect, afterEach, vi } from "vitest";

import ConnectionBanner from "./components/ConnectionBanner.svelte";
import { STALE_BOARD_SENTENCE } from "./connectionBanner";
import { click, cleanup, render } from "./test/render.svelte";

afterEach(cleanup);

function mount(props: Record<string, unknown>) {
  return render(ConnectionBanner as never, props);
}

function banner(container: HTMLElement): HTMLElement | null {
  return container.querySelector(".connection-banner");
}

function liveRegion(container: HTMLElement): HTMLElement | null {
  return container.querySelector('[role="status"][aria-live="polite"]');
}

describe("the connection banner", () => {
  it("is absent on a live connection", () => {
    const { container } = mount({ status: "connected" });
    expect(banner(container)).toBeNull();
  });

  // The opening dial of a fresh mount has no snapshot to be stale
  // about; a banner here would fire on every normal page load.
  it("is absent on the opening dial", () => {
    const { container } = mount({ status: "connecting" });
    expect(banner(container)).toBeNull();
  });

  it("appears while reconnecting, and says what the board is", () => {
    const { container } = mount({ status: "reconnecting", attempt: 2 });
    const el = banner(container);
    expect(el).not.toBeNull();
    expect(el?.textContent).toContain("Connection lost");
    expect(el?.textContent).toContain(STALE_BOARD_SENTENCE);
  });

  it("carries the attempt count", () => {
    const { container } = mount({ status: "reconnecting", attempt: 4 });
    expect(banner(container)?.textContent).toContain("attempt 4");
  });

  it("appears when the client has given up", () => {
    const { container } = mount({ status: "disconnected" });
    expect(banner(container)?.textContent).toContain("Disconnected");
  });

  it("distinguishes the two stale states by tone", () => {
    const { container, setProps } = mount({ status: "reconnecting", attempt: 1 });
    expect(banner(container)?.classList.contains("tone-retrying")).toBe(true);
    setProps({ status: "disconnected" });
    expect(banner(container)?.classList.contains("tone-lost")).toBe(true);
  });

  // The whole point: it goes away on its own when the table comes
  // back, without anything having to blank the board to prove it.
  it("clears when the connection returns", () => {
    const { container, setProps } = mount({ status: "reconnecting", attempt: 3 });
    expect(banner(container)).not.toBeNull();
    setProps({ status: "connected", attempt: 0 });
    expect(banner(container)).toBeNull();
  });
});

describe("the retry control", () => {
  it("calls back when pressed", () => {
    const onRetry = vi.fn();
    const { container } = mount({ status: "reconnecting", attempt: 1, onRetry });
    const button = container.querySelector("button.retry");
    expect(button).not.toBeNull();
    click(button as Element);
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("is omitted when no handler was given", () => {
    const { container } = mount({ status: "disconnected" });
    expect(container.querySelector("button.retry")).toBeNull();
  });
});

describe("the live region", () => {
  // Mounted outside the {#if} on purpose. A region that appears at the
  // same moment its content does may not be announced at all, and
  // recovery — which the banner signals by silently vanishing — has to
  // be announced too.
  it("exists even on a healthy connection", () => {
    const { container } = mount({ status: "connected" });
    expect(liveRegion(container)).not.toBeNull();
  });

  it("announces each of the three states distinctly", () => {
    const { container, setProps } = mount({ status: "connected" });
    const seen = new Set<string>();
    for (const status of ["connected", "reconnecting", "disconnected"] as const) {
      setProps({ status, attempt: 1 });
      seen.add(liveRegion(container)?.textContent?.trim() ?? "");
    }
    expect(seen.size).toBe(3);
  });

  it("says the table came back", () => {
    const { container, setProps } = mount({ status: "reconnecting", attempt: 1 });
    setProps({ status: "connected", attempt: 0 });
    expect(liveRegion(container)?.textContent).toContain("Reconnected");
  });

  // Exactly one live region. The banner itself carries no live role,
  // so a state change is never read twice.
  it("is the only live region the component renders", () => {
    const { container } = mount({ status: "disconnected" });
    expect(container.querySelectorAll("[aria-live]").length).toBe(1);
  });
});
