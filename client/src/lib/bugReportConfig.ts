// bugReportConfig — GET /bugreport/config, asked once per tab (#3492).
//
// The site header offers Report a bug on every page outside the table,
// and Page.svelte mounts a fresh header on each navigation, so the
// probe is shared rather than repeated. An answer that says reporting
// is off is not kept: fetchBugReportConfig also answers "off" when the
// request fails, and one blip must not hide the button for the rest of
// the session.

import { fetchBugReportConfig, type BugReportConfig } from "./api";

let cached: Promise<BugReportConfig> | null = null;

export function bugReportConfig(): Promise<BugReportConfig> {
  if (!cached) {
    const p = fetchBugReportConfig();
    cached = p;
    void p.then((cfg) => {
      if (!cfg.enabled && cached === p) cached = null;
    });
  }
  return cached;
}

/** Forget the cached answer. For tests. */
export function resetBugReportConfig(): void {
  cached = null;
}
