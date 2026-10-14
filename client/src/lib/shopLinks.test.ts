// ADR 0179: the shop link builders. The destination formats are each
// retailer's own public pages; these tests pin them, so a change to one
// is a deliberate edit here too.

import { describe, expect, it } from "vitest";

import {
  applyTemplate,
  cardSearchURL,
  cardShopLinks,
  cartLines,
  listCartURL,
  listShopLinks,
  parseShops,
  shopCardName,
  validTemplate,
  SHOP_DISCLOSURE,
  type Shop,
} from "./shopLinks";

describe("validTemplate", () => {
  it("accepts a wrapper and a direct link with pairs", () => {
    for (const t of [
      "{url}",
      "{url}&partner=cmdctrl",
      "{url}&partner=cmdctrl&utm_source=cmdctrl",
      "https://partner.example/c/123?u={url}",
      "https://partner.example/c/123/{url}",
      "https://partner.example/c/123?subId1=c2%2Cbuy&u={url}",
    ]) {
      expect(validTemplate(t), t).toBe(true);
    }
  });

  it("refuses everything else, as the server does", () => {
    for (const t of [
      "",
      "https://partner.example/c/123",
      "https://partner.example/?u={url}&v={url}",
      "http://partner.example/?u={url}",
      "javascript:alert(1)//{url}",
      "partner.example/?u={url}",
      "https://user@partner.example/?u={url}",
      "https://{url}.example/",
      "https://{url}",
      "https://partner.example/?u={url}&x=a b",
      'https://partner.example/?u={url}"',
      "https://partner.example/<x>?u={url}",
      "https://partner.example/?u={url}&x={y",
      "https://partner.example/?u={url}&x=é",
      "{url}&",
      "{url}/more",
      "{url}&=x",
      "{url}&x=%zz",
      "{url}&a;b=1",
      "x{url}",
      "https://partner.example/?u={url}&x=" + "a".repeat(1024),
    ]) {
      expect(validTemplate(t), t).toBe(false);
    }
  });
});

describe("parseShops", () => {
  it("keeps known shops with valid templates, once each, in order", () => {
    const got = parseShops([
      { id: "cardkingdom", template: "{url}&partner=c2" },
      { id: "tcgplayer", template: "https://partner.example/c/1?u={url}" },
      { id: "tcgplayer", template: "{url}" },
      { id: "nowhere", template: "{url}" },
      { id: "cardkingdom", template: "http://x.example/{url}" },
      null,
      "tcgplayer",
    ]);
    expect(got).toEqual([
      { id: "tcgplayer", name: "TCGplayer", template: "https://partner.example/c/1?u={url}" },
      { id: "cardkingdom", name: "Card Kingdom", template: "{url}&partner=c2" },
    ]);
  });

  it("is empty for anything but a list", () => {
    expect(parseShops(undefined)).toEqual([]);
    expect(parseShops({ id: "tcgplayer", template: "{url}" })).toEqual([]);
  });
});

describe("applyTemplate", () => {
  const dest = "https://www.cardkingdom.com/builder?c=1%20Sol%20Ring";

  it("percent-encodes the destination into a wrapper", () => {
    expect(applyTemplate("https://partner.example/c/1?u={url}", dest)).toBe(
      "https://partner.example/c/1?u=https%3A%2F%2Fwww.cardkingdom.com%2Fbuilder%3Fc%3D1%2520Sol%2520Ring",
    );
  });

  it("uses the destination as it is for a direct link, then the pairs", () => {
    expect(applyTemplate("{url}", dest)).toBe(dest);
    expect(applyTemplate("{url}&partner=c2", dest)).toBe(`${dest}&partner=c2`);
  });

  it("does not treat $ in a destination as a replacement pattern", () => {
    expect(applyTemplate("https://p.example/?u={url}", "https://x.example/?q=$&")).toBe(
      "https://p.example/?u=https%3A%2F%2Fx.example%2F%3Fq%3D%24%26",
    );
  });
});

describe("destinations", () => {
  it("searches by the front face", () => {
    expect(shopCardName("Delver of Secrets // Insectile Aberration")).toBe("Delver of Secrets");
    expect(shopCardName("  Sol Ring ")).toBe("Sol Ring");
  });

  it("builds each retailer's card search", () => {
    expect(cardSearchURL("tcgplayer", "Atraxa, Praetors' Voice")).toBe(
      "https://www.tcgplayer.com/search/magic/product?productLineName=magic&q=Atraxa%2C%20Praetors'%20Voice&view=grid",
    );
    expect(cardSearchURL("cardkingdom", "Fire // Ice")).toBe(
      "https://www.cardkingdom.com/catalog/search?search=header&filter%5Bname%5D=Fire",
    );
  });

  it("merges a list by name and drops empty lines", () => {
    expect(
      cartLines([
        { name: "Sol Ring", count: 1 },
        { name: "Forest", count: 30 },
        { name: "sol ring", count: 1 },
        { name: "", count: 2 },
        { name: "Island", count: 0 },
        { name: "Delver of Secrets // Insectile Aberration", count: 1 },
      ]),
    ).toEqual([
      { name: "Sol Ring", count: 2 },
      { name: "Forest", count: 30 },
      { name: "Delver of Secrets", count: 1 },
    ]);
  });

  it("builds TCGplayer Mass Entry: encoded lines joined by ||", () => {
    expect(
      listCartURL("tcgplayer", [
        { name: "Sol Ring", count: 1 },
        { name: "Forest", count: 30 },
      ]),
    ).toBe("https://www.tcgplayer.com/massentry?productline=Magic&c=1%20Sol%20Ring||30%20Forest");
  });

  it("builds the Card Kingdom Deck Builder: lines joined by newlines", () => {
    expect(
      listCartURL("cardkingdom", [
        { name: "Sol Ring", count: 1 },
        { name: "Forest", count: 30 },
      ]),
    ).toBe("https://www.cardkingdom.com/builder?c=1%20Sol%20Ring%0A30%20Forest");
  });
});

describe("link groups", () => {
  const shops: Shop[] = [
    { id: "tcgplayer", name: "TCGplayer", template: "{url}&partner=c2" },
    { id: "cardkingdom", name: "Card Kingdom", template: "https://p.example/?u={url}" },
  ];

  it("renders nothing with no shop configured", () => {
    expect(cardShopLinks([], "Sol Ring")).toEqual([]);
    expect(listShopLinks([], [{ name: "Sol Ring", count: 1 }])).toEqual([]);
  });

  it("renders nothing for no card", () => {
    expect(cardShopLinks(shops, " ")).toEqual([]);
    expect(listShopLinks(shops, [])).toEqual([]);
  });

  it("gives one https link per shop", () => {
    const card = cardShopLinks(shops, "Sol Ring");
    expect(card.map((l) => l.name)).toEqual(["TCGplayer", "Card Kingdom"]);
    expect(card[0].href).toBe(
      "https://www.tcgplayer.com/search/magic/product?productLineName=magic&q=Sol%20Ring&view=grid&partner=c2",
    );
    const list = listShopLinks(shops, [{ name: "Sol Ring", count: 1 }]);
    expect(list[1].href).toBe(
      "https://p.example/?u=" +
        encodeURIComponent("https://www.cardkingdom.com/builder?c=1%20Sol%20Ring"),
    );
    for (const l of [...card, ...list]) expect(l.href.startsWith("https://")).toBe(true);
  });

  it("names CMD & CTRL in the disclosure", () => {
    expect(SHOP_DISCLOSURE).toBe("Shop links may earn CMD & CTRL a commission.");
  });
});
