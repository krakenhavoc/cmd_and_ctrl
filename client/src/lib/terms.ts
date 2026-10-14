// terms.ts — what the terms of service (routes/Terms.svelte, #3358) state,
// in one place. The operator and its contact are in lib/legalNotice.ts,
// shared with the privacy notice.

/** The hash route of the terms. */
export const TERMS_HASH = "#/terms";

/** When the terms last changed. Bump it with any change to what they say. */
export const TERMS_UPDATED = "2026-10-14";

/** The youngest a person may be to use the site: Discord's own floor. */
export const MIN_AGE = 13;

/** The cap on the operator's liability, in US dollars (the "Limits" section). */
export const LIABILITY_CAP_USD = 50;

/**
 * The copyright agent's postal address and phone number. US law (17 U.S.C.
 * §512(c)(2)) asks that they be shown on the site once the agent is
 * registered with the US Copyright Office.
 *
 * TODO(owner): copyright agent. Fill in both after registering KrakenKey
 * LLC's agent at dmca.copyright.gov. Until then the page gives only the
 * email address. Nobody invented an address or number on the owner's
 * behalf.
 */
export const COPYRIGHT_AGENT_POSTAL = "";
export const COPYRIGHT_AGENT_PHONE = "";
