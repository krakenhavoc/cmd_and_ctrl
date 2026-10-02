// Who the client treats as an admin (ADR 0110 §3 item 4).
//
// Two kinds of session are admins on the server:
//   - the shared admin token's session (`role: "admin"`, from
//     POST /admin/login), and
//   - a signed-in person whose Discord ID is on the server's allowlist
//     (CMDCTRL_DISCORD_ADMIN_USER_IDS). Their session is an ordinary
//     identified, player or spectator session; nothing in it says
//     "admin". The server computes that per request, and GET /me
//     reports it as `admin`.
//
// So the client asks /me once per installed session that carries a
// user, and keeps the answer on the session as `admin`. Every admin
// affordance reads isAdmin(session), never `role === "admin"`, so a
// seated admin sees the admin UI exactly as the token does.
//
// This is a mirror of the server, never the gate: every admin route and
// every admin socket binding is decided server-side, and a stale `true`
// here only shows a control whose click is refused.

import { authFetch, session, type Session } from "./session";

/** isAdmin reports whether s is an admin session. */
export function isAdmin(s: Session | null | undefined): boolean {
  if (!s) return false;
  if (s.principal.role === "admin") return true;
  // A session with no user cannot be on the allowlist: the server
  // requires a user (a reclaim ticket carries a seat's Discord ID but
  // no user, and is never admin).
  return Boolean(s.principal.user_id) && s.admin === true;
}

// The tokens already asked about during this page load. A session is
// asked once; a reload asks again, so an ID taken off the list stops
// showing admin UI on the next visit.
const asked = new Set<string>();

/**
 * needsAdminCheck reports whether s should be asked about: a session
 * with a user that has not been asked during this page load. The token
 * session is already known to be admin, and a session with no user
 * never is.
 */
export function needsAdminCheck(s: Session | null | undefined): s is Session {
  if (!s || s.principal.role === "admin") return false;
  if (!s.principal.user_id) return false;
  return !asked.has(s.token);
}

/**
 * loadAdminStatus asks GET /me whether s is an admin and records the
 * answer on the session, if s is still the installed one. A failure
 * leaves the session as it was: a network blip must not take the admin
 * UI away, and an expired session is authFetch's business (it clears
 * it on a 401).
 */
export async function loadAdminStatus(s: Session): Promise<void> {
  asked.add(s.token);
  let admin: boolean;
  try {
    const res = await authFetch("/me");
    const body = (await res.json()) as { admin?: unknown };
    admin = body.admin === true;
  } catch {
    return;
  }
  session.update((cur) => {
    if (!cur || cur.token !== s.token || cur.admin === admin) return cur;
    return { ...cur, admin };
  });
}

/** Test seam: forget which sessions were asked. */
export function resetAdminChecksForTest(): void {
  asked.clear();
}
