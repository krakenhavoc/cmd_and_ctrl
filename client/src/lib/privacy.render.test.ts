// @vitest-environment jsdom
//
// #2983 / ADR 0149: the privacy notice and self-serve account deletion,
// rendered. The page is public and states the numbers lib/privacy.ts
// holds; it is linked from the footer, the sign-in screen and the
// table's legal notice; and "Delete my account" asks for the typed
// phrase, sends it, and signs the browser out.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";
import { get } from "svelte/store";

import App from "../App.svelte";
import DeleteAccount from "./components/DeleteAccount.svelte";
import LegalNotice from "./components/LegalNotice.svelte";
import SiteFooter from "./components/SiteFooter.svelte";
import SiteHeader from "./components/SiteHeader.svelte";
import FrontPage from "../routes/FrontPage.svelte";
import Privacy from "../routes/Privacy.svelte";
import { CONTACT_URL, DELETE_PHRASE, PRIVACY_FACTS, PRIVACY_HASH } from "./privacy";
import { navigate, route } from "./router";
import { session, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

vi.mock("../routes/Game.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));
vi.mock("../routes/Practice.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";

function person(): Session {
  return {
    token: "tok",
    expiresAt: new Date(Date.now() + 3_600_000).toISOString(),
    principal: { role: "identified", user_id: USER, name: "Alice" },
  } as Session;
}

function guest(): Session {
  return {
    token: "tok",
    expiresAt: new Date(Date.now() + 3_600_000).toISOString(),
    principal: {
      role: "player",
      user_id: "00000000-0000-0000-0000-000000000000",
      name: "Guesty",
    },
  } as Session;
}

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

function type(input: HTMLInputElement, text: string): void {
  input.value = text;
  input.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

describe("the privacy page", () => {
  it("renders signed out, with the contact placeholder and the stated numbers", () => {
    const { container } = render(Privacy as never, {} as never);
    const text = container.textContent!.replace(/\s+/g, " ");

    for (const heading of [
      "What is stored, and why",
      "Who it is shared with",
      "Deleting your account",
    ]) {
      expect(text).toContain(heading);
    }
    // The contact is the placeholder, not an invented address.
    expect(text).toContain("Contact: see the project's GitHub");
    const contact = [...container.querySelectorAll("a")].filter(
      (a) => a.getAttribute("href") === CONTACT_URL,
    );
    expect(contact.length).toBeGreaterThan(0);
    expect(text).not.toMatch(/[\w.+-]+@[\w-]+\.[\w.]+/);

    // The numbers come from lib/privacy.ts, which a Go test pins to the server.
    expect(text).toContain(`${PRIVACY_FACTS.endedTableDays} days after it ends`);
    expect(text).toContain(`${PRIVACY_FACTS.bugArtifactDays} days`);
    expect(text).toContain(`${PRIVACY_FACTS.bugArtifactMiB} MiB`);
    expect(text).toContain(`${PRIVACY_FACTS.signInDays} days after a Discord sign-in`);
    expect(text).toContain(`${PRIVACY_FACTS.logDays} days`);

    // Every party it names.
    for (const who of [
      "Discord",
      "GitHub",
      "model provider",
      "Cloudflare",
      "Google Fonts",
      "Scryfall",
    ]) {
      expect(text).toContain(who);
    }
    // A signed-out reader gets the explanation, not a delete button.
    expect(container.querySelector("button.danger")).toBeNull();
    expect(text).toContain("Only a signed-in Discord account can be deleted");
  });

  it("is a public route and is reachable by its hash", async () => {
    session.set(null);
    navigate(PRIVACY_HASH);
    await settle();
    expect(get(route).name).toBe("privacy");
    const { container } = render(App as never, {} as never);
    await settle();
    // No redirect to the sign-in page: the notice is readable first.
    expect(get(route).name).toBe("privacy");
    expect(container.textContent).toContain("What is stored, and why");
  });

  it("offers deletion to a signed-in account, and a guest is told there is nothing to delete", () => {
    session.set(person());
    let r = render(Privacy as never, {} as never);
    expect(r.container.querySelector("button.danger")).not.toBeNull();
    r.destroy();

    session.set(guest());
    r = render(Privacy as never, {} as never);
    expect(r.container.querySelector("button.danger")).toBeNull();
    expect(r.container.textContent).toContain("Only a signed-in Discord account can be deleted");
  });
});

describe("links to the notice", () => {
  it("the site footer links it", () => {
    const { container } = render(SiteFooter as never, {} as never);
    const link = [...container.querySelectorAll("a")].find(
      (a) => a.getAttribute("href") === PRIVACY_HASH,
    );
    expect(link).toBeDefined();
    expect(link!.textContent).toContain("Privacy");
  });

  it("the front page links it", async () => {
    const { container } = render(FrontPage as never, {} as never);
    await settle();
    const links = [...container.querySelectorAll("a")].filter(
      (a) => a.getAttribute("href") === PRIVACY_HASH,
    );
    expect(links.length).toBeGreaterThan(0);
  });

  it("the legal notice opens it in a new tab at the table, in the same tab elsewhere", () => {
    let r = render(LegalNotice as never, { newTab: true } as never);
    let a = [...r.container.querySelectorAll("a")].find(
      (x) => x.getAttribute("href") === PRIVACY_HASH,
    )!;
    expect(a.getAttribute("target")).toBe("_blank");
    expect(a.getAttribute("rel")).toContain("noopener");
    r.destroy();

    r = render(LegalNotice as never, {} as never);
    a = [...r.container.querySelectorAll("a")].find(
      (x) => x.getAttribute("href") === PRIVACY_HASH,
    )!;
    expect(a.getAttribute("target")).toBeNull();
  });
});

describe("Delete my account", () => {
  it("is live only once the phrase is typed, and sends it", async () => {
    session.set(person());
    const calls: Array<{ url: string; init: RequestInit }> = [];
    globalThis.fetch = vi.fn(async (url: string, init: RequestInit) => {
      calls.push({ url, init });
      return new Response(null, { status: 204 });
    }) as never;
    const { container } = render(DeleteAccount as never, {} as never);

    const button = container.querySelector<HTMLButtonElement>("button.danger")!;
    const input = container.querySelector<HTMLInputElement>("#delete-confirm")!;
    expect(button.disabled).toBe(true);
    type(input, "delete");
    expect(button.disabled).toBe(true);
    type(input, DELETE_PHRASE);
    expect(button.disabled).toBe(false);

    click(button);
    await settle();

    expect(calls).toHaveLength(1);
    expect(calls[0].url).toBe("/me/account");
    expect(calls[0].init.method).toBe("DELETE");
    expect(JSON.parse(calls[0].init.body as string)).toEqual({ confirm: DELETE_PHRASE });
    expect((calls[0].init.headers as Record<string, string>).Authorization).toBe("Bearer tok");
    // The browser is signed out and sent to the sign-in page.
    expect(get(session)).toBeNull();
    expect(get(route).name).toBe("login");
  });

  it("shows the server's refusal and keeps the session when a table is still live", async () => {
    session.set(person());
    const message =
      'you still hold a seat at the table "Friday pod", which has not ended. Leave the table or wait for it to end, then delete your account';
    globalThis.fetch = vi.fn(
      async () =>
        new Response(JSON.stringify({ error: message }), {
          status: 409,
          headers: { "Content-Type": "application/json" },
        }),
    ) as never;
    const { container } = render(DeleteAccount as never, {} as never);
    type(container.querySelector<HTMLInputElement>("#delete-confirm")!, DELETE_PHRASE);
    click(container.querySelector<HTMLButtonElement>("button.danger")!);
    await settle();

    expect(container.querySelector('[role="alert"]')?.textContent).toBe(message);
    expect(get(session)).not.toBeNull();
    expect(container.querySelector<HTMLButtonElement>("button.danger")!.disabled).toBe(false);
  });

  it("is in the account menu for a signed-in account and opens the dialog", async () => {
    session.set(person());
    const { container } = render(SiteHeader as never, {} as never);
    click(container.querySelector<HTMLButtonElement>('button[aria-controls="account-menu"]')!);
    await settle();
    const menu = container.querySelector<HTMLElement>("#account-menu")!;
    const items = [...menu.querySelectorAll(".acct-item")].map((e) => e.textContent!.trim());
    expect(items).toContain("Privacy and your data");
    expect(items).toContain("Delete my account");

    click(
      [...menu.querySelectorAll<HTMLElement>(".acct-item")].find(
        (e) => e.textContent!.trim() === "Delete my account",
      )!,
    );
    await settle();
    expect(
      document.querySelector('[role="dialog"][aria-label="Delete my account"]'),
    ).not.toBeNull();
  });

  it("is not offered in the account menu to a guest seat", async () => {
    session.set(guest());
    const { container } = render(SiteHeader as never, {} as never);
    click(container.querySelector<HTMLButtonElement>('button[aria-controls="account-menu"]')!);
    await settle();
    const menu = container.querySelector<HTMLElement>("#account-menu")!;
    const items = [...menu.querySelectorAll(".acct-item")].map((e) => e.textContent!.trim());
    expect(items).toContain("Privacy and your data");
    expect(items).not.toContain("Delete my account");
  });
});
