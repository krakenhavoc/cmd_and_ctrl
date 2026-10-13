// about.ts — the links the About page (routes/About.svelte, #2981) states,
// in one place.

/** The hash route of the About page. */
export const ABOUT_HASH = "#/about";

/**
 * The project's Discord invite.
 *
 * TODO(owner): no public invite URL is configured anywhere in this
 * repository, so the page links GitHub instead and says nothing about a
 * Discord server. Put the invite here (and the page shows it) once the
 * owner chooses one. Nobody invented one on the owner's behalf.
 */
export const DISCORD_INVITE_URL: string | null = null;

export const REPO_URL = "https://github.com/krakenhavoc/cmd_and_ctrl";
export const ISSUES_URL = `${REPO_URL}/issues`;
/** ADR 0033's 2026-10-12 amendment: hosted model spend is capped, switchable, never sold. */
export const NEVER_PAYWALLED_URL = `${REPO_URL}/blob/develop/docs/decisions/0033-ai-bot-seat.md`;
