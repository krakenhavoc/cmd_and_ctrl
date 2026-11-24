import { describe, it, expect } from "vitest";
import {
  DEFAULT_PAGE_SIZE,
  PAGE_SIZES,
  catalogHash,
  clampPage,
  pageCount,
  pageRange,
  pageSlice,
  pagerItems,
  parsePage,
  parsePageSize,
} from "./catalog";
import { parseHash } from "./router";

// #6120: the catalogue's paging. Pure functions, so the decisions a
// person can see go wrong (an empty last page, "showing 51–100", a page
// link that survives a filter) are pinned here.

const nums = (n: number) => Array.from({ length: n }, (_, i) => i + 1);

describe("page sizes", () => {
  it("offers 25, 50, 75, 100 and 250, 50 by default", () => {
    expect([...PAGE_SIZES]).toEqual([25, 50, 75, 100, 250]);
    expect(DEFAULT_PAGE_SIZE).toBe(50);
  });

  it("reads an offered size and treats anything else as the default", () => {
    expect(parsePageSize("100")).toBe(100);
    expect(parsePageSize("25")).toBe(25);
    expect(parsePageSize("250")).toBe(250);
    for (const raw of [null, undefined, "", "0", "10", "1000", "abc", "50.5", "-25"]) {
      expect(parsePageSize(raw)).toBe(50);
    }
  });
});

describe("parsePage", () => {
  it("reads a whole number from 1 and falls back to page 1", () => {
    expect(parsePage("3")).toBe(3);
    expect(parsePage("1")).toBe(1);
    for (const raw of [null, undefined, "", "0", "-2", "2.5", "x", "1e3"]) {
      expect(parsePage(raw)).toBe(1);
    }
  });
});

describe("pageCount and clampPage", () => {
  it("counts pages, never fewer than one", () => {
    expect(pageCount(0, 50)).toBe(1);
    expect(pageCount(1, 50)).toBe(1);
    expect(pageCount(50, 50)).toBe(1);
    expect(pageCount(51, 50)).toBe(2);
    expect(pageCount(23150, 250)).toBe(93);
  });

  it("keeps a page inside the list", () => {
    expect(clampPage(0, 120, 50)).toBe(1);
    expect(clampPage(2, 120, 50)).toBe(2);
    // A shared link to page 40 of a list that now has 3 pages.
    expect(clampPage(40, 120, 50)).toBe(3);
    expect(clampPage(5, 0, 50)).toBe(1);
  });
});

describe("pageSlice", () => {
  const items = nums(120);

  it("returns only the current page's cards", () => {
    expect(pageSlice(items, 1, 50)).toEqual(nums(50));
    expect(pageSlice(items, 2, 50)[0]).toBe(51);
    expect(pageSlice(items, 2, 50)).toHaveLength(50);
  });

  it("gives the last page what is left, never an empty page", () => {
    expect(pageSlice(items, 3, 50)).toEqual([
      101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119,
      120,
    ]);
    expect(pageSlice(items, 9, 50)).toEqual(pageSlice(items, 3, 50));
  });

  it("is empty only for an empty list", () => {
    expect(pageSlice([], 1, 50)).toEqual([]);
  });
});

describe("pageRange", () => {
  it("is the 1-based span for the 'Showing 51–100 of N' line", () => {
    expect(pageRange(2, 230, 50)).toEqual({ from: 51, to: 100 });
    expect(pageRange(5, 230, 50)).toEqual({ from: 201, to: 230 });
    expect(pageRange(1, 7, 50)).toEqual({ from: 1, to: 7 });
    expect(pageRange(1, 0, 50)).toEqual({ from: 0, to: 0 });
    expect(pageRange(99, 230, 50)).toEqual({ from: 201, to: 230 });
  });
});

describe("pagerItems", () => {
  it("lists every page when there are few", () => {
    expect(pagerItems(1, 1)).toEqual([1]);
    expect(pagerItems(1, 3)).toEqual([1, 2, 3]);
    expect(pagerItems(3, 5)).toEqual([1, 2, 3, 4, 5]);
    expect(pagerItems(1, 7)).toEqual([1, 2, 3, 4, 5, 6, 7]);
    expect(pagerItems(1, 8)).toEqual([1, 2, "gap", 8]);
  });

  it("keeps the ends and the neighbours, with a gap between", () => {
    expect(pagerItems(10, 20)).toEqual([1, "gap", 9, 10, 11, "gap", 20]);
    expect(pagerItems(1, 20)).toEqual([1, 2, "gap", 20]);
    expect(pagerItems(20, 20)).toEqual([1, "gap", 19, 20]);
  });

  it("never lets a gap stand for a single page", () => {
    expect(pagerItems(4, 20)).toEqual([1, 2, 3, 4, 5, "gap", 20]);
    expect(pagerItems(17, 20)).toEqual([1, "gap", 16, 17, 18, 19, 20]);
  });

  it("clamps a page outside the range", () => {
    expect(pagerItems(50, 5)).toEqual([1, 2, 3, 4, 5]);
    expect(pagerItems(0, 20)).toEqual([1, 2, "gap", 20]);
  });
});

describe("catalogHash", () => {
  it("leaves the defaults out, so the plain page is #/catalog", () => {
    expect(catalogHash({ query: "", page: 1, size: 50 })).toBe("#/catalog");
    expect(catalogHash({ query: "  ", page: 1, size: 50 })).toBe("#/catalog");
  });

  it("carries the search, the page and the size", () => {
    expect(catalogHash({ query: "Sol Ring", page: 3, size: 100 })).toBe(
      "#/catalog?q=Sol+Ring&page=3&size=100",
    );
    expect(catalogHash({ query: "", page: 2, size: 50 })).toBe("#/catalog?page=2");
    expect(catalogHash({ query: "", page: 1, size: 25 })).toBe("#/catalog?size=25");
  });

  it("round-trips through the router", () => {
    const loc = { query: "Borrowing 100,000 Arrows & co", page: 4, size: 250 };
    expect(parseHash(catalogHash(loc))).toEqual({
      name: "catalog",
      query: loc.query,
      page: 4,
      size: 250,
    });
    expect(parseHash(catalogHash({ query: "", page: 1, size: 50 }))).toEqual({ name: "catalog" });
  });
});
