// #4213: the hover zoom scales with the screen, height first. jsdom has
// no layout, so these read the stylesheet and evaluate its sizing rules
// at the three heights the owner named: a laptop (768) keeps the size it
// had, a 1080p screen grows, a 1440p screen gets a large zoom, and a
// double-faced card's art shrinks only when the screen is short.
import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";

const src = (p: string) => readFileSync(join(process.cwd(), p), "utf8");
const rule = (css: string, sel: string) => {
  const at = css.indexOf(`${sel} {`);
  if (at < 0) throw new Error(`no rule ${sel}`);
  return css.slice(at).split("}")[0];
};
const decl = (block: string, name: string) => {
  const m = new RegExp(`\\n\\s*${name.replace(/[-]/g, "\\-")}: ([^;]+);`).exec(block);
  if (!m) throw new Error(`no ${name} in ${block}`);
  return m[1].trim();
};

const zoom = src("src/lib/components/board/HoverZoomOverlay.svelte");

/** clamp(Apx, Bvh, Cpx) at a viewport height. */
function artAt(vh: number): number {
  const m = /^clamp\((\d+)px, (\d+)vh, (\d+)px\)$/.exec(
    decl(rule(zoom, ":global(:root)"), "--zoom-art"),
  );
  if (!m) throw new Error("--zoom-art is not clamp(px, vh, px)");
  const [lo, v, hi] = [Number(m[1]), Number(m[2]), Number(m[3])];
  return Math.min(hi, Math.max(lo, (v * vh) / 100));
}

/** The double-faced card's art width at a viewport height. */
function twoSidedAt(vh: number): number {
  const w = decl(rule(zoom, "  .overlay.two-sided .scan:not(.sideways)"), "width");
  const m = /^min\(clamp\((\d+)px, calc\((\d+)vh - (\d+)px\), var\(--zoom-art\)\), 100%\)$/.exec(w);
  if (!m) throw new Error(`unexpected two-sided width ${w}`);
  const [lo, v, off] = [Number(m[1]), Number(m[2]), Number(m[3])];
  return Math.min(artAt(vh), Math.max(lo, (v * vh) / 100 - off));
}

describe("hover zoom size (#4213)", () => {
  it("keeps a laptop's 240px art and grows it on taller screens", () => {
    expect(artAt(768)).toBe(240);
    expect(artAt(1080)).toBeGreaterThan(300);
    expect(artAt(1440)).toBeGreaterThanOrEqual(360);
  });

  it("sizes the panel and the art from the variables, never wider than half the board", () => {
    const overlay = rule(zoom, "  .overlay");
    expect(decl(overlay, "width")).toBe("min(var(--zoom-w), calc(50% - 132px))");
    expect(decl(overlay, "right")).toBe("var(--zoom-right, 10px)");
    expect(decl(overlay, "max-height")).toBe("calc(100% - 20px - var(--dock-zoom-clear, 0px))");
    expect(decl(rule(zoom, "  .scan"), "width")).toBe("min(var(--zoom-art), 100%)");
    expect(decl(rule(zoom, ":global(:root)"), "--zoom-w")).toBe("calc(var(--zoom-art) + 92px)");
    // The sideways (split, battle) scan still fills the panel's width.
    expect(decl(rule(zoom, "  .scan.sideways"), "width")).toBe("100%");
  });

  it("shrinks a double-faced card's art only on a short screen", () => {
    expect(twoSidedAt(768)).toBe(200);
    expect(twoSidedAt(1080)).toBe(artAt(1080));
    expect(twoSidedAt(1440)).toBe(artAt(1440));
  });

  it("keeps the panel on screen beside a wide dock sheet", () => {
    const game = src("src/routes/Game.svelte");
    expect(decl(rule(game, "  section.has-dock.sheet-open"), "--zoom-right")).toContain(
      "calc(100vw - 28px - var(--zoom-w, 332px))",
    );
  });
});
