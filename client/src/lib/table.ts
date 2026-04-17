// PixiJS table layout for the 4-player Commander sandbox. At S05
// this is a deliberately static scene — colored rectangles per zone
// with the player name + card count overlaid. Interaction (drag,
// zoom, hover preview) lands in later sprints; the point here is to
// prove the seam from GameView snapshots to a rendered table and
// get the four seats laid out correctly.
//
// The module exports a single TableRenderer class that owns a
// PIXI.Application and exposes a `render(view, viewerID)` method
// consumers call whenever a new snapshot arrives. PIXI's resizeTo
// handles canvas sizing; callers re-invoke render() on resize to
// recompute the anchor layout against the new dimensions.
//
// Scryfall card-back art is Wizards' IP, so we never fetch it. The
// placeholder card "back" is a solid rectangle with a thin border,
// reused everywhere a face-down card is shown (opponent hand,
// library). That keeps S05 off the asset-pipeline critical path.

import { Application, Container, Graphics, Text, TextStyle } from "pixi.js";
import { CardTile, TILE_H, TILE_W } from "./card-tile";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";

// SeatPosition places one of the four seats around the table. "self"
// always renders at the bottom; opponents rotate clockwise. With
// fewer than four players we keep the self position fixed and spread
// the remaining seats around the other three anchors, leaving gaps.
type SeatPosition = "self" | "left" | "top" | "right";
const seatOrder: SeatPosition[] = ["self", "left", "top", "right"];

// Per-zone dimensions in layout units (pre-scale). The renderer
// scales the whole scene so the table fits the canvas; individual
// zone sizes stay proportional. Units are roughly pixels at 1080p.
const Z_WIDTH = 280; // zone panel width
const Z_HEIGHT = 120; // zone panel height
const PAD = 12; // inner padding between zones
const SHARED_BAND = 180; // height of the shared (battlefield/stack/exile) band

// SeatColor is the base fill per seat, for quick visual
// disambiguation before the proper UI chrome lands. These are
// intentionally desaturated so they don't fight future card art.
const seatColors: Record<SeatPosition, number> = {
  self: 0x1f3a52,
  left: 0x4a2e52,
  top: 0x4f3f1a,
  right: 0x1f4a38,
};

// BG_COLOR is the Pixi stage clear color. Kept next to the zone
// palette so theme tweaks live in one place; the canvas CSS background
// has been dropped to avoid a double source-of-truth.
const BG_COLOR = 0x0b1220;

// TextStyle instances are expensive to allocate — Pixi internally
// parses font strings and builds a shared FontMetrics. Cache one per
// (size,color) tuple so a full redraw doesn't churn the GC.
const styleCache = new Map<string, TextStyle>();
function labelStyle(size: number, color: number): TextStyle {
  const key = `${size}:${color}`;
  let s = styleCache.get(key);
  if (!s) {
    s = new TextStyle({
      fontFamily: "system-ui, -apple-system, sans-serif",
      fontSize: size,
      fill: color,
    });
    styleCache.set(key, s);
  }
  return s;
}

// ActionSender mirrors GameClient.sendAction so the renderer can
// emit mutations without pulling the websocket client into its
// import graph. Params shape is caller-defined; the server-side
// docs/protocol.md is the source of truth for each action type.
export type ActionSender = (type: string, params?: unknown, player?: string) => void;

export interface RenderOptions {
  // viewerID is the seat the current client belongs to. The
  // matching PlayerView is rendered at the "self" anchor; others
  // rotate to left/top/right in their original seat order.
  viewerID: string | null;
  // sendAction, when present, is invoked from user interactions
  // (click-to-tap, click-to-play). Omitting it makes the renderer
  // read-only — useful for spectator / admin views.
  sendAction?: ActionSender;
}

export class TableRenderer {
  readonly app: Application;
  private root: Container;
  private initialized = false;

  constructor() {
    this.app = new Application();
    this.root = new Container();
  }

  // init is async because PIXI v8 moved initialisation off the
  // constructor. Call once, before render().
  async init(container: HTMLElement): Promise<void> {
    await this.app.init({
      resizeTo: container,
      background: BG_COLOR,
      antialias: true,
    });
    container.appendChild(this.app.canvas);
    this.app.stage.addChild(this.root);
    this.initialized = true;
  }

  destroy(): void {
    if (!this.initialized) return;
    this.app.destroy(true, { children: true });
    this.initialized = false;
  }

  // render wipes and rebuilds the scene. At S05 this is the simplest
  // possible approach — later sprints can diff the PIXI graph for
  // smoother transitions, but at four seats this redraws in well
  // under a millisecond and "simple" wins until it's measurably
  // wrong.
  render(view: GameView, opts: RenderOptions): void {
    if (!this.initialized) return;
    this.root.removeChildren();

    // Assign every seat to a SeatPosition. The viewer goes to "self";
    // remaining seats fill left → top → right in their original
    // order. With fewer than four seats the trailing anchors are
    // left empty (no rectangle drawn).
    const placements = assignSeats(view.seats, opts.viewerID);

    // Canvas size for scaling. Falls back to a reasonable default
    // when the canvas hasn't laid out yet (first render, before
    // the resize observer has fired).
    const width = this.app.renderer.width || 1280;
    const height = this.app.renderer.height || 720;

    // Seat anchor points (centre of each seat's zone cluster).
    const anchors: Record<SeatPosition, { x: number; y: number }> = {
      self: { x: width / 2, y: height - Z_HEIGHT - PAD },
      top: { x: width / 2, y: Z_HEIGHT / 2 + PAD },
      left: { x: Z_WIDTH / 2 + PAD, y: height / 2 },
      right: { x: width - Z_WIDTH / 2 - PAD, y: height / 2 },
    };

    // Shared zones band across the middle.
    drawSharedBand(this.root, view, width, height, opts);

    // Per-seat panels.
    for (const [pos, seat] of Object.entries(placements) as [SeatPosition, PlayerView | null][]) {
      if (!seat) continue;
      const a = anchors[pos];
      drawSeat(this.root, seat, pos, a.x, a.y, opts.viewerID === seat.id, opts);
    }

    // Self-hand fan, drawn last so it sits on top of the seat panel
    // and the shared band. Only the viewer has full-fidelity hand
    // contents (FilterViewFor zeroes out opponent hand cards on the
    // server), so this branch is a no-op for spectator / admin views.
    const self = placements.self;
    if (self && self.hand.cards.length > 0) {
      drawHandFan(this.root, self.hand.cards, width, height, opts);
    }
  }
}

function assignSeats(
  seats: PlayerView[],
  viewerID: string | null,
): Record<SeatPosition, PlayerView | null> {
  const out: Record<SeatPosition, PlayerView | null> = {
    self: null,
    left: null,
    top: null,
    right: null,
  };
  const selfIdx = viewerID ? seats.findIndex((s) => s.id === viewerID) : -1;
  if (selfIdx >= 0) {
    out.self = seats[selfIdx];
    // Rotate remaining seats clockwise: left, top, right.
    const others = [...seats.slice(selfIdx + 1), ...seats.slice(0, selfIdx)];
    const spots: SeatPosition[] = ["left", "top", "right"];
    others.slice(0, 3).forEach((s, i) => {
      out[spots[i]] = s;
    });
  } else {
    // Spectator / missing viewerID — drop seats into the first N
    // anchors in original order.
    seats.slice(0, 4).forEach((s, i) => {
      out[seatOrder[i]] = s;
    });
  }
  return out;
}

function drawSeat(
  root: Container,
  seat: PlayerView,
  pos: SeatPosition,
  cx: number,
  cy: number,
  isSelf: boolean,
  opts: RenderOptions,
): void {
  const seatColor = seatColors[pos];
  const panel = new Container();
  panel.x = cx;
  panel.y = cy;

  // Header: player name + life + seat #.
  const header = new Graphics();
  header.roundRect(-Z_WIDTH / 2, -Z_HEIGHT / 2, Z_WIDTH, 32, 4);
  header.fill({ color: seatColor, alpha: 0.9 });
  panel.addChild(header);

  const nameText = new Text({
    text: `${seat.name}  ·  ${seat.life} life  ·  seat ${seat.seat}`,
    style: labelStyle(14, 0xffffff),
  });
  nameText.x = -Z_WIDTH / 2 + 8;
  nameText.y = -Z_HEIGHT / 2 + 6;
  panel.addChild(nameText);

  // Zone rectangles. Hand at the front (for self) or as a face-down
  // stack with count (for opponents). Library, graveyard, command
  // side-by-side below the header.
  const zoneY = -Z_HEIGHT / 2 + 40;
  const zones: [string, ZoneView, boolean][] = [
    ["cmd", seat.command, false],
    ["lib", seat.library, true], // library is always face-down
    ["grv", seat.graveyard, false],
    ["hand", seat.hand, !isSelf], // opponent hand is face-down
  ];
  const perZoneW = (Z_WIDTH - 8) / zones.length;
  zones.forEach(([label, zone, faceDown], i) => {
    const zx = -Z_WIDTH / 2 + 4 + i * perZoneW;
    const g = new Graphics();
    g.roundRect(zx, zoneY, perZoneW - 4, Z_HEIGHT - 44, 3);
    g.fill({ color: faceDown ? 0x2b2b3c : 0x1b1b2a });
    g.stroke({ color: 0x5a5a78, width: 1 });
    panel.addChild(g);

    const lab = new Text({ text: label, style: labelStyle(10, 0x999db5) });
    lab.x = zx + 4;
    lab.y = zoneY + 2;
    panel.addChild(lab);

    const count = new Text({
      text: String(zone.count),
      style: labelStyle(18, 0xffffff),
    });
    count.anchor.set(0.5);
    count.x = zx + (perZoneW - 4) / 2;
    count.y = zoneY + (Z_HEIGHT - 44) / 2 + 4;
    panel.addChild(count);

    // Self-seat library click → draw one card. Other zones stay
    // passive until Phase 5's modal browser lands — this is the one
    // interaction the S06 exit criteria needs ("draws 7 cards").
    if (isSelf && label === "lib" && opts.sendAction) {
      const send = opts.sendAction;
      const viewerID = opts.viewerID ?? undefined;
      g.eventMode = "static";
      g.cursor = "pointer";
      g.on("pointertap", () => {
        send("draw_card", undefined, viewerID);
      });
    }
  });

  root.addChild(panel);
}

function drawSharedBand(
  root: Container,
  view: GameView,
  width: number,
  height: number,
  opts: RenderOptions,
): void {
  const band = new Graphics();
  const y = height / 2 - SHARED_BAND / 2;
  band.roundRect(PAD, y, width - PAD * 2, SHARED_BAND, 8);
  band.fill({ color: 0x111a2b, alpha: 0.85 });
  band.stroke({ color: 0x2e3a55, width: 1 });
  root.addChild(band);

  // Band layout: the battlefield takes the left two-thirds (so cards
  // have room to breathe at (battle_x, battle_y)); stack + exile are
  // compact count chips on the right. This is a transitional layout —
  // Phase 4 will give the battlefield its own full row when we add
  // drag-to-place.
  const battleW = ((width - PAD * 2 - 16) * 2) / 3;
  const sideW = ((width - PAD * 2 - 16) * 1) / 3 / 2;
  const battleX = PAD + 8;
  const battleY = y + 8;
  const battleH = SHARED_BAND - 16;

  // Battlefield panel — cards are positioned inside this rect via
  // their normalised battle_x/battle_y (both in [0, 1]).
  const bfPanel = new Graphics();
  bfPanel.roundRect(battleX, battleY, battleW, battleH, 4);
  bfPanel.fill({ color: 0x1a2540 });
  bfPanel.stroke({ color: 0x3e4a70, width: 1 });
  root.addChild(bfPanel);

  const bfLab = new Text({ text: "battlefield", style: labelStyle(12, 0xbbc4dd) });
  bfLab.x = battleX + 8;
  bfLab.y = battleY + 6;
  root.addChild(bfLab);

  drawBattlefieldCards(root, view.battlefield.cards, battleX, battleY, battleW, battleH, opts);

  // Stack + exile — compact count chips.
  const sides: [string, ZoneView][] = [
    ["stack", view.stack],
    ["exile", view.exile],
  ];
  const sideStart = battleX + battleW + 8;
  sides.forEach(([label, zone], i) => {
    const zx = sideStart + i * (sideW + 8);
    const g = new Graphics();
    g.roundRect(zx, battleY, sideW, battleH, 4);
    g.fill({ color: 0x1a2540 });
    g.stroke({ color: 0x3e4a70, width: 1 });
    root.addChild(g);

    const lab = new Text({ text: label, style: labelStyle(12, 0xbbc4dd) });
    lab.x = zx + 8;
    lab.y = battleY + 6;
    root.addChild(lab);

    const count = new Text({ text: String(zone.count), style: labelStyle(26, 0xffffff) });
    count.anchor.set(0.5);
    count.x = zx + sideW / 2;
    count.y = battleY + battleH / 2;
    root.addChild(count);
  });
}

// drawHandFan lays the viewer's hand in an arc at the bottom of the
// canvas. Cards nearer the centre sit slightly higher, and each card
// is rotated proportional to its offset from the centre so the fan
// reads as a real hand rather than a flat strip.
//
// The fan uses a virtual arc of radius R = ~2× tile height, which is
// tight enough that 7+ cards still fit in 1280px-wide canvases but
// wide enough that adjacent cards don't over-occlude each other.
function drawHandFan(
  root: Container,
  cards: CardView[],
  width: number,
  height: number,
  opts: RenderOptions,
): void {
  const n = cards.length;
  // Per-card angular step, clamped so very large hands don't fan
  // past 60° total (at which point cards start pointing sideways).
  const maxTotal = Math.PI / 3; // 60°
  const perCard = Math.min(0.12, maxTotal / Math.max(1, n - 1));
  const total = perCard * Math.max(0, n - 1);
  const radius = TILE_H * 2;

  const centerX = width / 2;
  // baseY places the fan's arc pivot below the canvas so card centres
  // sit just inside the bottom edge.
  const baseY = height + radius - TILE_H * 0.9;

  const viewerID = opts.viewerID ?? undefined;
  for (let i = 0; i < n; i++) {
    const c = cards[i];
    const angle = -total / 2 + i * perCard;
    const x = centerX + radius * Math.sin(angle);
    const y = baseY - radius * Math.cos(angle);
    const tile = new CardTile(c, { faceDown: false });
    tile.setPosition(x, y);
    tile.setRotation(angle);
    tile.makeInteractive();
    // Click to play — emits play_card with the viewer as the player.
    // Read-only views (no sendAction) skip this so spectators can't
    // mutate state.
    if (opts.sendAction) {
      const send = opts.sendAction;
      tile.view.on("pointertap", () => {
        send("play_card", { instance_id: c.instance_id }, viewerID);
      });
    }
    root.addChild(tile.view);
  }
}

// drawBattlefieldCards paints one CardTile per card, positioned by
// its normalised (battle_x, battle_y) inside the given rect. Cards
// without positions (default 0,0) land at the top-left — the Phase 4
// drag UX will stamp a real position on release.
//
// Tiles are inset by half their size so an edge-aligned position
// still fits inside the rect rather than bleeding past the border.
function drawBattlefieldCards(
  root: Container,
  cards: CardView[],
  rx: number,
  ry: number,
  rw: number,
  rh: number,
  opts: RenderOptions,
): void {
  const innerX = rx + TILE_W / 2 + 4;
  const innerY = ry + TILE_H / 2 + 4;
  const innerW = rw - TILE_W - 8;
  const innerH = rh - TILE_H - 8;
  for (const c of cards) {
    const tile = new CardTile(c, { faceDown: false });
    const bx = c.battle_x ?? 0;
    const by = c.battle_y ?? 0;
    tile.setPosition(innerX + bx * innerW, innerY + by * innerH);
    tile.setTapped(Boolean(c.tapped));
    if (opts.sendAction) {
      // Click-to-toggle-tap on the battlefield. Drag-to-reposition
      // lands in a later phase — this handler fires on a static
      // pointertap (press + release without movement), so a future
      // drag implementation won't collide with it.
      wireTapClick(tile, c, opts.sendAction, innerX, innerY, innerW, innerH);
    }
    root.addChild(tile.view);
  }
}

// wireTapClick attaches click-to-toggle-tap and drag-to-reposition
// to a battlefield tile. A pointerdown that stays within a small
// move threshold before pointerup is treated as a tap-toggle;
// anything further is a drag that ends with set_battlefield_position
// on release, carrying the tile's new (x, y) renormalised against
// the battlefield rect.
//
// Drag state is local to this closure — each tile gets its own.
function wireTapClick(
  tile: CardTile,
  card: CardView,
  send: ActionSender,
  innerX: number,
  innerY: number,
  innerW: number,
  innerH: number,
): void {
  tile.view.eventMode = "static";
  tile.view.cursor = "pointer";
  // DRAG_THRESHOLD is the cursor-movement distance (in pixels) that
  // separates a click from a drag. Below the threshold we emit
  // tap/untap; above it we treat the gesture as a reposition.
  const DRAG_THRESHOLD = 4;
  let downAt: { x: number; y: number } | null = null;
  let dragging = false;

  tile.view.on("pointerdown", (e) => {
    downAt = { x: e.global.x, y: e.global.y };
    dragging = false;
  });
  tile.view.on("globalpointermove", (e) => {
    if (!downAt) return;
    const dx = e.global.x - downAt.x;
    const dy = e.global.y - downAt.y;
    if (!dragging && dx * dx + dy * dy > DRAG_THRESHOLD * DRAG_THRESHOLD) {
      dragging = true;
    }
    if (dragging) {
      // Track the cursor under the tile's centre. Parent space maps
      // 1:1 to screen here (no transforms on root).
      tile.setPosition(e.global.x, e.global.y);
    }
  });
  const finish = (e: { global: { x: number; y: number } }) => {
    if (!downAt) return;
    const started = downAt;
    downAt = null;
    if (!dragging) {
      send(card.tapped ? "untap" : "tap", { instance_id: card.instance_id });
      return;
    }
    // Drop position → renormalise against the battlefield rect.
    // Clamp so a release outside the rect still stamps a valid
    // [0, 1] coordinate; the server also clamps defensively.
    const nx = clamp01((e.global.x - innerX) / innerW);
    const ny = clamp01((e.global.y - innerY) / innerH);
    send("set_battlefield_position", { instance_id: card.instance_id, x: nx, y: ny });
    // Silence "unused started" noise — we may want to re-introduce
    // gesture-timing heuristics (long-press menu) later.
    void started;
  };
  tile.view.on("pointerup", finish);
  tile.view.on("pointerupoutside", finish);
}

function clamp01(v: number): number {
  if (v < 0) return 0;
  if (v > 1) return 1;
  return v;
}
