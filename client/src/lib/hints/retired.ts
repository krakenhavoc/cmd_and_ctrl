// retired.ts — the ids of hints that were removed (ADR 0125 §3.2, §7).
//
// A hint id is never reused. Someone who dismissed the old hint has its
// id in `settings.help.seen`, so a new hint under the same id would never
// be offered to them. When you delete a `*.hint.ts`, add its id here in
// the same PR: hints.test.ts fails when a live hint carries a retired id,
// and settings.ts drops retired ids from the seen map (unknown ids are
// kept; only these are pruned).
//
// Imports nothing, so settings.ts can read it without pulling in the
// hint collection.

export const RETIRED_HINT_IDS: readonly string[] = [];
