// @vitest-environment jsdom
//
// inviteBox.render.test.ts — the front page's "Have an invite?" form
// (#3836): Join stays enabled, an empty submit says what to do and puts
// focus in the field, and the field opts out of autofill and spellcheck.

import { afterEach, describe, expect, it } from "vitest";

import InviteBox from "./components/InviteBox.svelte";
import { cleanup, flushSync, render } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  location.hash = "";
});

function submit(form: HTMLFormElement): void {
  form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
  flushSync();
}

describe("InviteBox", () => {
  it("keeps Join enabled and turns an empty submit into an inline error and focus", () => {
    const { container } = render(InviteBox as never, {} as never);
    const input = container.querySelector<HTMLInputElement>(
      'input[aria-label="invite code or link"]',
    )!;
    const join = container.querySelector<HTMLButtonElement>('button[type="submit"]')!;
    expect(join.disabled).toBe(false);

    submit(input.form!);

    const alert = container.querySelector('[role="alert"]');
    expect(alert?.textContent).toMatch(/paste an invite link or type a code/i);
    expect(document.activeElement).toBe(input);
    expect(join.disabled).toBe(false);
  });

  it("opts the code field out of autofill, autocapitalise and spellcheck", () => {
    const { container } = render(InviteBox as never, {} as never);
    const input = container.querySelector<HTMLInputElement>(
      'input[aria-label="invite code or link"]',
    )!;
    expect(input.getAttribute("name")).toBe("invite");
    expect(input.getAttribute("autocomplete")).toBe("off");
    expect(input.getAttribute("autocapitalize")).toBe("off");
    expect(input.getAttribute("spellcheck")).toBe("false");
  });
});
