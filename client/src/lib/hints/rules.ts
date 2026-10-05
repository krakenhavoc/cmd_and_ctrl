// rules.ts — the rules every hint is held to (ADR 0125 §3.2–3.4).
//
// hints.test.ts runs these over every collected hint, so a hint that
// breaks one fails the client job on its PR. They are functions rather
// than inline assertions so the test can also prove each rule catches
// what it should, on hints made to break it.

import { copyText, type Copy, type CopyContext } from "../tutorial";
import { HINT_PLACES, type Hint } from "./hint";

/** A title of at most this many characters (§3.4). */
export const TITLE_MAX = 32;
/** A body of at most this many characters (§3.4). */
export const BODY_MAX = 140;

/**
 * Words a hint never uses (§3.4): engine and project vocabulary. A hint
 * says where something is and how to use it, in the player's words.
 * "database" is not here: the admin hint may say it, its readers run it.
 */
export const BANNED_WORDS: readonly string[] = [
  "engine",
  "server",
  "client",
  "snapshot",
  "seam",
  "oracle",
  "payload",
  "websocket",
];

/**
 * The key bindings copy is checked under: the shipped defaults, and a
 * set of long rebinds, so a function of the bindings stays inside the
 * limits whatever the player chose.
 */
export const COPY_CONTEXTS: readonly CopyContext[] = [
  { helpKey: "?", settingsKey: ",", nextKey: "Space" },
  { helpKey: "Ctrl+Shift+F12", settingsKey: "Ctrl+Shift+F11", nextKey: "Ctrl+Shift+Enter" },
  { helpKey: "", settingsKey: "", nextKey: "" },
];

const ID = new RegExp(`^(${HINT_PLACES.join("|")})\\.[a-z0-9]+(?:-[a-z0-9]+)*$`);

function words(text: string): string[] {
  return text.toLowerCase().split(/[^a-z0-9]+/);
}

/** copyProblems lists what is wrong with one piece of copy. */
function copyProblems(id: string, part: "title" | "body", copy: Copy, max: number): string[] {
  const out: string[] = [];
  for (const ctx of COPY_CONTEXTS) {
    const text = copyText(copy, ctx);
    const keys = `keys ${JSON.stringify(ctx)}`;
    if (text.trim() === "") out.push(`${id}: its ${part} is empty (${keys})`);
    if (text.length > max) {
      out.push(`${id}: its ${part} is ${text.length} characters, over ${max} (${keys}): "${text}"`);
    }
    const banned = words(text).filter((w) => BANNED_WORDS.includes(w));
    for (const w of new Set(banned)) out.push(`${id}: its ${part} says "${w}" (${keys})`);
    if (part === "body" && !/[.]$/.test(text.trim())) {
      out.push(`${id}: its body does not end in a full stop (${keys}): "${text}"`);
    }
  }
  return [...new Set(out)];
}

/**
 * hintProblems lists every rule one hint breaks: its id's form and place,
 * its version, its copy, and its one optional action.
 */
export function hintProblems(h: Hint): string[] {
  const out: string[] = [];
  if (!ID.test(h.id)) {
    out.push(`${h.id}: an id is "<place>.<feature>", lower-case words joined by hyphens`);
  }
  if (!(HINT_PLACES as readonly string[]).includes(h.place)) {
    out.push(`${h.id}: "${h.place}" is not a known place (${HINT_PLACES.join(", ")})`);
  } else if (!h.id.startsWith(`${h.place}.`)) {
    out.push(`${h.id}: its id must start with its place, "${h.place}."`);
  }
  if (!Number.isInteger(h.version) || h.version < 1) {
    out.push(`${h.id}: its version must be a positive integer, not ${String(h.version)}`);
  }
  if (!Number.isFinite(h.order)) out.push(`${h.id}: its order must be a number`);
  out.push(...copyProblems(h.id, "title", h.title, TITLE_MAX));
  out.push(...copyProblems(h.id, "body", h.body, BODY_MAX));
  if (h.action) {
    if (h.action.label.trim() === "") out.push(`${h.id}: its action has no label`);
    if (!h.action.href.startsWith("#/")) {
      out.push(`${h.id}: its action must link to a page of the site ("#/…")`);
    }
  }
  return out;
}

/**
 * collectionProblems lists what is wrong with a set of hints together:
 * an id used twice, or a retired id used again.
 */
export function collectionProblems(hints: readonly Hint[], retired: readonly string[]): string[] {
  const out: string[] = [];
  const seen = new Set<string>();
  for (const h of hints) {
    if (seen.has(h.id)) out.push(`${h.id}: two hints have this id`);
    seen.add(h.id);
    if (retired.includes(h.id)) {
      out.push(
        `${h.id}: this id is retired (lib/hints/retired.ts) and may never be used again; ` +
          "someone who saw the old hint would never see this one",
      );
    }
  }
  return out;
}
