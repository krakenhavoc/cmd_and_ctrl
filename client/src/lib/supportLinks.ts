// supportLinks.ts — the optional Ko-fi and Patreon links on the funding
// page (#3558, ADR 0179 "Supporter links").
//
// Pure: no stores, no DOM. The operator's links come from GET /config
// (env.ts parses them with parseSupport). The server checks the same rule
// at boot (server/internal/supportlinks); this client re-checks and drops
// a link that fails, so a server and client that ever disagree fail
// closed. Unset means the link is not shown.

/** A support service this client can link to. */
export type SupportId = "kofi" | "patreon";

/** One configured support link, ready to render. */
export interface SupportLink {
  id: SupportId;
  /** What the link says. */
  name: string;
  /** The https URL on the service's own domain. */
  href: string;
}

/** Every service, in the order the links render, with its own domain. */
const SERVICES: readonly { id: SupportId; name: string; domain: string }[] = [
  { id: "kofi", name: "Ko-fi", domain: "ko-fi.com" },
  { id: "patreon", name: "Patreon", domain: "patreon.com" },
];

/** GitHub Sponsors, as .github/FUNDING.yml has it. */
export const GITHUB_SPONSORS_URL = "https://github.com/sponsors/krakenhavoc";

/** The configured links, as GET /config serves them under `support`. */
export type Support = Partial<Record<SupportId, string>>;

/** No support link configured: the default and every failure path. */
export const NO_SUPPORT: Support = {};

/** Whether `raw` is an https URL on `domain` (or www.domain), nothing else. */
export function validSupportUrl(raw: string, domain: string): boolean {
  if (raw.length > 512 || !/^[\x21-\x7e]+$/.test(raw) || /["'`\\<>{}|^]/.test(raw)) return false;
  let u: URL;
  try {
    u = new URL(raw);
  } catch {
    return false;
  }
  return (
    u.protocol === "https:" &&
    u.username === "" &&
    u.password === "" &&
    u.port === "" &&
    (u.hostname === domain || u.hostname === `www.${domain}`)
  );
}

/** parseSupport narrows GET /config's `support` onto the valid links. Never throws. */
export function parseSupport(raw: unknown): Support {
  if (typeof raw !== "object" || raw === null) return NO_SUPPORT;
  const b = raw as Record<string, unknown>;
  const out: Support = {};
  for (const s of SERVICES) {
    const v = b[s.id];
    if (typeof v === "string" && validSupportUrl(v, s.domain)) out[s.id] = v;
  }
  return out;
}

/** supportLinks lists the configured links in render order. */
export function supportLinks(support: Support): SupportLink[] {
  return SERVICES.flatMap((s) => {
    const href = support[s.id];
    return href ? [{ id: s.id, name: s.name, href }] : [];
  });
}
