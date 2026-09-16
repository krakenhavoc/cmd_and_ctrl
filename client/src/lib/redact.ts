// redact.ts — strips credentials out of text before it can reach a
// bug report (#721).
//
// The in-app report inlines the client log in a GitHub issue body
// (ADR 0017 §7). That log once carried the WebSocket URL verbatim,
// `?token=` and all, so every report with a connect line published
// the reporter's session token. Durable sessions (#517) would make
// such a token outlive a server restart, which is why this landed
// first.
//
// Two layers use this module:
//
//   - at the source: ws.ts logs redactURL(url), never the raw URL;
//   - as a backstop: the ws log buffer, the console/error capture, the
//     merged bug-report log and the submitted draft all pass through
//     redactSecrets, so a log line added tomorrow can't reintroduce
//     the leak.
//
// The server applies the same rules again before rendering the issue
// (server/internal/util/redact), so a stale client can't leak either.
// Keep the two rule sets in step.
//
// Pure and dependency-free: vitest covers it directly.

export const REDACTED = "REDACTED";

// SECRET_KEY matches a key that names a credential anywhere it appears
// as key=value: `token`, `access_token`, `refresh_token`, `id_token`,
// `client_secret`, `password`, the reclaim `ticket`, … A suffix match
// rather than a list, so `csrf_token` or `session_token` are covered
// without anyone remembering to add them.
const SECRET_KEY = String.raw`[a-z0-9_-]*(?:token|secret|password|passwd|ticket)`;

// QUERY_ONLY_KEY is a credential key too short to trust in free text.
// `t` is the invite and reclaim token (`#/games/<id>/join?t=…`); it is
// only redacted when it is unambiguously a query parameter, so a log
// line saying "t=3" about a turn is left alone.
const QUERY_ONLY_KEY = "t";

// VALUE is one key's value: everything up to the next parameter or
// fragment separator, whitespace, quote or bracket. Percent-escapes are
// consumed as part of the value EXCEPT %23 (#) and %26 (&), which end
// it — that is what keeps a URL-encoded URL (`…%3Ftoken%3Dabc%26game%3D1`)
// from losing its trailing parameters to the redaction.
const VALUE = String.raw`(?:[^&#\s"'<>()\[\]{},;%\\]|%(?:[013-9a-f][0-9a-f]|2[0-2457-9a-f]))+`;

// EQ is a literal or percent-encoded equals sign.
const EQ = "(=|%3d)";

// Key=value anywhere: the key must not be glued to a preceding word
// character, so "bearer_tokens=3" style prose is the worst false
// positive and "notatoken=x" is not one.
const KEY_VALUE_RE = new RegExp(String.raw`(^|[^a-z0-9_-])(${SECRET_KEY})${EQ}(${VALUE})`, "gi");

// `?t=` / `&t=` (and their encoded forms `%3Ft=`, `%26t%3D`, and the
// HTML-escaped `&amp;t=`).
const QUERY_ONLY_RE = new RegExp(
  String.raw`(\?|&amp;|&|%3f|%26)(${QUERY_ONLY_KEY})${EQ}(${VALUE})`,
  "gi",
);

// JSON-shaped: "token":"abc", including the escaped form a stringified
// object picks up inside another string (\"token\":\"abc\").
const JSON_RE = new RegExp(
  String.raw`(\\?")(${SECRET_KEY})(\\?"\s*:\s*\\?")((?:[^"\\]|\\(?!"))*)(\\?")`,
  "gi",
);

// `Authorization: Bearer <credential>`, `Bearer%20<credential>`. Case
// sensitive and length-gated on purpose: "Bearer of the Heavens" is a
// card, and a real credential is never four letters long. The guard
// before the word is "not a letter" rather than \b so the encoded
// `%20Bearer%20…` form (where a digit touches the B) still matches.
const BEARER_RE = /(^|[^A-Za-z])(Bearer)(\s+|%20|\+)([A-Za-z0-9._~+/=-]{16,})/g;

// redactSecrets replaces every credential value it recognises in s with
// REDACTED, keeping the key so a triager can still see that one was
// sent. Idempotent: redacting a redacted string changes nothing.
export function redactSecrets(s: string): string {
  if (!s) return s;
  return s
    .replace(
      KEY_VALUE_RE,
      (_m, pre: string, key: string, eq: string) => `${pre}${key}${eq}${REDACTED}`,
    )
    .replace(
      QUERY_ONLY_RE,
      (_m, pre: string, key: string, eq: string) => `${pre}${key}${eq}${REDACTED}`,
    )
    .replace(
      JSON_RE,
      (_m, open: string, key: string, mid: string, _v: string, close: string) =>
        `${open}${key}${mid}${REDACTED}${close}`,
    )
    .replace(
      BEARER_RE,
      (_m, pre: string, word: string, sep: string) => `${pre}${word}${sep}${REDACTED}`,
    );
}

// redactURL renders a URL safe to log: scheme, host, path, fragment and
// the non-secret parameters survive (`game` and `player` are what a
// triager needs to find the table); credential values become REDACTED.
//
// Deliberately string-based rather than URL/URLSearchParams: those
// re-encode the parameters, so "REDACTED" would be the only thing in
// the line that looked like the original, and a relative or malformed
// URL would throw where a log line must never throw.
export function redactURL(url: string): string {
  return redactSecrets(url);
}
