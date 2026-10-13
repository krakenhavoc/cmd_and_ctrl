import { describe, expect, it } from "vitest";

import { gateRefusalNotice } from "./gateRefusal";

describe("gateRefusalNotice", () => {
  it("explains a refusal", () => {
    expect(gateRefusalNotice("#/request-access")).toContain("members of the C2 Discord");
  });

  it("asks for a retry when Discord did not answer", () => {
    expect(gateRefusalNotice("#/request-access?retry=1")).toContain("didn't answer");
  });

  it("says nothing on any other route", () => {
    for (const h of ["", "#/login", "#/lobby", "#/request-accessx", "#/games/1/join?t=x"]) {
      expect(gateRefusalNotice(h)).toBe("");
    }
  });
});
