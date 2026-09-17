import { describe, expect, it } from "vitest";

import { doubledTriggerLabel } from "./triggerDoubling";

describe("doubledTriggerLabel", () => {
  it("uses the public doubler name in the compact attribution", () => {
    expect(doubledTriggerLabel("permanent-id", "Panharmonicon")).toBe("additional (Panharmonicon)");
  });

  it("does not manufacture an attribution without server metadata", () => {
    expect(doubledTriggerLabel()).toBeNull();
  });

  it("keeps attribution generic when the name is unavailable", () => {
    expect(doubledTriggerLabel("permanent-id")).toBe("additional trigger");
  });
});
