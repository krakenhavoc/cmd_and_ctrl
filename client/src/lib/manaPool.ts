import type { ManaPoolRun } from "./protocol";

// manaPool.ts: reading PlayerView.mana_pool, which is runs of one
// colour (#3041), not one letter per mana.

// manaPoolRuns is the pool as runs, also reading the pre-#3041 form
// (one bare letter per mana) as runs of one, so a frame recorded
// before runs (a replay, a stored view) still draws.
export function manaPoolRuns(pool: readonly (ManaPoolRun | string)[] | undefined): ManaPoolRun[] {
  const out: ManaPoolRun[] = [];
  for (const run of pool ?? []) {
    out.push(typeof run === "string" ? { color: run, count: 1 } : run);
  }
  return out;
}

// manaPoolTotal is how many mana a seat has floating.
export function manaPoolTotal(pool: readonly ManaPoolRun[] | undefined): number {
  let n = 0;
  for (const run of manaPoolRuns(pool)) n += run.count;
  return n;
}

// manaPoolOf builds a pool from one letter per mana, in order, folding
// adjacent letters of one colour into a run, as the server does. For
// tests and fixtures that think in mana.
export function manaPoolOf(...colors: string[]): ManaPoolRun[] {
  const out: ManaPoolRun[] = [];
  for (const color of colors) {
    const last = out[out.length - 1];
    if (last && last.color === color) last.count++;
    else out.push({ color, count: 1 });
  }
  return out;
}
