// tablesChanged.ts — a window event that says "your tables changed
// somewhere other than this page".
//
// The Tables page polls only while a table is Now, so a table made
// elsewhere (the Played with drawer's New table, ADR 0213 §1) would not
// show until the next focus. The drawer dispatches this; the page
// listens and re-reads. No payload: the page asks the server.

export const TABLES_CHANGED_EVENT = "c2:tables-changed";
