// @vitest-environment jsdom
//
// pileBarLibraryPlay.render.test.ts — #1440. `libraryTopPlayable` had
// no caller: a human could see a revealed land (Oracle of Mul Daya,
// Courser of Kruphix) sitting face up on the library pile and had no
// way to play it, and a cross-seat grant (Xanathar, Guild Kingpin)
// had no surface at all. This pins the PileBar affordance that hands
// a playable top card to the Board's one cast chain with
// `fromZone: "library"`, exactly the way the graveyard's flashback
// button and the exile impulse button do (see impulseCast.render.test.ts).
//
// A render test rather than a pure one because the thing under test
// is what the BUTTON does and when it appears at all — both need real
// markup. `libraryTopActionLabel` (libraryTop.test.ts) pins the pure
// label logic underneath it.

import { describe, it, expect, afterEach } from "vitest";

import PileBar from "./components/board/PileBar.svelte";
import type { CardView, PlayerView, ZoneView } from "./protocol";
import type { CastSourceZone } from "./targeting";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function emptyZone(kind: string): ZoneView {
  return { kind, count: 0, cards: [] };
}

function card(over: Partial<CardView> = {}): CardView {
  return {
    instance_id: "top",
    name: "Top Card",
    owner: "them",
    controller: "them",
    known_by_you: true,
    ...over,
  };
}

function seatWith(libraryCards: CardView[]): PlayerView {
  return {
    id: "them",
    name: "Them",
    seat: 1,
    life: 40,
    library: { kind: "library", count: libraryCards.length, cards: libraryCards },
    hand: emptyZone("hand"),
    graveyard: emptyZone("graveyard"),
    command: emptyZone("command"),
    commander_damage: {},
    life_history: [],
  } as unknown as PlayerView;
}

interface HandedUp {
  card: CardView;
  zone?: CastSourceZone;
  face?: number;
}

function mount(libraryCards: CardView[], opts: { isSelf?: boolean; withHandler?: boolean } = {}) {
  const { isSelf = false, withHandler = true } = opts;
  const handed: HandedUp[] = [];
  const view = render(
    PileBar as never,
    {
      seat: seatWith(libraryCards),
      exile: emptyZone("exile"),
      isSelf,
      sendAction: () => {},
      onPlayCard: withHandler
        ? (c: CardView, zone?: CastSourceZone, face?: number) =>
            handed.push({ card: c, zone, face })
        : undefined,
    } as never,
  );
  return { container: view.container, handed };
}

const playButton = (c: HTMLElement): HTMLElement | null => c.querySelector(".pile-action");

describe("the library-top play affordance — #1440", () => {
  it("shows no button when the library is empty", () => {
    const { container } = mount([]);
    expect(playButton(container)).toBeNull();
  });

  it("shows no button for a revealed card the permission doesn't open", () => {
    // Oracle of Mul Daya reveals the top card to the whole table but
    // opens only lands: visible, unplayable, no button.
    const { container } = mount([card({ type_line: "Sorcery" })]);
    expect(playButton(container)).toBeNull();
  });

  it("shows no button when there is nothing behind it", () => {
    const { container } = mount([card({ type_line: "Basic Land — Forest", castable_here: true })], {
      withHandler: false,
    });
    expect(playButton(container)).toBeNull();
  });

  it("labels a land 'play' and hands it to the chain with fromZone library", () => {
    const land = card({
      instance_id: "land1",
      name: "Top Forest",
      type_line: "Basic Land — Forest",
      castable_here: true,
    });
    const { container, handed } = mount([land]);
    const button = playButton(container);
    expect(button).not.toBeNull();
    expect(button!.textContent?.trim()).toBe("play");

    click(button!);
    expect(handed).toHaveLength(1);
    expect(handed[0].card.instance_id).toBe("land1");
    expect(handed[0].zone).toBe("library");
  });

  it("labels anything else 'cast'", () => {
    const spell = card({
      instance_id: "spell1",
      name: "Top Sorcery",
      type_line: "Sorcery",
      castable_here: true,
    });
    const { container, handed } = mount([spell]);
    const button = playButton(container);
    expect(button!.textContent?.trim()).toBe("cast");

    click(button!);
    expect(handed[0].zone).toBe("library");
  });

  it("shows the button even when the viewer isn't this seat's owner", () => {
    // Xanathar, Guild Kingpin: the permission's holder and the pile's
    // owner are different seats. `castable_here` is already the
    // VIEWER's own answer (#1055), so the affordance is not gated on
    // `isSelf` the way the draw-a-card click is.
    const land = card({ instance_id: "land2", type_line: "Land", castable_here: true });
    const { container, handed } = mount([land], { isSelf: false });
    const button = playButton(container);
    expect(button).not.toBeNull();
    click(button!);
    expect(handed[0].card.instance_id).toBe("land2");
  });
});
