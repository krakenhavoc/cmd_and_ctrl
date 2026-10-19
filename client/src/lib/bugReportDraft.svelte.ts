// bugReportDraft — the bug report in progress (#4029).
//
// BugReportModal used to keep its fields in local state, so the only way
// out was Cancel, which threw the title, the details and every
// screenshot away. A reporter who wants a second screenshot has to leave
// the form and use the page, so the fields live here instead: the modal
// can be unmounted (minimised to a pill) and mounted again with
// everything where they left it.
//
// One draft per tab, in memory only. It is module state rather than a
// prop or a context because two different screens open the form (the
// site header and the table) and a draft started on one must be
// restorable from the other after a navigation. It does not survive a
// reload.
//
// A minimised draft is simply a draft with no modal mounted. That is the
// whole trick: nothing is hidden, so the modal layer and the paste/drop
// listeners are gone with the component, and the page underneath is
// untouched.
//
// The thumbnails' object URLs belong to the draft, not to the modal,
// for the same reason: they must outlive an unmount, and a leaked blob
// URL pins its whole image in memory for the life of the tab. They are
// revoked when the draft is discarded (Cancel) or sent, and nowhere else.

import {
  DEFAULT_BUG_KIND,
  type BugLogEntry,
  type BugReportContext,
  type BugReportKind,
} from "./bugReport";

/** What a draft is born with, taken once when the form first opens. */
export interface BugDraftOrigin {
  // The client log, snapshotted so the log the reporter reviews is the
  // log that is sent, and a restore still sends the log from the moment
  // the bug happened, not from the moment they came back to the form.
  capturedLog: BugLogEntry[];
  // The game coordinates (or page) the report was opened from, for the
  // same reason: by restore time the game has moved on.
  context: BugReportContext;
}

export class BugReportDraft {
  kind = $state<BugReportKind>(DEFAULT_BUG_KIND);
  title = $state("");
  description = $state("");
  includeContext = $state(true);
  includeLog = $state(true);
  showLog = $state(false);

  // Raw, not deep: Files are opaque and the arrays are only ever replaced.
  images = $state.raw<File[]>([]);
  // #4116: the picture of the board taken with the form's checkbox. It is
  // also in `images`; this names which one, so unticking removes it and
  // removing it from the list unticks the box.
  boardShot = $state.raw<File | null>(null);
  // One object URL per image, index-aligned with `images`.
  previews = $state.raw<string[]>([]);

  readonly capturedLog: BugLogEntry[];
  readonly context: BugReportContext;

  // The URL each live image owns, so adding or removing one neither
  // revokes nor re-creates the others' thumbnails.
  private urls = new Map<File, string>();

  constructor(origin: BugDraftOrigin) {
    this.capturedLog = origin.capturedLog;
    this.context = origin.context;
  }

  /** setImages replaces the attached set, minting and revoking URLs to match. */
  setImages(next: File[]): void {
    const keep = new Set(next);
    for (const [file, url] of this.urls) {
      if (keep.has(file)) continue;
      URL.revokeObjectURL(url);
      this.urls.delete(file);
    }
    for (const file of next) {
      if (!this.urls.has(file)) this.urls.set(file, URL.createObjectURL(file));
    }
    this.images = next;
    this.previews = next.map((f) => this.urls.get(f)!);
  }

  /** dispose revokes every object URL. Called by discardBugDraft only. */
  dispose(): void {
    for (const url of this.urls.values()) URL.revokeObjectURL(url);
    this.urls.clear();
    this.previews = [];
    this.images = [];
    this.boardShot = null;
  }
}

let current = $state.raw<BugReportDraft | null>(null);

/** currentBugDraft is the draft in progress, or null. Reactive. */
export function currentBugDraft(): BugReportDraft | null {
  return current;
}

/**
 * openBugDraft returns the draft in progress, starting one from `origin`
 * when there is none. `restored` says which, so the form can put focus
 * back where a returning reporter wants it.
 */
export function openBugDraft(origin: () => BugDraftOrigin): {
  draft: BugReportDraft;
  restored: boolean;
} {
  if (current) return { draft: current, restored: true };
  current = new BugReportDraft(origin());
  return { draft: current, restored: false };
}

/** discardBugDraft drops the draft and frees its images: Cancel, or a successful send. */
export function discardBugDraft(): void {
  current?.dispose();
  current = null;
}

/** _resetForTests is the vitest teardown hook. Not for prod use. */
export function _resetForTests(): void {
  discardBugDraft();
}
