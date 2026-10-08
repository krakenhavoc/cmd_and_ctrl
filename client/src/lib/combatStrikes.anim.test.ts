// @vitest-environment jsdom
//
// ADR 0134 PR 2: the parts of the lethal hit and the crumble that the
// real animation helpers do at once, before any tween has run. The
// timings and the shard plan are pinned in combatStrikes.test.ts.

import { afterEach, describe, expect, it } from "vitest";

import { IMPACT_LETHAL_WASH, crumble, impactShake, setAnimationConfig } from "./animations";
import { crumbleShards } from "./combatStrikes";

afterEach(() => {
  setAnimationConfig({ enabled: true, combat: true });
  document.body.innerHTML = "";
});

function face(): HTMLElement {
  const el = document.createElement("div");
  el.appendChild(Object.assign(document.createElement("img"), { src: "/bear.jpg" }));
  document.body.appendChild(el);
  return el;
}

describe("impactShake", () => {
  it("turns the flash red for a lethal hit", () => {
    const el = face();
    void impactShake(el, { lethal: true });
    expect(el.style.getPropertyValue("--impact-wash")).toBe(IMPACT_LETHAL_WASH);
  });

  it("leaves an ordinary hit's flash white", () => {
    const el = face();
    void impactShake(el);
    expect(el.style.getPropertyValue("--impact-wash")).toBe("");
  });

  it("does nothing with combat motion off", async () => {
    setAnimationConfig({ combat: false });
    const el = face();
    await impactShake(el, { lethal: true });
    expect(el.style.getPropertyValue("--impact-wash")).toBe("");
  });
});

describe("crumble", () => {
  it("lays one shard per piece over the face, each the face's art clipped to it", () => {
    const el = face();
    const shards = crumbleShards("bear", 80, 112);
    void crumble(el, shards);
    const made = [...el.querySelectorAll<HTMLElement>("[data-strike-shard]")];
    expect(made).toHaveLength(shards.length);
    expect(made[0].style.clipPath).toBe(shards[0].clip);
    expect(made[0].querySelector("img")?.getAttribute("src")).toBe("/bear.jpg");
    expect(el.classList.contains("crumbling")).toBe(true);
  });

  it("makes nothing with combat motion off", async () => {
    setAnimationConfig({ enabled: false });
    const el = face();
    await crumble(el, crumbleShards("bear", 80, 112));
    expect(el.querySelector("[data-strike-shard]")).toBeNull();
  });
});
