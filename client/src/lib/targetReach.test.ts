// @vitest-environment jsdom
//
// targetReach.test.ts — #4435. Which legal player avatars are covered at
// their centre. jsdom has no layout, so each avatar is given a box and the
// hit test is a stub that says what is on top where.

import { afterEach, describe, expect, it } from "vitest";

import { coveredPlayerTargets, TARGET_AVATAR_SELECTOR } from "./targetReach";

afterEach(() => {
  document.body.innerHTML = "";
});

interface Box {
  x: number;
  y: number;
  w: number;
  h: number;
}

// seatAvatar draws one legal avatar for a seat: the identity ringed as a
// cast target, holding the avatar disc at `box`.
function seatAvatar(
  root: HTMLElement,
  seatID: string,
  box: Box | null,
  ring = "cast-targetable",
): HTMLElement {
  const identity = document.createElement("div");
  identity.className = `identity ${ring}`;
  const wrap = document.createElement("div");
  wrap.className = "avatar-wrap";
  wrap.dataset.seatId = seatID;
  wrap.appendChild(document.createElement("img"));
  identity.appendChild(wrap);
  root.appendChild(identity);
  const b = box ?? { x: 0, y: 0, w: 0, h: 0 };
  wrap.getBoundingClientRect = () =>
    ({
      left: b.x,
      top: b.y,
      right: b.x + b.w,
      bottom: b.y + b.h,
      width: b.w,
      height: b.h,
      x: b.x,
      y: b.y,
    }) as DOMRect;
  return wrap;
}

const table = (): HTMLElement => {
  const root = document.createElement("div");
  document.body.appendChild(root);
  return root;
};

describe("coveredPlayerTargets (#4435)", () => {
  it("lists a seat whose avatar has something else on top at its centre", () => {
    const root = table();
    seatAvatar(root, "bob", { x: 100, y: 100, w: 60, h: 60 });
    const cat = seatAvatar(root, "cat", { x: 300, y: 100, w: 60, h: 60 });
    const dialog = document.createElement("div");
    root.appendChild(dialog);
    // The centre of Bob is (130,130); a dialog sits there. Cat is clear.
    const hit = (x: number): Element => (x < 200 ? dialog : cat.firstElementChild!);
    expect(coveredPlayerTargets(root, hit)).toEqual(["bob"]);
  });

  it("is empty when every legal avatar is the top element, or something inside it", () => {
    const root = table();
    const bob = seatAvatar(root, "bob", { x: 100, y: 100, w: 60, h: 60 });
    const cat = seatAvatar(root, "cat", { x: 300, y: 100, w: 60, h: 60 });
    const hit = (x: number): Element => (x < 200 ? bob : cat.firstElementChild!);
    expect(coveredPlayerTargets(root, hit)).toEqual([]);
  });

  it("counts a point nobody is at (off the screen) as covered", () => {
    const root = table();
    seatAvatar(root, "bob", { x: -500, y: 10, w: 60, h: 60 });
    expect(coveredPlayerTargets(root, () => null)).toEqual(["bob"]);
  });

  it("a seat drawn twice (the table and an expanded board) is reachable by either copy", () => {
    const root = table();
    seatAvatar(root, "bob", { x: 100, y: 100, w: 60, h: 60 });
    const copy = seatAvatar(root, "bob", { x: 600, y: 300, w: 60, h: 60 });
    const dialog = document.createElement("div");
    root.appendChild(dialog);
    const hit = (x: number): Element => (x < 200 ? dialog : copy);
    expect(coveredPlayerTargets(root, hit)).toEqual([]);
  });

  it("a picked avatar still counts, and one with no layout is skipped", () => {
    const root = table();
    seatAvatar(root, "bob", { x: 100, y: 100, w: 60, h: 60 }, "cast-picked");
    seatAvatar(root, "cat", null);
    const dialog = document.createElement("div");
    root.appendChild(dialog);
    expect(coveredPlayerTargets(root, () => dialog)).toEqual(["bob"]);
  });

  it("only avatars the prompt may take are looked at", () => {
    const root = table();
    seatAvatar(root, "bob", { x: 100, y: 100, w: 60, h: 60 }, "targetable");
    expect(root.querySelectorAll(TARGET_AVATAR_SELECTOR)).toHaveLength(0);
    expect(coveredPlayerTargets(root, () => root)).toEqual([]);
  });
});
