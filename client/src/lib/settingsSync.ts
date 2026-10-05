// settingsSync.ts — a signed-in person's settings, on their account
// (ADR 0110 §4, Delivery PR 5).
//
// The per-person fields (SYNCED_FIELDS in settings.ts) are kept on the
// account through GET and PUT /me/settings. The per-device ones never
// leave the browser. A guest, the admin token, and every session on a
// server with no database get a 403 from both routes, and then this
// module does nothing: their settings stay in the browser, as before.
//
// # When it talks to the server
//
//   - A download when a session with a user is installed: at sign-in,
//     and at page load while signed in. A token swap for the SAME user
//     (a seat claimed, the practice table, a renewal) is not a sign-in
//     and downloads nothing.
//   - A debounced upload (DEBOUNCE_MS) after a change to a synced field.
//   - A flush on pagehide, so a change made just before closing the tab
//     is not lost to the debounce.
//
// # Who wins (owner answer 6)
//
// At sign-in the ACCOUNT's copy wins. If there is none, this browser's
// values become the first copy. If there is one and it differs from
// this browser's, it is applied and `settingsSyncToast` offers "Keep
// this browser's instead" for the rest of the visit, which puts the
// browser's values back and uploads them. Signing out leaves the
// browser holding whatever it had; nothing reverts.
//
// A 412 (another device wrote first) is merged field by field against
// the copy both sides last agreed on: a field only this browser changed
// keeps this browser's value, and every other field takes the
// account's. So on a field both changed, the account wins again. The
// merged result is then uploaded over the new revision.
//
// # `help.seen`, the one field that merges by union (ADR 0125 §4)
//
// The hints a person has dismissed are a set that only grows, so losing
// either side's entries is the only way a merge can go wrong. At
// sign-in, `help.seen` is the UNION of this browser's map and the
// account's (the higher version for an id in both); on a 412 it is the
// union of the two copies instead of field-wins. A difference in
// `help.seen` alone never raises the "keep this browser's" toast: a
// union cannot lose anything either side had. Every other field,
// `help.tipsOff` included, follows the rules above.
//
// A copy written by a NEWER client (a version above SETTINGS_VERSION,
// or a 409) is applied through migrate, and then this tab stops writing
// until the page reloads: a stale tab must not stamp an older schema
// over a newer client's copy.
//
// # The practice table (ADR 0076 §2.2, ADR 0110 §4 item 5)
//
// The tutorial forces four settings through settings.update. While a
// practice record exists, every upload is of
// `withSettings(current, record.saved)` — the settings as they will be
// after the restore — so the forced values never reach the account. A
// copy that arrives during the tutorial updates the record's saved
// values for the forced fields and the live settings for the rest.
//
// # Failure
//
// Nothing here blocks the UI or throws into it. A failed upload or
// download retries with a backoff (RETRY_MIN_MS up to RETRY_MAX_MS),
// and at once when the browser comes back online.

import { get, type Readable } from "svelte/store";
import { fetchAccountSettings, putAccountSettings, type AccountSettings } from "./api";
import { guardedWritable } from "./guardedStore";
import { signedInUserID } from "./myGames";
import { captureSettings, practiceSaved, updatePracticeSaved, withSettings } from "./practiceTable";
import { LobbyApiError, session, type Session } from "./session";
import { normalizeSeen, unionSeen, type SeenMap } from "./hints/seen";
import {
  applySyncedCopy,
  canonicalJSON,
  settings,
  syncedPaths,
  syncedSubset,
  SETTINGS_VERSION,
  type Settings,
  type SyncedSettings,
} from "./settings";

export const DEBOUNCE_MS = 1_000;
export const RETRY_MIN_MS = 5_000;
export const RETRY_MAX_MS = 5 * 60_000;

/** What the "keep this browser's" toast holds: the values it would restore. */
export interface SettingsSyncToast {
  browser: SyncedSettings;
}

const toastStore = guardedWritable<SettingsSyncToast | null>(null, "settingsSyncToast");

/**
 * settingsSyncToast is non-null for the rest of the visit after a
 * sign-in applied an account copy that differed from this browser's.
 */
export const settingsSyncToast: Readable<SettingsSyncToast | null> = toastStore;

interface SyncState {
  /** The user being synced; null when there is nobody to sync for. */
  user: string | null;
  /** The account revision this browser last agreed with; null before the first download. */
  revision: number | null;
  /** The synced subset at that revision: the base of a 412 merge. */
  base: SyncedSettings | null;
  /** 403: this session is browser-only. */
  disabled: boolean;
  /** A newer client owns the copy: no writes until reload. */
  blocked: boolean;
}

function freshState(user: string | null): SyncState {
  return { user, revision: null, base: null, disabled: false, blocked: false };
}

let state: SyncState = freshState(null);
// Bumped on every change of user, so a reply for the previous one is
// dropped on arrival.
let generation = 0;
let applying = false;
let inflight = false;
let again = false;
let debounce: ReturnType<typeof setTimeout> | null = null;
let retry: ReturnType<typeof setTimeout> | null = null;
let retryDelay = RETRY_MIN_MS;
let retryFn: (() => void) | null = null;

/** uploadable is the settings as the account should see them. */
function uploadable(): Settings {
  const cur = get(settings);
  const saved = practiceSaved();
  return saved === null ? cur : withSettings(cur, saved);
}

/** applyCopy puts an account copy into the live settings, through the practice rule. */
function applyCopy(copy: unknown, version: number): void {
  const cur = get(settings);
  const saved = practiceSaved();
  const restored = saved === null ? cur : withSettings(cur, saved);
  const next = applySyncedCopy(restored, copy, version);
  let live = next;
  if (saved !== null) {
    updatePracticeSaved(captureSettings(next));
    live = withSettings(next, captureSettings(cur));
  }
  applying = true;
  try {
    settings.set(live);
  } finally {
    applying = false;
  }
}

function canWrite(): boolean {
  return state.user !== null && state.revision !== null && !state.disabled && !state.blocked;
}

function clearTimers(): void {
  if (debounce !== null) clearTimeout(debounce);
  if (retry !== null) clearTimeout(retry);
  debounce = null;
  retry = null;
  retryFn = null;
  retryDelay = RETRY_MIN_MS;
}

function scheduleRetry(fn: () => void): void {
  if (retry !== null) clearTimeout(retry);
  retryFn = fn;
  retry = setTimeout(() => {
    retry = null;
    retryFn = null;
    fn();
  }, retryDelay);
  retryDelay = Math.min(retryDelay * 2, RETRY_MAX_MS);
}

function onOnline(): void {
  const fn = retryFn;
  if (fn === null) return;
  if (retry !== null) clearTimeout(retry);
  retry = null;
  retryFn = null;
  fn();
}

/** isStatus reports whether err is an API refusal with one of these statuses. */
function isStatus(err: unknown, ...codes: number[]): err is LobbyApiError {
  return err instanceof LobbyApiError && codes.includes(err.status);
}

function copyOf(err: LobbyApiError): AccountSettings | null {
  const b = err.body as Partial<AccountSettings> | undefined;
  if (!b || typeof b.revision !== "number") return null;
  return b as AccountSettings;
}

// --- help.seen, merged by union (ADR 0125 §4) ---------------------------

/** seenOf is a synced subset's `help.seen`, checked. */
function seenOf(sub: SyncedSettings | null | undefined): SeenMap {
  return normalizeSeen(sub?.help?.seen);
}

/** withSeenMap is `sub` with its `help.seen` replaced. */
function withSeenMap(sub: SyncedSettings, seen: SeenMap): SyncedSettings {
  return { ...sub, help: { ...(sub.help ?? {}), seen } };
}

/** sansSeen is `sub` without `help.seen`, for deciding whether to raise the toast. */
function sansSeen(sub: SyncedSettings | null | undefined): SyncedSettings {
  const out: SyncedSettings = { ...(sub ?? {}) };
  if (out.help) {
    const help = { ...out.help };
    delete help.seen;
    out.help = help;
  }
  return out;
}

/** differsBeyondSeen reports whether two copies differ in anything but `help.seen`. */
function differsBeyondSeen(a: SyncedSettings | null, b: SyncedSettings | null): boolean {
  return canonicalJSON(sansSeen(a)) !== canonicalJSON(sansSeen(b));
}

// --- download -----------------------------------------------------------

async function download(gen: number): Promise<void> {
  let copy: AccountSettings;
  try {
    copy = await fetchAccountSettings();
  } catch (err) {
    if (gen !== generation) return;
    if (isStatus(err, 403)) {
      state.disabled = true;
      return;
    }
    // 401: authFetch has already signed this tab out.
    if (isStatus(err, 401)) return;
    scheduleRetry(() => void download(gen));
    return;
  }
  if (gen !== generation) return;
  retryDelay = RETRY_MIN_MS;
  atSignIn(copy);
}

/**
 * atSignIn is owner answer 6: the account's copy wins, except for
 * `help.seen`, which takes the union of both (ADR 0125 §4).
 */
function atSignIn(copy: AccountSettings): void {
  if (copy.revision === 0 || !copy.settings) {
    // No account copy: this browser's values become the first one.
    state.revision = 0;
    state.base = null;
    void upload();
    return;
  }
  const browser = syncedSubset(uploadable());
  const version = copy.version ?? SETTINGS_VERSION;
  // The account's copy in this client's shape: what both sides agree on.
  const account = syncedSubset(applySyncedCopy(uploadable(), copy.settings, version));
  const seen = unionSeen(seenOf(browser), seenOf(account));
  applyCopy(withSeenMap(account, seen), SETTINGS_VERSION);
  state.revision = copy.revision;
  state.base = account;
  if (version > SETTINGS_VERSION) state.blocked = true;
  if (!state.blocked && differsBeyondSeen(browser, account)) {
    toastStore.set({ browser });
  }
  // Hints dismissed in this browser and not yet on the account go up.
  schedule();
}

// --- upload -------------------------------------------------------------

function schedule(): void {
  if (!canWrite()) return;
  if (canonicalJSON(syncedSubset(uploadable())) === canonicalJSON(state.base)) return;
  if (debounce !== null) clearTimeout(debounce);
  debounce = setTimeout(() => {
    debounce = null;
    void upload();
  }, DEBOUNCE_MS);
}

async function upload(opts: { keepalive?: boolean } = {}): Promise<void> {
  if (!canWrite()) return;
  if (inflight) {
    again = true;
    return;
  }
  const subset = syncedSubset(uploadable());
  if (state.base !== null && canonicalJSON(subset) === canonicalJSON(state.base)) return;
  const gen = generation;
  inflight = true;
  try {
    const saved = await putAccountSettings(state.revision ?? 0, SETTINGS_VERSION, subset, opts);
    if (gen !== generation) return;
    retryDelay = RETRY_MIN_MS;
    state.revision = saved.revision;
    state.base = subset;
  } catch (err) {
    if (gen !== generation) return;
    if (isStatus(err, 412) && copyOf(err)) {
      merge(copyOf(err)!);
    } else if (isStatus(err, 409)) {
      // A newer client saved the copy. Take it, and stop writing.
      const copy = copyOf(err);
      if (copy?.settings) applyCopy(copy.settings, copy.version ?? SETTINGS_VERSION);
      state.blocked = true;
    } else if (isStatus(err, 403)) {
      state.disabled = true;
    } else if (!isStatus(err, 401, 400, 413)) {
      // A network failure, a 5xx or a 429: try again later. A 400 or
      // 413 would fail the same way again, so it is not retried.
      scheduleRetry(() => void upload());
    }
  } finally {
    inflight = false;
    if (again) {
      again = false;
      schedule();
    }
  }
}

/**
 * merge resolves a 412: the account moved since this browser last
 * agreed with it. Field by field against that agreed copy, a field only
 * this browser changed keeps this browser's value; everything else
 * takes the account's.
 */
function merge(copy: AccountSettings): void {
  const local = syncedSubset(uploadable());
  const base = state.base;
  const version = copy.version ?? SETTINGS_VERSION;
  const remote = syncedSubset(applySyncedCopy(uploadable(), copy.settings ?? {}, version));
  const merged: SyncedSettings = {};
  for (const [group, key] of syncedPaths()) {
    if (group === "help" && key === "seen") {
      // ADR 0125 §4: a union, not field-wins.
      (merged[group] ??= {})[key] = unionSeen(seenOf(local), seenOf(remote));
      continue;
    }
    const l = local[group]?.[key];
    const r = remote[group]?.[key];
    const b = base?.[group]?.[key];
    const onlyLocalChanged =
      base !== null &&
      canonicalJSON(l) !== canonicalJSON(b) &&
      canonicalJSON(r) === canonicalJSON(b);
    (merged[group] ??= {})[key] = onlyLocalChanged ? l : r;
  }
  // With no agreed copy (this browser's first upload lost a race with
  // another device's) this is a sign-in after all: the account wins
  // outright, and the toast offers this browser's values back.
  if (base === null && version <= SETTINGS_VERSION && differsBeyondSeen(local, remote)) {
    toastStore.set({ browser: local });
  }
  applyCopy(merged, SETTINGS_VERSION);
  state.revision = copy.revision;
  state.base = remote;
  if (version > SETTINGS_VERSION) state.blocked = true;
  schedule();
}

// --- the toast's action ---------------------------------------------------

/**
 * keepBrowserSettings is the toast's "Keep this browser's instead": put
 * this browser's values back and upload them over the account's.
 */
export function keepBrowserSettings(): void {
  const t = get(toastStore);
  if (t === null) return;
  toastStore.set(null);
  // Every other field goes back to this browser's; the seen hints keep
  // the union, so nothing dismissed on either side is offered again.
  const seen = unionSeen(seenOf(t.browser), normalizeSeen(get(settings).help.seen));
  applyCopy(withSeenMap(t.browser, seen), SETTINGS_VERSION);
  if (debounce !== null) clearTimeout(debounce);
  debounce = null;
  void upload();
}

/** dismissSettingsSyncToast hides the toast and keeps the account's values. */
export function dismissSettingsSyncToast(): void {
  toastStore.set(null);
}

// --- wiring -------------------------------------------------------------

function onSession(s: Session | null): void {
  const user = signedInUserID(s);
  if (user === state.user) return;
  generation++;
  clearTimers();
  inflight = false;
  again = false;
  toastStore.set(null);
  state = freshState(user);
  if (user !== null) void download(generation);
}

let installed = false;
let teardown: Array<() => void> = [];

/**
 * installSettingsSync arms the sync: it follows the session and the
 * settings store from here on. Idempotent; call once at boot, after
 * recoverPractice (so a crashed tutorial's forced values are already
 * restored). Returns a function that disarms it, for tests.
 */
export function installSettingsSync(): () => void {
  if (installed) return uninstall;
  installed = true;
  teardown.push(session.subscribe(onSession));
  teardown.push(
    settings.subscribe(() => {
      if (!applying) schedule();
    }),
  );
  if (typeof window !== "undefined") {
    const onPageHide = () => {
      if (debounce === null) return;
      clearTimeout(debounce);
      debounce = null;
      void upload({ keepalive: true });
    };
    window.addEventListener("pagehide", onPageHide);
    window.addEventListener("online", onOnline);
    teardown.push(() => window.removeEventListener("pagehide", onPageHide));
    teardown.push(() => window.removeEventListener("online", onOnline));
  }
  return uninstall;
}

function uninstall(): void {
  for (const fn of teardown) fn();
  teardown = [];
  installed = false;
  generation++;
  clearTimers();
  state = freshState(null);
  toastStore.set(null);
}
