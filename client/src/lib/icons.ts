// Stroke icon set for the client chrome. One weight (1.75 on a 24-unit
// grid), round caps, currentColor — so every glyph takes the seat or
// accent colour it sits in and renders identically on every OS. This
// replaces the emoji glyphs (⚙ 🐞 🔊 👑 ⚔ 🟢 ⚡ ⭐ ☢ ◆ ✦) that used to
// carry these meanings; see docs/design for the foundations sheet.
//
// Each icon is a list of primitives so Icon.svelte can render them
// without {@html}. Keep paths on the 24 grid.

export type IconPrimitive =
  | { t: "path"; d: string; fill?: boolean }
  | { t: "circle"; cx: number; cy: number; r: number; fill?: boolean }
  | { t: "rect"; x: number; y: number; w: number; h: number; rx?: number; fill?: boolean };

const p = (d: string, fill = false): IconPrimitive => ({ t: "path", d, fill });
const c = (cx: number, cy: number, r: number, fill = false): IconPrimitive => ({
  t: "circle",
  cx,
  cy,
  r,
  fill,
});

// A cog: eight trapezoid teeth around a round root, with a centre hole.
// Generated so the teeth stay evenly spaced. (It used to be a circle with
// eight rays, which read as a sun, i.e. a light/dark toggle: #1952.)
const gear: IconPrimitive[] = [];
{
  const root = 7.2;
  const tip = 9.8;
  const pt = (deg: number, r: number) => {
    const a = (deg * Math.PI) / 180;
    return `${(12 + Math.cos(a) * r).toFixed(2)} ${(12 + Math.sin(a) * r).toFixed(2)}`;
  };
  let d = "";
  for (let i = 0; i < 8; i++) {
    const a = i * 45 - 90;
    d += `${i === 0 ? "M" : "A" + root + " " + root + " 0 0 1 "}${pt(a - 13, root)}`;
    d += `L${pt(a - 8, tip)}L${pt(a + 8, tip)}L${pt(a + 13, root)}`;
  }
  d += `A${root} ${root} 0 0 1 ${pt(-90 - 13, root)}Z`;
  gear.push(p(d), c(12, 12, 3));
}

export const ICONS = {
  gear,
  // Site-nav hamburger toggle (PR #1386, the shared header).
  menu: [p("M4 7h16"), p("M4 12h16"), p("M4 17h16")],
  // A house glyph for the site portal's "Home" link — distinct from
  // the diamond brand `mark` so the two never read as duplicates.
  home: [p("M4 11.5 12 4l8 7.5"), p("M6 10v10h12V10")],
  bug: [
    p("M8 7a4 4 0 0 1 8 0v2H8z"),
    { t: "rect", x: 7, y: 9, w: 10, h: 10, rx: 5 } as IconPrimitive,
    p("M4 12h3"),
    p("M17 12h3"),
    p("M5 18l3-2"),
    p("M19 18l-3-2"),
    p("M5 7l3 2"),
    p("M19 7l-3 2"),
  ],
  volume: [
    p("M11 5 6 9H3v6h3l5 4V5z"),
    p("M15.5 8.5a5 5 0 0 1 0 7"),
    p("M18.5 5.5a9 9 0 0 1 0 13"),
  ],
  volumeOff: [p("M11 5 6 9H3v6h3l5 4V5z"), p("M16 9l5 6"), p("M21 9l-5 6")],
  more: [c(5, 12, 1.4, true), c(12, 12, 1.4, true), c(19, 12, 1.4, true)],
  chevronLeft: [p("M15 6l-6 6 6 6")],
  chevronRight: [p("M9 6l6 6-6 6")],
  x: [p("M6 6l12 12"), p("M18 6L6 18")],
  // ADR 0111 PR 6: an action-dock sheet's minimise and restore.
  "chevron-down": [p("M6 9l6 6 6-6")],
  "chevron-up": [p("M6 15l6-6 6 6")],
  check: [p("M5 12l5 5 9-10")],
  link: [
    p("M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1.5 1.5"),
    p("M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1.5-1.5"),
  ],
  // player markers
  crown: [p("M3 18h18l1-11-5 4-5-7-5 7-5-4z"), p("M6 21h12")],
  sword: [p("M14 4l6 6-9 9-6-6z"), p("M8 16l-5 5"), p("M17 3l4 4")],
  drop: [p("M12 3s-6 6.5-6 11a6 6 0 0 0 12 0c0-4.5-6-11-6-11z")],
  bolt: [p("M13 2 4 14h7l-1 8 9-12h-7l1-8z")],
  star: [p("M12 3l2.7 5.6 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.9 1-6.1L3.2 9.5l6.1-.9z")],
  rad: [
    c(12, 12, 2),
    p("M12 3a9 9 0 0 1 7.8 4.5l-5.2 3a3 3 0 0 0-2.6-1.5z"),
    p("M4.2 7.5A9 9 0 0 1 12 3v6a3 3 0 0 0-2.6 1.5z"),
    p("M12 21a9 9 0 0 1-7.8-4.5l5.2-3A3 3 0 0 0 12 15z"),
  ],
  dot: [c(12, 12, 3, true)],
  // ADR 0114 (#2076): the Ring — a band seen at an angle, so it reads
  // as a ring and not as the exile circle or a target. The Ring chip's
  // glyph and the Ring-bearer's marker on the card.
  ring: [
    p("M3.5 10.5a8.5 4.5 0 1 0 17 0a8.5 4.5 0 1 0 -17 0"),
    p("M3.5 10.5v2.5a8.5 4.5 0 0 0 17 0v-2.5"),
    p("M7.5 10.5a4.5 2.2 0 1 0 9 0a4.5 2.2 0 1 0 -9 0"),
  ],
  spark: [p("M12 3l2 7 7 2-7 2-2 7-2-7-7-2 7-2z")],
  // brand mark: rotated square with a filled centre
  mark: [
    p("M12 3l9 9-9 9-9-9z"),
    { t: "rect", x: 10, y: 10, w: 4, h: 4, rx: 0.5, fill: true } as IconPrimitive,
  ],
  // zones and sandbox actions
  library: [{ t: "rect", x: 3, y: 6, w: 12, h: 15, rx: 1.5 } as IconPrimitive, p("M8 3h12v15")],
  grave: [p("M6 21V10a6 6 0 0 1 12 0v11"), p("M4 21h16"), p("M9 13h6"), p("M12 10v6")],
  exile: [c(12, 12, 8), p("M6.5 6.5l11 11")],
  command: [p("M12 3l8 7-8 11-8-11z"), p("M4 10h16")],
  undo: [p("M4 10h10a5 5 0 0 1 0 10H9"), p("M4 10l4-4"), p("M4 10l4 4")],
  // ADR 0121 §8: "Roll a die". A die showing three, and a coin.
  die: [
    { t: "rect", x: 4, y: 4, w: 16, h: 16, rx: 3.5 } as IconPrimitive,
    c(8.5, 8.5, 1.3, true),
    c(12, 12, 1.3, true),
    c(15.5, 15.5, 1.3, true),
  ],
  coin: [c(12, 12, 8), c(12, 12, 4.5)],
  shuffle: [
    p("M3 6h3l10 12h5"),
    p("M21 18l-3-3"),
    p("M21 18l-3 3"),
    p("M3 18h3l3-3.5"),
    p("M14 6h7"),
    p("M21 6l-3-3"),
    p("M21 6l-3 3"),
  ],
  draw: [p("M12 4v11"), p("M7 10l5 5 5-5"), p("M4 20h16")],
  untap: [p("M4 12a8 8 0 1 0 2.5-5.8"), p("M4 4v5h5")],
  flag: [p("M5 21V4h12l-2 4 2 4H5")],
  hand: [
    p("M6 20V9a2 2 0 0 1 4 0v5"),
    p("M10 12V6a2 2 0 0 1 4 0v8"),
    p("M14 13V8a2 2 0 0 1 4 0v9a6 6 0 0 1-6 6H9l-5-6a2 2 0 0 1 3-2.5l1 1.2"),
  ],
  // S31: the public game log — a ruled sheet
  scroll: [
    { t: "rect", x: 4, y: 3, w: 16, h: 18, rx: 2 } as IconPrimitive,
    p("M8 8h8"),
    p("M8 12h8"),
    p("M8 16h5"),
  ],
  // S31 bot seats. A bot has no Discord portrait, so this stands in
  // as the seat's avatar mark — deliberately mechanical rather than a
  // face, so a bot never reads as a player at a glance.
  robot: [
    { t: "rect", x: 4, y: 8, w: 16, h: 12, rx: 3 } as IconPrimitive,
    c(9, 13, 1.4, true),
    c(15, 13, 1.4, true),
    p("M9.5 16.5h5"),
    p("M12 8V4"),
    c(12, 3, 1.4),
    p("M4 12H2"),
    p("M22 12h-2"),
  ],
} satisfies Record<string, IconPrimitive[]>;

export type IconName = keyof typeof ICONS;
