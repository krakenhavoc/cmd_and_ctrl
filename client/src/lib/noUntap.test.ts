import { describe, expect, it } from "vitest";

import { noUntapAppliesToController, noUntapFooterLines } from "./noUntap";
import type { CardView } from "./protocol";

const card = (extra: Partial<CardView> = {}): CardView => ({
  instance_id: "card",
  name: "Mana Vault",
  owner: "alice",
  controller: "alice",
  ...extra,
});

describe("no untap projection helpers", () => {
  it("only marks a tapped permanent when the restriction applies to its controller", () => {
    expect(noUntapAppliesToController(card({ tapped: true, no_untap: { static: true } }))).toBe(
      true,
    );
    expect(noUntapAppliesToController(card({ tapped: true, no_untap: { next: ["bob"] } }))).toBe(
      false,
    );
    expect(noUntapAppliesToController(card({ no_untap: { static: true } }))).toBe(false);
  });

  it("names static and player-keyed next-step restrictions with a fallback", () => {
    expect(
      noUntapFooterLines(card({ no_untap: { static: true, next: ["bob", "gone"] } }), [
        { id: "alice", name: "Alice" },
        { id: "bob", name: "Bob" },
      ]),
    ).toEqual([
      "doesn't untap during its controller's untap step",
      "doesn't untap during Bob's next untap step",
      "doesn't untap during that player's next untap step",
    ]);
  });
});
