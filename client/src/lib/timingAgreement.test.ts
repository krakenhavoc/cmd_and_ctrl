// The TypeScript half of a cross-language contract test.
//
// S31 sub-PR 2 deleted `timing.ts`'s reimplementation of the cast
// timing rules in favour of a lookup over the server's `legal_moves`.
// Deleting a second implementation of the rules is only safe if you
// have proved the two agree first, and "agree" has to mean something
// stronger than "the survivor matches itself".
//
// So the fixture carries a third statement of the truth. Each
// scenario in server/internal/legal/testdata/timing_agreement.json
// declares BY HAND what should be castable and why;
// server/internal/legal/agreement_test.go asserts the Go enumerator
// matches that declaration, and this file asserts canCastFromHand
// matches the same declaration against the real filtered wire frame
// the enumerator produced. Either side drifting fails a test, and
// neither side is the other's oracle.
//
// Regenerate the fixture after any enumerator or view change:
//
//   go test ./internal/legal/ -run TestTimingAgreement -update
//
// Two of the scenarios — land_drop_spent and the unaffordable arm of
// sorcery_window_open — did NOT pass before the port. The old client
// checked neither the land drop nor mana, and said "castable" to
// both. They are here as the receipt.

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, it, expect } from "vitest";

import { canCastFromHand } from "./timing";
import type { CardView, GameView } from "./protocol";

interface Expectation {
  instance_id: string;
  name: string;
  castable: boolean;
  why: string;
}

interface Scenario {
  name: string;
  note: string;
  viewer: string;
  view: GameView;
  expect: Expectation[];
}

const fixtureURL = new URL(
  "../../../server/internal/legal/testdata/timing_agreement.json",
  import.meta.url,
);

const scenarios: Scenario[] = JSON.parse(readFileSync(fileURLToPath(fixtureURL), "utf8"));

// findCard walks the viewer's own hand and command zone, which is
// exactly the surface canCastFromHand is called against in
// Hand.svelte and priority.ts.
function findCard(view: GameView, viewer: string, instanceID: string): CardView | undefined {
  const seat = view.seats.find((s) => s.id === viewer);
  if (!seat) return undefined;
  return [...(seat.hand?.cards ?? []), ...(seat.command?.cards ?? [])].find(
    (c) => c.instance_id === instanceID,
  );
}

describe("timing.ts agrees with server/internal/legal", () => {
  it("the fixture is present and non-trivial", () => {
    expect(scenarios.length).toBeGreaterThan(0);
    for (const sc of scenarios) {
      expect(sc.expect.length, `${sc.name} declares no expectations`).toBeGreaterThan(0);
    }
  });

  for (const sc of scenarios) {
    describe(sc.name, () => {
      for (const e of sc.expect) {
        it(`${e.name}: castable=${e.castable} — ${e.why}`, () => {
          const card = findCard(sc.view, sc.viewer, e.instance_id);
          expect(card, `${e.name} (${e.instance_id}) is not in the viewer's zones`).toBeDefined();
          const got = canCastFromHand(card as CardView, sc.view, sc.viewer);
          expect(got.legal, `reason: ${got.reason ?? "(none)"} | ${sc.note}`).toBe(e.castable);
        });
      }
    });
  }
});

// The other direction, stated as an invariant rather than a table:
// anything the enumerator offered must come back legal, and anything
// it withheld must come back illegal, for EVERY card in the viewer's
// hand — not only the ones a scenario happened to name.
describe("every hand card agrees, not just the declared ones", () => {
  for (const sc of scenarios) {
    it(sc.name, () => {
      const seat = sc.view.seats.find((s) => s.id === sc.viewer);
      expect(seat).toBeDefined();
      const offered = new Set(
        (sc.view.legal_moves ?? [])
          .filter((m) => m.kind === "cast" || m.kind === "land")
          .map((m) => m.source),
      );
      for (const c of seat?.hand?.cards ?? []) {
        const got = canCastFromHand(c, sc.view, sc.viewer);
        expect(
          got.legal,
          `${c.name || c.instance_id}: enumerator ${offered.has(c.instance_id) ? "offered" : "withheld"} it, ` +
            `client said ${got.legal} (${got.reason ?? "legal"})`,
        ).toBe(offered.has(c.instance_id));
      }
    });
  }
});

// The hidden-information gate, re-checked on the exact committed
// bytes the client consumes: a frame must never carry a move
// belonging to another seat.
describe("no fixture frame leaks another seat's moves", () => {
  for (const sc of scenarios) {
    it(sc.name, () => {
      for (const m of sc.view.legal_moves ?? []) {
        expect(m.player, `${sc.name} leaked a move for ${m.player}`).toBe(sc.viewer);
      }
    });
  }
});
