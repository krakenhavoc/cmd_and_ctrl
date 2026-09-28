import { describe, expect, it } from "vitest";
import { COUNTER_MINUS_ONE, COUNTER_PLUS_ONE, COUNTER_STYLES, counterStyle } from "./counterTypes";

// counterTypes.test.ts — #1664. Every P/T counter kind changes power
// and toughness on the server now, so the pip has to say which kind it
// is: the whole name, not its first three characters.

describe("counterStyle", () => {
  it("keeps the pinned +1/+1 and -1/-1 styles", () => {
    expect(counterStyle(COUNTER_PLUS_ONE)).toBe(COUNTER_STYLES[COUNTER_PLUS_ONE]);
    expect(counterStyle(COUNTER_MINUS_ONE)).toBe(COUNTER_STYLES[COUNTER_MINUS_ONE]);
  });

  it("shows a P/T counter's whole name, green when it grows and red when it shrinks", () => {
    const green = COUNTER_STYLES[COUNTER_PLUS_ONE]?.color;
    const red = COUNTER_STYLES[COUNTER_MINUS_ONE]?.color;
    expect(counterStyle("+1/+0")).toEqual({ abbr: "+1/+0", color: green });
    expect(counterStyle("+1/+2")).toEqual({ abbr: "+1/+2", color: green });
    expect(counterStyle("-2/-1")).toEqual({ abbr: "-2/-1", color: red });
    expect(counterStyle("-0/-1")).toEqual({ abbr: "-0/-1", color: red });
  });

  it("leaves a non-P/T unknown name on the neutral three-character fallback", () => {
    expect(counterStyle("polyp").abbr).toBe("pol");
    expect(counterStyle("1/1").abbr).toBe("1/1");
    expect(counterStyle("+X/+X").abbr).toBe("+X/");
  });
});
