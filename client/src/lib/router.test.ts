// router.test.ts — parseHash coverage for the site portal (#1386):
// #/home and #/roadmap, plus a couple of the existing routes so a
// regression here doesn't silently start returning the default
// (login) route for either of them.

import { describe, expect, it } from "vitest";

import { parseHash } from "./router";

describe("parseHash", () => {
  it("parses #/home", () => {
    expect(parseHash("#/home")).toEqual({ name: "home" });
  });

  it("parses #/roadmap", () => {
    expect(parseHash("#/roadmap")).toEqual({ name: "roadmap" });
  });

  it("still parses the pre-existing routes", () => {
    expect(parseHash("#/login")).toEqual({ name: "login" });
    expect(parseHash("#/lobby")).toEqual({ name: "lobby" });
    expect(parseHash("#/catalog")).toEqual({ name: "catalog" });
    expect(parseHash("#/my-games")).toEqual({ name: "myGames" });
  });

  it("falls back to login for an empty or unknown hash", () => {
    expect(parseHash("")).toEqual({ name: "login" });
    expect(parseHash("#/nowhere")).toEqual({ name: "login" });
  });
});
