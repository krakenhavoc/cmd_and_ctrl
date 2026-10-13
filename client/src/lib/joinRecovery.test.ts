// #1475: what Join.svelte says to a player who lost their session and
// reopened an old player invite on a table that has already started.
// Prose assertions, deliberately — the wording IS the fix. #520
// shipped two real ways back and Join.svelte used to name neither,
// pointing everyone at "ask your host for a spectator link" instead.

import { describe, expect, it } from "vitest";

import { GUEST_RETURN_ADVICE, SIGNED_IN_RETURN_LINK } from "./joinRecovery";
import { parseHash } from "./router";

describe("GUEST_RETURN_ADVICE", () => {
  it("points a guest at a seat reclaim link", () => {
    expect(GUEST_RETURN_ADVICE).toContain("seat reclaim link");
  });

  // The bug this file exists to catch: silently reverting to the old,
  // wrong advice.
  it("never says to ask for a spectator link", () => {
    expect(GUEST_RETURN_ADVICE).not.toContain("spectator link");
  });

  it("names who to ask", () => {
    expect(GUEST_RETURN_ADVICE).toContain("host");
  });
});

describe("SIGNED_IN_RETURN_LINK", () => {
  // Tables, whose Now lists the open tables they hold a seat at (ADR
  // 0158 §1; it was My games before).
  it("sends a signed-in identity to Tables", () => {
    expect(SIGNED_IN_RETURN_LINK).toEqual({ href: "#/lobby", label: "Tables" });
    expect(parseHash(SIGNED_IN_RETURN_LINK.href)).toEqual({ name: "lobby" });
  });
});
