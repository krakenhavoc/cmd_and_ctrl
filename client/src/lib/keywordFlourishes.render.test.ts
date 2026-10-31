// @vitest-environment jsdom
//
// #4556: the keyword flourishes on battlefield cards. What only exists
// once the markup is rendered: the classes and elements a card wears for
// its EFFECTIVE keywords, that they follow a grant and a loss, and that
// the Settings toggle, reduced motion and the animations switch take the
// right parts away.

import { afterEach, beforeEach, describe, expect, it } from "vitest";

import Card from "./components/board/Card.svelte";
import panel from "./components/board/PlayerPanel.svelte?raw";
import type { CardView } from "./protocol";
import { flourishesFor, swayDelay } from "./keywordFlourishes";
import { defaultSettings, resetSettings, updateSettings } from "./settings";
import { render, cleanup } from "./test/render.svelte";

beforeEach(() => resetSettings());
afterEach(cleanup);

const creature = (abilities: string[], extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "c1",
    name: "Test Creature",
    owner: "me",
    controller: "me",
    scryfall_id: "aaaa-bbbb",
    type_line: "Creature — Test",
    power: 2,
    toughness: 2,
    known_by_you: true,
    abilities,
    ...extra,
  }) as unknown as CardView;

function mount(card: CardView, props: Record<string, unknown> = { flourishes: true }) {
  const r = render(Card as never, { card, ...props } as Record<string, unknown>);
  const el = () => r.container.querySelector(".card") as HTMLElement;
  return { ...r, el };
}

const blades = (el: HTMLElement) => el.querySelectorAll(".kw-blade").length;

describe("flourishesFor", () => {
  it("reads each keyword, case-insensitively", () => {
    expect(flourishesFor(["Flying"]).flying).toBe(true);
    expect(flourishesFor(["reach"]).reach).toBe(true);
    expect(flourishesFor(["deathtouch"]).deathtouch).toBe(true);
    expect(flourishesFor(["lifelink"]).lifelink).toBe(true);
    expect(flourishesFor(["first strike"]).blades).toBe(1);
    expect(flourishesFor(["double strike"]).blades).toBe(2);
  });

  it("shows two blades for double strike with first strike, not three", () => {
    expect(flourishesFor(["first strike", "double strike"]).blades).toBe(2);
  });

  it("has nothing for keywords it has no flourish for, or for no list", () => {
    expect(flourishesFor(["trample", "haste"])).toEqual(flourishesFor([]));
    expect(flourishesFor(undefined).blades).toBe(0);
    expect(flourishesFor(null).flying).toBe(false);
  });

  it("lets flying win over reach: a creature that hovers does not also stretch", () => {
    const f = flourishesFor(["flying", "reach"]);
    expect(f.flying).toBe(true);
    expect(f.reach).toBe(false);
  });

  it("gives each card its own stable sway phase", () => {
    expect(swayDelay("a")).toBe(swayDelay("a"));
    expect(swayDelay("a")).not.toBe(swayDelay("b"));
    expect(swayDelay("anything")).toMatch(/^-\d+ms$/);
  });
});

describe("a battlefield card's keyword flourishes", () => {
  it("wears none with no keyword", () => {
    const { el, container } = mount(creature([]));
    expect(el().className).not.toMatch(/kw-/);
    expect(container.querySelector(".kw-fx")).toBeNull();
  });

  it("flying hovers and sways", () => {
    const { el } = mount(creature(["flying"]));
    expect(el().classList.contains("kw-fly")).toBe(true);
    expect(el().classList.contains("kw-sway")).toBe(true);
    expect(el().style.getPropertyValue("--kw-delay")).toMatch(/^-\d+ms$/);
  });

  it("reach draws a bow and does not hover", () => {
    const { el, container } = mount(creature(["reach"]));
    expect(el().classList.contains("kw-reach")).toBe(true);
    expect(el().classList.contains("kw-fly")).toBe(false);
    expect(container.querySelector(".kw-bow")).not.toBeNull();
  });

  it("deathtouch gets the dark edge", () => {
    const { container, el } = mount(creature(["deathtouch"]));
    expect(el().classList.contains("kw-deathtouch")).toBe(true);
    expect(container.querySelector(".kw-fx.kw-edge")).not.toBeNull();
  });

  it("first strike carries one blade, double strike two, both together two", () => {
    expect(blades(mount(creature(["first strike"])).el())).toBe(1);
    cleanup();
    expect(blades(mount(creature(["double strike"])).el())).toBe(2);
    cleanup();
    expect(blades(mount(creature(["first strike", "double strike"])).el())).toBe(2);
  });

  it("lifelink glows", () => {
    expect(
      mount(creature(["lifelink"]))
        .el()
        .classList.contains("kw-lifelink"),
    ).toBe(true);
  });

  it("combines: a flying lifelink deathtouch card hovers, glows and has the edge", () => {
    const { el, container } = mount(creature(["flying", "lifelink", "deathtouch", "first strike"]));
    const cl = el().classList;
    expect(
      cl.contains("kw-fly") && cl.contains("kw-lifelink") && cl.contains("kw-deathtouch"),
    ).toBe(true);
    expect(container.querySelector(".kw-fx.kw-edge")).not.toBeNull();
    expect(blades(el())).toBe(1);
  });

  it("follows the effective keywords: a granted one appears, a lost one goes", () => {
    const { el, setProps, container } = mount(creature([]));
    expect(el().classList.contains("kw-fly")).toBe(false);
    setProps({ card: creature(["flying", "first strike"]) });
    expect(el().classList.contains("kw-fly")).toBe(true);
    expect(blades(el())).toBe(1);
    setProps({ card: creature(["first strike"]) });
    expect(el().classList.contains("kw-fly")).toBe(false);
    expect(blades(el())).toBe(1);
    setProps({ card: creature([]) });
    expect(blades(el())).toBe(0);
    expect(container.querySelector(".kw-fx")).toBeNull();
  });

  it("leaves a card that is not on the battlefield row alone", () => {
    const { el, container } = mount(creature(["flying", "deathtouch", "double strike"]), {});
    expect(el().className).not.toMatch(/kw-/);
    expect(container.querySelector(".kw-fx")).toBeNull();
  });

  it("draws nothing on a face-down card", () => {
    const { el } = mount(creature(["flying"], { face_down: true } as Partial<CardView>), {
      flourishes: true,
      faceDown: true,
    });
    expect(el().classList.contains("kw-fly")).toBe(false);
  });

  it("keeps the keyword chips", () => {
    const { container } = mount(creature(["flying", "lifelink"]));
    expect(container.querySelectorAll(".kw-badge").length).toBe(2);
  });
});

describe("the setting and reduced motion", () => {
  it("defaults on, and Off draws none of them", () => {
    expect(defaultSettings().display.keywordFlourishes).toBe(true);
    updateSettings("display", "keywordFlourishes", false);
    const { el, container } = mount(
      creature(["flying", "deathtouch", "double strike", "lifelink"]),
    );
    expect(el().className).not.toMatch(/kw-/);
    expect(container.querySelector(".kw-fx")).toBeNull();
    expect(container.querySelector(".kw-badge")).not.toBeNull();
  });

  it("switching it back on brings them back", () => {
    updateSettings("display", "keywordFlourishes", false);
    const { el } = mount(creature(["flying"]));
    expect(el().classList.contains("kw-fly")).toBe(false);
    updateSettings("display", "keywordFlourishes", true);
    // The store is read reactively; a new mount is the plain check.
    cleanup();
    expect(
      mount(creature(["flying"]))
        .el()
        .classList.contains("kw-fly"),
    ).toBe(true);
  });

  it("reduced motion drops the sway and keeps every still cue", () => {
    updateSettings("accessibility", "reduceMotion", true);
    const { el, container } = mount(
      creature(["flying", "deathtouch", "double strike", "lifelink"]),
    );
    const cl = el().classList;
    expect(cl.contains("kw-sway")).toBe(false);
    expect(el().style.getPropertyValue("--kw-delay")).toBe("");
    expect(cl.contains("kw-fly")).toBe(true);
    expect(cl.contains("kw-lifelink")).toBe(true);
    expect(container.querySelector(".kw-fx.kw-edge")).not.toBeNull();
    expect(blades(el())).toBe(2);
  });

  it("the animations master switch off drops the sway too", () => {
    updateSettings("animations", "enabled", false);
    const { el } = mount(creature(["flying"]));
    expect(el().classList.contains("kw-sway")).toBe(false);
    expect(el().classList.contains("kw-fly")).toBe(true);
  });
});

// The facing is CSS: PlayerPanel says how each seat's cards are drawn, and
// Card reads --card-turn (lift and shadow) and --blade-turn (which edge
// carries the blades). A source check, since jsdom has no layout.
describe("how a seat's cards are drawn, for the flourishes", () => {
  it("marks a top-row opponent's turned rows, whose top edge faces the centre", () => {
    expect(panel).toMatch(
      /\.row-cards\) \{\s*transform: rotate\(180deg\);[\s\S]*?--card-turn: -1;\s*\}/,
    );
  });

  it("flips the blade edge where #4555 turns those cards back upright", () => {
    expect(panel).toMatch(
      /faces-viewer :is\(\.grid-creatures, \.grid-middle\) :global\(\.pile\) \{[^}]*--card-turn: 1;\s*--blade-turn: -1;/,
    );
  });
});
