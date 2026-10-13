// gateRefusal.ts — what the login page says when the sign-up gate turned
// someone away (ADR 0156 §1 item 4).
//
// The server's Discord callback sends a refused sign-in to
// #/request-access, and one Discord did not answer to
// #/request-access?retry=1. The gate page itself (ADR 0156 §5) is a later
// slice (#3182); until it ships the router does not know that route and
// falls back to the login page. This keeps that gap from looking like
// nothing happened: the login page shows one line saying why. #3182
// replaces it with the real page and can delete this file.

/** The notice for a hash the gate redirected to, or "" for any other. */
export function gateRefusalNotice(hash: string): string {
  const raw = hash.replace(/^#\/?/, "");
  const [path, query = ""] = raw.split("?");
  if (path.replace(/\/+$/, "") !== "request-access") return "";
  if (new URLSearchParams(query).get("retry") === "1") {
    return "Discord didn't answer when we checked your account. Try signing in again.";
  }
  return "Accounts are for members of the C2 Discord server who have been given access. Ask in the server, then sign in again. An invite link from a host still works without an account.";
}
