// about.ts — the links the About page (routes/About.svelte, #2981) states,
// in one place.

/** The hash route of the About page. */
export const ABOUT_HASH = "#/about";

/** The hash route of the getting-started guide (routes/Guide.svelte). */
export const GUIDE_HASH = "#/guide";

export const REPO_URL = "https://github.com/krakenhavoc/cmd_and_ctrl";
export const ISSUES_URL = `${REPO_URL}/issues`;
/** ADR 0033's 2026-10-12 amendment: hosted model spend is capped, switchable, never sold. */
export const NEVER_PAYWALLED_URL = `${REPO_URL}/blob/develop/docs/decisions/0033-ai-bot-seat.md`;
