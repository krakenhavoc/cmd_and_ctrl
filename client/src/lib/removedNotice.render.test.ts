// @vitest-environment jsdom
//
// #2785: the banner a removed player reads, instead of a silent
// disconnect. It must appear when the store is set, carry the way out
// on the table, and go away when dismissed.

import { afterEach, describe, expect, it } from "vitest";

import RemovedNotice from "./components/RemovedNotice.svelte";
import { REMOVED_NOTICE, removedNotice } from "./session";
import { click, cleanup, render } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  removedNotice.set("");
});

describe("the removed-from-table notice", () => {
  it("is absent until the client learns its seat is gone", () => {
    const { container } = render(RemovedNotice as never, {});
    expect(container.querySelector(".removed-notice")).toBeNull();
  });

  it("says what happened, as an alert", () => {
    removedNotice.set(REMOVED_NOTICE);
    const { container } = render(RemovedNotice as never, {});
    const el = container.querySelector('.removed-notice[role="alert"]');
    expect(el?.textContent).toContain("You were removed from this table");
  });

  it("offers the way back to the lobby only on the table", () => {
    removedNotice.set(REMOVED_NOTICE);
    const lobby = render(RemovedNotice as never, {});
    expect(lobby.container.querySelector("a.go")).toBeNull();
    const table = render(RemovedNotice as never, { withLink: true });
    expect(table.container.querySelector("a.go")?.getAttribute("href")).toBe("#/lobby");
  });

  it("goes away when dismissed", () => {
    removedNotice.set(REMOVED_NOTICE);
    const { container } = render(RemovedNotice as never, {});
    click(container.querySelector("button.dismiss") as HTMLElement);
    expect(container.querySelector(".removed-notice")).toBeNull();
  });
});
