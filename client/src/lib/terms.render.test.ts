// @vitest-environment jsdom
//
// #3358: the terms of service, rendered. The page is public, names the
// operator and its contact from lib/legalNotice.ts (the same constants the
// privacy notice uses), and is linked from the footer and the legal notice.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";
import { get } from "svelte/store";

import App from "../App.svelte";
import LegalNotice from "./components/LegalNotice.svelte";
import SiteFooter from "./components/SiteFooter.svelte";
import Terms from "../routes/Terms.svelte";
import {
  LEGAL_CONTACT_EMAIL,
  LEGAL_CONTACT_URL,
  OPERATOR_NAME,
  OPERATOR_STATE,
} from "./legalNotice";
import { CONTACT_TEXT, CONTACT_URL } from "./privacy";
import { MIN_AGE, TERMS_HASH, TERMS_UPDATED } from "./terms";
import { REPO_URL } from "./about";
import { navigate, parseHash, route } from "./router";
import { session } from "./session";
import { cleanup, flushSync, render } from "./test/render.svelte";

vi.mock("../routes/Game.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));
vi.mock("../routes/Practice.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

const realFetch = globalThis.fetch;

beforeEach(() => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  globalThis.fetch = vi.fn(
    async () =>
      new Response("{}", { status: 200, headers: { "Content-Type": "application/json" } }),
  ) as never;
});

afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  session.set(null);
  navigate("#/");
  vi.restoreAllMocks();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

describe("the terms page", () => {
  it("names the operator, the contact, the age and the source licence", () => {
    const { container } = render(Terms as never, {} as never);
    const text = container.textContent!.replace(/\s+/g, " ");

    expect(text).toContain(`Last updated ${TERMS_UPDATED}`);
    expect(text).toContain(
      `Run by ${OPERATOR_NAME}, a ${OPERATOR_STATE} limited liability company`,
    );
    expect(text).toContain(`at least ${MIN_AGE} years old`);
    expect(text).toContain(`laws of the State of ${OPERATOR_STATE}`);
    for (const heading of [
      "Who may use it",
      "How to behave",
      "What you put on the site",
      "Copyright notices",
      "Magic: The Gathering and card data",
      "The source code",
      "No guarantees",
    ]) {
      expect(text).toContain(heading);
    }

    const hrefs = [...container.querySelectorAll("a")].map((a) => a.getAttribute("href"));
    expect(hrefs).toContain(LEGAL_CONTACT_URL);
    expect(hrefs).toContain(REPO_URL);
    expect(hrefs).toContain(`${REPO_URL}/blob/main/LICENSE`);
    expect(text).toContain("GNU Affero General Public License, version 3 or later");

    // The only address on the page is the operator's.
    const addresses = new Set(text.match(/[\w.+-]+@[\w-]+(?:\.[\w-]+)+/g));
    expect([...addresses]).toEqual([LEGAL_CONTACT_EMAIL]);
    // No copyright agent line until the owner fills in the postal details.
    expect(container.querySelector(".agent")).toBeNull();
  });

  it("shares its contact with the privacy notice", () => {
    expect(CONTACT_URL).toBe(LEGAL_CONTACT_URL);
    expect(CONTACT_TEXT).toBe(LEGAL_CONTACT_EMAIL);
  });

  it("is a public route and is reachable by its hash", async () => {
    expect(parseHash(TERMS_HASH)).toEqual({ name: "terms" });
    session.set(null);
    navigate(TERMS_HASH);
    await settle();
    const { container } = render(App as never, {} as never);
    await settle();
    // No redirect to the sign-in page: the terms are readable first.
    expect(get(route).name).toBe("terms");
    expect(container.textContent).toContain("Terms of service");
  });
});

describe("links to the terms", () => {
  it("the site footer links them", () => {
    const { container } = render(SiteFooter as never, {} as never);
    const link = [...container.querySelectorAll("a")].find(
      (a) => a.getAttribute("href") === TERMS_HASH,
    );
    expect(link).toBeDefined();
    expect(link!.textContent).toContain("Terms");
  });

  it("the legal notice links them, in a new tab at the table", () => {
    let r = render(LegalNotice as never, { newTab: true } as never);
    let a = [...r.container.querySelectorAll("a")].find(
      (x) => x.getAttribute("href") === TERMS_HASH,
    )!;
    expect(a.getAttribute("target")).toBe("_blank");
    r.destroy();

    r = render(LegalNotice as never, {} as never);
    a = [...r.container.querySelectorAll("a")].find((x) => x.getAttribute("href") === TERMS_HASH)!;
    expect(a.getAttribute("target")).toBeNull();
  });
});
