// statsCharts.ts — the geometry of the stats charts (ADR 0217 §8): scales,
// ticks, bar widths, stacked segments, line paths and sparkline cells.
// Pure and DOM-free, so it is unit-tested in node like myStats.ts. The
// Svelte components in lib/components/stats/ draw what this returns as
// inline SVG; nothing here knows about colour.

/** A linear map from a domain to a range. A flat domain maps to the range's start. */
export function linearScale(
  domain: readonly [number, number],
  range: readonly [number, number],
): (v: number) => number {
  const [d0, d1] = domain;
  const [r0, r1] = range;
  const span = d1 - d0;
  if (span === 0) return () => r0;
  return (v) => r0 + ((v - d0) / span) * (r1 - r0);
}

/**
 * niceMax rounds a maximum up to a readable axis top: 1, 2, 2.5, 5 or 10
 * times a power of ten. Zero or less is 1, so an empty chart still has
 * an axis.
 */
export function niceMax(max: number): number {
  if (!(max > 0)) return 1;
  const mag = 10 ** Math.floor(Math.log10(max));
  for (const step of [1, 2, 2.5, 5, 10]) {
    if (max <= step * mag + 1e-9) return step * mag;
  }
  return 10 * mag;
}

/**
 * ticks are round steps from 0 to the first step at or above max, about
 * `count` of them: a step of 1, 2, 2.5 or 5 times a power of ten, so an
 * axis never reads 1.3 or 3.8.
 */
export function ticks(max: number, count = 4): number[] {
  const top = max > 0 ? max : 1;
  const raw = top / Math.max(1, count);
  const mag = 10 ** Math.floor(Math.log10(raw));
  const step = [1, 2, 2.5, 5, 10].map((m) => m * mag).find((s) => s >= raw - 1e-9) ?? 10 * mag;
  const n = Math.ceil(top / step - 1e-9);
  const out: number[] = [];
  for (let i = 0; i <= n; i++) out.push(Number((step * i).toFixed(6)));
  return out;
}

/** One horizontal bar: its own label and value text, so colour is never the only cue. */
export interface ChartBar {
  label: string;
  value: number;
  /** What the bar says beside it: "38%, 5 of 13". */
  text: string;
  /** A CSS colour (a token); var(--chart-1) by default. */
  tone?: string;
  /** A letter drawn before the label (a colour's W). */
  mark?: string;
  muted?: boolean;
  /** A comparison on the same scale as value. */
  tick?: number;
}

/** A bar's share of the widest, 0–1, for a horizontal bar. */
export function barShares(values: readonly number[], max?: number): number[] {
  const top = max ?? Math.max(0, ...values);
  if (!(top > 0)) return values.map(() => 0);
  return values.map((v) => Math.max(0, Math.min(1, v / top)));
}

/** One segment of a 100% stacked bar: where it starts and how wide it is, 0–1. */
export interface Segment {
  index: number;
  start: number;
  share: number;
  value: number;
}

/** stackSegments splits a 100% bar; a zero part gets no segment. */
export function stackSegments(parts: readonly number[]): Segment[] {
  const total = parts.reduce((t, p) => t + Math.max(0, p), 0);
  if (total <= 0) return [];
  const out: Segment[] = [];
  let start = 0;
  parts.forEach((p, index) => {
    if (p <= 0) return;
    const share = p / total;
    out.push({ index, start, share, value: p });
    start += share;
  });
  return out;
}

/** The inner plot box of a chart: its viewBox less the margins. */
export interface Plot {
  width: number;
  height: number;
  left: number;
  top: number;
  right: number;
  bottom: number;
}

export function plotBox(
  width: number,
  height: number,
  margin: { top: number; right: number; bottom: number; left: number },
): Plot {
  return {
    width,
    height,
    left: margin.left,
    top: margin.top,
    right: width - margin.right,
    bottom: height - margin.bottom,
  };
}

/** The centre x of category i of n across a plot (a band scale's midpoint). */
export function bandCentre(plot: Plot, i: number, n: number): number {
  const band = (plot.right - plot.left) / Math.max(1, n);
  return plot.left + band * (i + 0.5);
}

/** A band's width, less its padding share (0–1). */
export function bandWidth(plot: Plot, n: number, padding = 0.25): number {
  return ((plot.right - plot.left) / Math.max(1, n)) * (1 - padding);
}

/** One vertical bar of a histogram, in viewBox units. */
export interface ColumnRect {
  x: number;
  y: number;
  width: number;
  height: number;
  value: number;
}

/** columns lays out vertical bars over n categories, 0 at the plot's bottom. */
export function columns(plot: Plot, values: readonly number[], max?: number): ColumnRect[] {
  const top = niceMax(max ?? Math.max(0, ...values));
  const y = linearScale([0, top], [plot.bottom, plot.top]);
  const w = bandWidth(plot, values.length);
  return values.map((v, i) => {
    const yy = y(Math.max(0, v));
    return {
      x: bandCentre(plot, i, values.length) - w / 2,
      y: yy,
      width: w,
      height: plot.bottom - yy,
      value: v,
    };
  });
}

/**
 * linePath draws a series over n categories as an SVG path. A null value
 * is a gap: the line lifts and starts again after it, so a turn with no
 * data is never drawn as zero.
 */
export function linePath(plot: Plot, values: readonly (number | null)[], max: number): string {
  const y = linearScale([0, max > 0 ? max : 1], [plot.bottom, plot.top]);
  let d = "";
  let pen = false;
  values.forEach((v, i) => {
    if (v == null) {
      pen = false;
      return;
    }
    const x = round(bandCentre(plot, i, values.length));
    const yy = round(y(v));
    d += `${pen ? "L" : "M"}${x} ${yy}`;
    pen = true;
  });
  return d;
}

/** The points of a series, for its markers; null values have none. */
export function linePoints(
  plot: Plot,
  values: readonly (number | null)[],
  max: number,
): { x: number; y: number; value: number; index: number }[] {
  const y = linearScale([0, max > 0 ? max : 1], [plot.bottom, plot.top]);
  const out: { x: number; y: number; value: number; index: number }[] = [];
  values.forEach((v, i) => {
    if (v == null) return;
    out.push({ x: round(bandCentre(plot, i, values.length)), y: round(y(v)), value: v, index: i });
  });
  return out;
}

/** The x of a value on a histogram's category axis, by interpolating between bands. */
export function valueX(plot: Plot, categories: readonly number[], v: number): number | null {
  if (categories.length === 0) return null;
  const n = categories.length;
  if (v <= categories[0]) return bandCentre(plot, 0, n);
  for (let i = 1; i < n; i++) {
    if (v <= categories[i]) {
      const a = categories[i - 1];
      const b = categories[i];
      const t = b === a ? 0 : (v - a) / (b - a);
      return bandCentre(plot, i - 1, n) + t * (bandCentre(plot, i, n) - bandCentre(plot, i - 1, n));
    }
  }
  return bandCentre(plot, n - 1, n);
}

/** A result's sparkline cell: its kind and its height share (a win is tallest). */
export interface SparkCell {
  result: "win" | "loss" | "draw" | "none";
  height: number;
}

/**
 * sparkCells orders results oldest first (the wire sends newest first),
 * so the line reads left to right, and gives each a height: a win full, a
 * draw half, a loss a quarter, no result a sliver. Height and position
 * carry the result as well as colour.
 */
export function sparkCells(newestFirst: readonly string[]): SparkCell[] {
  const h: Record<string, number> = { win: 1, draw: 0.5, loss: 0.25, none: 0.1 };
  return [...newestFirst].reverse().map((r) => {
    const result = (r in h ? r : "none") as SparkCell["result"];
    return { result, height: h[result] };
  });
}

/** Evenness 0–1 as a word for the spread tile. */
export function evennessWord(v: number | null | undefined): string {
  if (v == null) return "-";
  if (v >= 0.8) return "even";
  if (v >= 0.5) return "fairly even";
  if (v >= 0.25) return "focused";
  return "all at one";
}

function round(v: number): number {
  return Math.round(v * 100) / 100;
}
