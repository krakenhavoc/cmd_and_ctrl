// @vitest-environment jsdom
//
// dockKeys.test.ts — the action dock's ONE Enter / Escape handler (ADR
// 0111 §1, Delivery PR 4). Enter presses the open request's primary and
// Escape its cancel, but only a button that advertises the key; and the
// handler stands down for a step row (so it never presses `next`),
// behind a modal layer, while typing, and for Enter on a focused
// control. dockKeyFor is the whole decision ActionDock's window handler
// makes, so it is pinned here without mounting anything.

import { describe, it, expect, vi, afterEach } from "vitest";

import { dockKeyAction, dockKeyFor, type DockAction, type DockRequest } from "./dock";
import { blockRequest, combatSelectionRequest } from "./combatDock";

afterEach(() => {
  document.body.innerHTML = "";
});

const press = (id: string, keyShortcuts?: string, over: Partial<DockAction> = {}): DockAction => ({
  id,
  label: id,
  keyShortcuts,
  onPress: vi.fn(),
  ...over,
});

const flow = (over: Partial<DockRequest> = {}): DockRequest => ({
  rank: "flow",
  label: "Select target for Arc Lightning",
  primary: press("done", "Enter"),
  secondary: [press("cancel", "Escape")],
  ...over,
});

type KeyInit = Partial<
  Pick<
    KeyboardEvent,
    "target" | "defaultPrevented" | "isComposing" | "shiftKey" | "ctrlKey" | "altKey" | "metaKey"
  >
>;
const key = (k: string, over: KeyInit = {}) => ({
  key: k,
  target: document.body as EventTarget | null,
  defaultPrevented: false,
  isComposing: false,
  shiftKey: false,
  ctrlKey: false,
  altKey: false,
  metaKey: false,
  ...over,
});
const open = { modalOpen: false };

describe("which button a key presses", () => {
  it("Enter presses the primary and Escape the cancel", () => {
    const r = flow();
    expect(dockKeyFor(key("Enter"), r, open)?.id).toBe("done");
    expect(dockKeyFor(key("Escape"), r, open)?.id).toBe("cancel");
  });

  it("only a button that advertises the key, and never a disabled one", () => {
    // A yes/no question's Yes does not advertise Enter (PR 5): a stray
    // Enter must not accept an optional effect.
    expect(dockKeyAction(flow({ primary: press("yes") }), "Enter")).toBeNull();
    expect(
      dockKeyAction(flow({ primary: press("done", "Enter", { disabled: true }) }), "Enter"),
    ).toBeNull();
    expect(dockKeyAction(flow({ secondary: [press("cast-anyway")] }), "Escape")).toBeNull();
    expect(dockKeyAction(flow(), "Tab")).toBeNull();
  });

  it("never presses anything for a step row: Enter is not next", () => {
    const row: DockRequest = {
      rank: "step",
      label: "declare attackers",
      row: [press("attack-all", "Enter")],
      primary: press("x", "Enter"),
    };
    expect(dockKeyFor(key("Enter"), row, open)).toBeNull();
    expect(dockKeyFor(key("Escape"), row, open)).toBeNull();
    expect(dockKeyFor(key("Enter"), null, open)).toBeNull();
    expect(dockKeyFor(key("Escape"), null, open)).toBeNull();
  });

  it("each request kind: targeting, combat selection, blocks", () => {
    const cancel = vi.fn();
    const sel = combatSelectionRequest("attacker", "Grizzly Bears", cancel);
    expect(dockKeyFor(key("Enter"), sel, open)).toBeNull();
    dockKeyFor(key("Escape"), sel, open)!.onPress();
    expect(cancel).toHaveBeenCalledTimes(1);

    const finish = vi.fn();
    // Done blocking commits what was staged: Enter presses it.
    dockKeyFor(key("Enter"), blockRequest(2, finish), open)!.onPress();
    expect(finish).toHaveBeenCalledTimes(1);
    // No blocks declines a window that cannot be got back, and passes:
    // a click, never a stray Enter. Neither has an Escape.
    expect(dockKeyFor(key("Enter"), blockRequest(0, finish), open)).toBeNull();
    expect(dockKeyFor(key("Escape"), blockRequest(2, finish), open)).toBeNull();
  });
});

describe("when the handler stands down", () => {
  it("behind a modal layer (#1659): the modal owns its keys", () => {
    expect(dockKeyFor(key("Enter"), flow(), { modalOpen: true })).toBeNull();
    expect(dockKeyFor(key("Escape"), flow(), { modalOpen: true })).toBeNull();
  });

  it("while typing, for both keys", () => {
    for (const html of [
      "<input />",
      "<textarea></textarea>",
      "<select><option>a</option></select>",
      "<div contenteditable='true'><span>x</span></div>",
    ]) {
      document.body.innerHTML = html;
      const el = (document.body.querySelector("span") ??
        document.body.firstElementChild) as HTMLElement;
      expect(dockKeyFor(key("Enter", { target: el }), flow(), open), html).toBeNull();
      expect(dockKeyFor(key("Escape", { target: el }), flow(), open), html).toBeNull();
    }
  });

  it("Enter on a focused control belongs to it; Escape still cancels", () => {
    document.body.innerHTML =
      "<button id='b'>Pass turn</button><div id='card' role='button' tabindex='0'><img /></div>";
    for (const id of ["b", "card"]) {
      const el = document.getElementById(id)!;
      expect(dockKeyFor(key("Enter", { target: el }), flow(), open)).toBeNull();
      expect(dockKeyFor(key("Escape", { target: el }), flow(), open)?.id).toBe("cancel");
    }
  });

  it("with a modifier, mid-composition, or once something handled the key", () => {
    expect(dockKeyFor(key("Enter", { shiftKey: true }), flow(), open)).toBeNull();
    expect(dockKeyFor(key("Enter", { ctrlKey: true }), flow(), open)).toBeNull();
    expect(dockKeyFor(key("Escape", { altKey: true }), flow(), open)).toBeNull();
    expect(dockKeyFor(key("Escape", { metaKey: true }), flow(), open)).toBeNull();
    expect(dockKeyFor(key("Enter", { isComposing: true }), flow(), open)).toBeNull();
    expect(dockKeyFor(key("Escape", { defaultPrevented: true }), flow(), open)).toBeNull();
  });
});
