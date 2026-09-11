// Inline SVG keyword icons for the S18 combat-keyword badge row.
// All icons are 24x24 currentColor so the badge background colors
// them via CSS. Keys match the server's canonical lowercase tokens
// (space-delimited for multi-word keywords — the wire format).
// Unknown tokens fall back to the 3-letter text badge in
// KeywordBadgeRow.

const flying = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M4 14 Q 14 3 20 6"/><path d="M20 6 Q 18 9 15 9 Q 12 11 10 11 Q 7 13 4 14"/><path d="M9 11 L 11 7"/><path d="M13 10 L 15 6"/></svg>`;

const reach = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22 L 12 11"/><path d="M12 15 Q 8 14 7 11"/><path d="M12 13 Q 16 12 17 9"/><path d="M12 11 Q 10 7 9 5"/><path d="M12 11 Q 14 7 15 5"/></svg>`;

const firstStrike = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 L 8 16"/><path d="M5 13 L 11 19"/><path d="M8 16 L 5 19"/><circle cx="4" cy="20" r="1.1" fill="currentColor" stroke="none"/></svg>`;

const doubleStrike = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M6 18 L 20 4"/><path d="M4 14 L 8 18"/><path d="M18 18 L 4 4"/><path d="M20 14 L 16 18"/></svg>`;

const deathtouch = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path fill-rule="evenodd" clip-rule="evenodd" d="M12 3a8 8 0 0 0-6 13.3V20a1 1 0 0 0 1 1h1.5v-2h1.5v2h4v-2h1.5v2H17a1 1 0 0 0 1-1v-3.7A8 8 0 0 0 12 3ZM7 11a2 2 0 1 0 4 0 2 2 0 0 0-4 0ZM13 11a2 2 0 1 0 4 0 2 2 0 0 0-4 0Z"/></svg>`;

const lifelink = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20 C 4 14 3 9 6 6 C 9 3 12 5 12 8 C 12 5 15 3 18 6 C 21 9 20 14 12 20 Z"/><path d="M12 10 L 12 14"/><path d="M10 12 L 14 12"/></svg>`;

const trample = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M10 6 C 7 7 7 12 9 14 C 11 14 11.5 12 11 10 Z"/><path d="M14 6 C 17 7 17 12 15 14 C 13 14 12.5 12 13 10 Z"/><path d="M4 14 L 6 14"/><path d="M18 14 L 20 14"/><path d="M12 18 L 12 20"/><path d="M6 18 L 7.5 17"/><path d="M18 18 L 16.5 17"/></svg>`;

const vigilance = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12 C 5 6 9 5 12 5 C 15 5 19 6 22 12 C 19 18 15 19 12 19 C 9 19 5 18 2 12 Z"/><circle cx="12" cy="12" r="3"/></svg>`;

const menace = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path d="M5 5 L 10 5 C 10 9 8.5 14 7 19 C 6 15 5 10 5 5 Z"/><path d="M14 5 L 19 5 C 19 10 18 15 17 19 C 15.5 14 14 9 14 5 Z"/></svg>`;

const defender = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path d="M12 3 L 5 5 L 5 11 C 5 16 8 20 12 22 C 16 20 19 16 19 11 L 19 5 Z"/></svg>`;

const haste = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M10 5 L 18 12 L 10 19"/><path d="M3 8 L 6 8"/><path d="M2 12 L 6 12"/><path d="M3 16 L 6 16"/></svg>`;

const flash = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2 L 12 8"/><path d="M12 16 L 12 22"/><path d="M2 12 L 8 12"/><path d="M16 12 L 22 12"/><path d="M6 6 L 8.5 8.5"/><path d="M18 18 L 15.5 15.5"/><path d="M18 6 L 15.5 8.5"/><path d="M6 18 L 8.5 15.5"/><circle cx="12" cy="12" r="1.1" fill="currentColor" stroke="none"/></svg>`;

// S23 targeting pair. Hexproof is a shield with a crossed-out
// targeting reticle — "your opponents can't point at this"; shroud
// is the same reticle behind a full curtain, because shroud stops
// everyone including you.
const hexproof = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3 L 5 5.5 L 5 11 C 5 15.5 8 19.5 12 21 C 16 19.5 19 15.5 19 11 L 19 5.5 Z"/><circle cx="12" cy="11.5" r="3"/><path d="M9.2 8.7 L 14.8 14.3"/></svg>`;

const shroud = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4 C 8 7 16 7 20 4 L 20 15 C 20 19 16.5 21 12 21 C 7.5 21 4 19 4 15 Z"/><path d="M8 11 L 16 11"/><path d="M8.5 15 L 15.5 15"/></svg>`;

export const KEYWORD_ICONS: Record<string, string> = {
  flying,
  reach,
  "first strike": firstStrike,
  "double strike": doubleStrike,
  deathtouch,
  lifelink,
  trample,
  vigilance,
  menace,
  defender,
  haste,
  flash,
  hexproof,
  shroud,
};
