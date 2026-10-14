// shopLinks.ts — links to buy the real cards at a retailer (ADR 0179).
//
// Pure: no stores, no DOM. The operator's templates come from GET
// /config (env.ts parses them with parseShops); which ones a viewer sees
// is shopLinksStore.ts. The rules every caller keeps:
//
//   - Nothing renders unless the operator configured a shop.
//   - Links are built from the card name (the front face), never from
//     Scryfall's purchase_uris, and no price is ever shown.
//   - Never on the game table while a game is running.
//   - SHOP_DISCLOSURE is shown next to every group of links.
//
// The destination formats are each retailer's own public pages:
//   - TCGplayer search: /search/magic/product?productLineName=magic&q=…,
//     the route its own search box builds.
//   - TCGplayer Mass Entry: /massentry?productline=Magic&c=…, each line
//     "<count> <name>" percent-encoded and joined by "||", the format its
//     own "share this list" link builds.
//   - Card Kingdom search: /catalog/search?search=header&filter[name]=….
//   - Card Kingdom Deck Builder: /builder?c=…, lines joined by newlines;
//     the page fills its list box from it.

/** The disclosure shown with every group of shop links. */
export const SHOP_DISCLOSURE = "Shop links may earn CMD & CTRL a commission.";

/** A retailer this client can build links for. */
export type ShopId = "tcgplayer" | "cardkingdom";

/** Every retailer this client knows, in the order the links render. */
export const SHOP_IDS: readonly ShopId[] = ["tcgplayer", "cardkingdom"];

/** One configured shop, as GET /config serves it. */
export interface Shop {
  id: ShopId;
  /** What the link says: "TCGplayer", "Card Kingdom". */
  name: string;
  /** The operator's URL template, holding {url} once. */
  template: string;
}

/** One line of a list cart: a card and how many copies. */
export interface CardLine {
  name: string;
  count: number;
}

/** A link ready to render. */
export interface ShopLink {
  id: ShopId;
  name: string;
  href: string;
}

const PLACEHOLDER = "{url}";
const MAX_TEMPLATE_BYTES = 1024;

const SHOP_NAMES: Record<ShopId, string> = {
  tcgplayer: "TCGplayer",
  cardkingdom: "Card Kingdom",
};

/**
 * validTemplate applies the server's rules (shoplinks.ValidateTemplate)
 * again, so a client that disagrees with its server fails closed: a
 * template either is an https URL with {url} once after the host, or
 * starts with {url} and adds only &name=value pairs.
 */
export function validTemplate(t: string): boolean {
  if (t.length === 0 || t.length > MAX_TEMPLATE_BYTES) return false;
  if (t.split(PLACEHOLDER).length !== 2) return false;
  const rest = t.replace(PLACEHOLDER, "");
  // Plain printable ASCII, and none of the characters that could leave
  // an attribute or start another placeholder.
  if (!/^[\x21-\x7e]*$/.test(rest) || /["'`\\<>{}|^]/.test(rest)) return false;
  if (t.startsWith(PLACEHOLDER)) {
    const tail = t.slice(PLACEHOLDER.length);
    if (tail === "") return true;
    if (!tail.startsWith("&") || tail.length === 1) return false;
    for (const pair of tail.slice(1).split("&")) {
      const [name, value = ""] = pair.split("=", 2);
      // Go's url.ParseQuery refuses a semicolon, so the server does too.
      if (!name || pair.includes(";")) return false;
      try {
        decodeURIComponent(name);
        decodeURIComponent(value);
      } catch {
        return false;
      }
    }
    return true;
  }
  const m = /^https:\/\/([^/?#]+)(.*)$/.exec(t);
  if (!m) return false;
  const host = m[1];
  // No userinfo, and the placeholder after the host.
  if (host.includes("@") || host.includes(PLACEHOLDER)) return false;
  try {
    const u = new URL(t.replace(PLACEHOLDER, "x"));
    return u.protocol === "https:" && u.hostname !== "";
  } catch {
    return false;
  }
}

/**
 * parseShops narrows GET /config's `shops` onto the shops this client
 * can build links for: a known id, a valid template, each id once, in
 * SHOP_IDS order. Anything else is dropped, never thrown.
 */
export function parseShops(raw: unknown): Shop[] {
  if (!Array.isArray(raw)) return [];
  const byId = new Map<ShopId, Shop>();
  for (const item of raw) {
    if (typeof item !== "object" || item === null) continue;
    const { id, template } = item as { id?: unknown; template?: unknown };
    if (typeof id !== "string" || !(SHOP_IDS as readonly string[]).includes(id)) continue;
    if (typeof template !== "string" || !validTemplate(template)) continue;
    const sid = id as ShopId;
    if (!byId.has(sid)) byId.set(sid, { id: sid, name: SHOP_NAMES[sid], template });
  }
  return SHOP_IDS.flatMap((id) => {
    const s = byId.get(id);
    return s ? [s] : [];
  });
}

/**
 * applyTemplate puts a destination into a template. A wrapper gets it
 * percent-encoded; a template that starts with {url} gets it as it is,
 * followed by the template's own pairs (every destination here already
 * has a query string).
 */
export function applyTemplate(template: string, destination: string): string {
  if (template.startsWith(PLACEHOLDER)) {
    return destination + template.slice(PLACEHOLDER.length);
  }
  return template.replace(PLACEHOLDER, () => encodeURIComponent(destination));
}

/**
 * shopCardName is the name a retailer searches by: the front face of a
 * double-faced, split or adventure card ("Fire // Ice" is "Fire"), which
 * both retailers find, trimmed.
 */
export function shopCardName(name: string): string {
  return name.split(" // ")[0].trim();
}

/** The retailer's search for one card by name. */
export function cardSearchURL(id: ShopId, name: string): string {
  const q = encodeURIComponent(shopCardName(name));
  switch (id) {
    case "tcgplayer":
      return `https://www.tcgplayer.com/search/magic/product?productLineName=magic&q=${q}&view=grid`;
    case "cardkingdom":
      return `https://www.cardkingdom.com/catalog/search?search=header&filter%5Bname%5D=${q}`;
  }
}

/**
 * cartLines merges a list by front-face name (summing copies), drops
 * blank names and counts below one, and keeps first-seen order.
 */
export function cartLines(list: readonly CardLine[]): CardLine[] {
  const merged = new Map<string, CardLine>();
  for (const l of list) {
    const name = shopCardName(l.name);
    const count = Math.floor(l.count);
    if (!name || !Number.isFinite(count) || count < 1) continue;
    const prev = merged.get(name.toLowerCase());
    if (prev) prev.count += count;
    else merged.set(name.toLowerCase(), { name, count });
  }
  return [...merged.values()];
}

/** The retailer's pasted-list cart for a list of cards. */
export function listCartURL(id: ShopId, list: readonly CardLine[]): string {
  const lines = cartLines(list).map((l) => `${l.count} ${l.name}`);
  switch (id) {
    case "tcgplayer":
      return `https://www.tcgplayer.com/massentry?productline=Magic&c=${lines
        .map((l) => encodeURIComponent(l))
        .join("||")}`;
    case "cardkingdom":
      return `https://www.cardkingdom.com/builder?c=${encodeURIComponent(lines.join("\n"))}`;
  }
}

/** One link per shop to search for a card. Empty with no shops. */
export function cardShopLinks(shops: readonly Shop[], name: string): ShopLink[] {
  if (!shopCardName(name)) return [];
  return shops.map((s) => ({
    id: s.id,
    name: s.name,
    href: applyTemplate(s.template, cardSearchURL(s.id, name)),
  }));
}

/** One link per shop to put a list in a cart. Empty with no shops or no cards. */
export function listShopLinks(shops: readonly Shop[], list: readonly CardLine[]): ShopLink[] {
  if (cartLines(list).length === 0) return [];
  return shops.map((s) => ({
    id: s.id,
    name: s.name,
    href: applyTemplate(s.template, listCartURL(s.id, list)),
  }));
}
