// CardTile is the visual primitive for a single card in the PixiJS
// scene. One CardTile wraps a Sprite (art), a Graphics (border +
// fallback rectangle when no art is loaded), and a Text (the card
// name, rendered inside the border while the texture is in-flight).
// Tiles flip between face-up and face-down modes — face-down renders
// a solid "back" rectangle because Scryfall's actual card back art
// is Wizards IP and not ours to ship.
//
// Textures are cached at module scope so re-rendering the same card
// doesn't refetch bytes. The cache is keyed on (scryfall_id, size)
// and reuses Pixi's own Assets.load mechanism under the hood, which
// dedupes concurrent loads for us.
//
// Coordinate conventions:
//   - `setPosition(cx, cy)` sets the tile's centre, not its top-left.
//   - The tile's intrinsic dimensions are fixed at TILE_W × TILE_H;
//     callers scale the whole container if they want a different
//     render size.
//   - Tap rotation is 90° clockwise around the tile centre.
//
// PixiJS note: `Assets.load` returns a Texture that we assign to an
// existing Sprite, so the fallback rectangle stays onscreen until
// the bytes arrive (no pop-in to empty space on slow connections).

import { Assets, Container, Graphics, Sprite, Text, Texture, TextStyle } from "pixi.js";
import type { CardView } from "./protocol";

// MTG cards are 63mm × 88mm — aspect ~0.716:1. 80×112 hits that
// ratio exactly and is large enough to read a name at 1080p without
// dominating the scene.
export const TILE_W = 80;
export const TILE_H = 112;

// BACK_COLOR is the fill for a face-down card. Keeping it a single
// constant so opponent hand, opponent library, and top-of-library
// backs all look identical.
const BACK_COLOR = 0x1a1f35;

// LOADING_COLOR is the fallback fill used while the art texture is
// in-flight. Slightly lighter than the back so users can distinguish
// "face-up, art loading" from "face-down on purpose".
const LOADING_COLOR = 0x1e2638;

// BORDER_COLOR is the thin outline drawn around every tile. Darker
// than the backgrounds so it reads as a frame on both seat colors
// and the shared band.
const BORDER_COLOR = 0x4a5270;

// The tile size is fixed, so we can cache the TextStyle at module
// scope rather than re-allocating per tile. Matches the pattern in
// table.ts — Pixi internally parses font strings and builds shared
// font metrics; instantiating once saves GC churn on redraws.
const nameStyle = new TextStyle({
  fontFamily: "system-ui, -apple-system, sans-serif",
  fontSize: 10,
  fill: 0xe0e6f5,
  wordWrap: true,
  wordWrapWidth: TILE_W - 12,
  align: "center",
});

// Texture cache. Pixi's Assets.load dedupes concurrent requests by
// URL, so we cache the Promise, not the Texture — callers await it
// and always see the same backing Texture for a given key.
const textureCache = new Map<string, Promise<Texture>>();

export function textureFor(
  scryfallID: string,
  size: "small" | "normal" | "large" = "small",
): Promise<Texture> {
  const key = `${scryfallID}:${size}`;
  let pending = textureCache.get(key);
  if (!pending) {
    // Same-origin; the client's Vite proxy / deployed reverse proxy
    // sends /cards/* to the Go server which handles CDN fallback.
    //
    // loadParser is explicit because our URL path ends in /image with
    // no file extension — Pixi's default parser selection sniffs the
    // extension and bails out, so Assets.load rejects without this
    // hint. The response is always a JPEG from the Go image cache.
    pending = Assets.load<Texture>({
      src: `/cards/${scryfallID}/image?size=${size}`,
      loadParser: "loadTextures",
    });
    textureCache.set(key, pending);
  }
  return pending;
}

export interface CardTileOptions {
  // faceDown forces a face-down render regardless of the card's
  // real identity — used for opponent hand + library. Art is never
  // fetched in this mode, so no network traffic for hidden cards.
  faceDown: boolean;
}

// HOVER_LIFT is how many pixels the hovered card translates "up"
// (negative y) while hovered. Tuned so a fanned hand card clears
// its neighbors without fully detaching from the fan.
const HOVER_LIFT = 24;

// hoveredCardID tracks which card instance currently has the
// pointer over it, at module scope. This survives across tile
// destroy / recreate cycles: when a snapshot arrives mid-hover,
// the old tile is torn down without a synthetic pointerout and
// the new tile (same instance_id) is constructed at resting
// position — visually a "drop flash." Persisting the hovered ID
// lets the caller re-apply hover on the replacement tile so the
// lift stays put until the user actually moves their pointer.
let hoveredCardID: string | null = null;

// isHovered reports whether the given instance_id is the current
// hover target. Callers use it after addChild to restore hover
// state on freshly-constructed tiles.
export function isHovered(cardInstanceID: string): boolean {
  return hoveredCardID === cardInstanceID;
}

// CardTile owns its own Container; callers add it to a parent scene
// graph via `tile.view`.
export class CardTile {
  readonly view: Container;
  readonly cardID: string;
  private fill: Graphics;
  private sprite: Sprite | null = null;
  private label: Text | null = null;
  // Resting position + rotation so hover + tap transforms compose
  // without the fan layout being "sticky" across state changes.
  private restingX = 0;
  private restingY = 0;
  private restingRotation = 0;
  private tapped = false;
  private hovered = false;

  constructor(card: CardView, opts: CardTileOptions) {
    this.view = new Container();
    this.cardID = card.instance_id;
    this.fill = new Graphics();
    this.view.addChild(this.fill);

    this.draw(card, opts);
  }

  // setPosition places the tile's centre at (cx, cy). Mirrors the
  // anchor-at-centre convention used for seat panels in table.ts.
  setPosition(cx: number, cy: number): void {
    this.restingX = cx;
    this.restingY = cy;
    this.applyTransform();
  }

  // setRotation stores a "resting" rotation (radians) that composes
  // with tap state. Callers use this for the fan layout; the tile
  // applies it alongside tap to produce the final transform.
  setRotation(rotation: number): void {
    this.restingRotation = rotation;
    this.applyTransform();
  }

  // setTapped rotates the tile 90° clockwise around its centre.
  setTapped(tapped: boolean): void {
    this.tapped = tapped;
    this.applyTransform();
  }

  // setHover lifts the tile toward screen-up while hovered. The lift
  // is in parent (screen) space, not tile-local, so a rotated fan
  // card still visibly rises rather than drifting sideways.
  setHover(hovered: boolean): void {
    this.hovered = hovered;
    this.applyTransform();
    // Z-order: while hovered, pop the tile to the top of its parent
    // so it's not clipped by later-drawn neighbors in the fan.
    if (hovered) {
      const parent = this.view.parent;
      if (parent) parent.setChildIndex(this.view, parent.children.length - 1);
    }
  }

  // makeInteractive wires pointerover / pointerout to setHover and
  // keeps the module-scope hoveredCardID in sync so freshly-created
  // replacement tiles (after a snapshot rebuild) can restore hover
  // state without waiting for a pointermove.
  //
  // The caller opts in — the battlefield draw path keeps tiles
  // static so a hand-card hover doesn't bleed into battlefield
  // behaviour.
  makeInteractive(): void {
    this.view.eventMode = "static";
    this.view.cursor = "pointer";
    this.view.on("pointerover", () => {
      hoveredCardID = this.cardID;
      this.setHover(true);
    });
    this.view.on("pointerout", () => {
      if (hoveredCardID === this.cardID) hoveredCardID = null;
      this.setHover(false);
    });
  }

  destroy(): void {
    this.view.destroy({ children: true });
  }

  // applyTransform composes resting position + rotation + tap + hover
  // into the tile's final PIXI transform. Called after any state
  // change so ordering of set* calls doesn't matter.
  private applyTransform(): void {
    this.view.x = this.restingX;
    this.view.y = this.restingY + (this.hovered ? -HOVER_LIFT : 0);
    this.view.rotation = this.restingRotation + (this.tapped ? Math.PI / 2 : 0);
  }

  private draw(card: CardView, opts: CardTileOptions): void {
    // Fallback rectangle. Drawn first so the sprite can layer on top
    // once it loads.
    this.fill.clear();
    this.fill.roundRect(-TILE_W / 2, -TILE_H / 2, TILE_W, TILE_H, 6);
    this.fill.fill({ color: opts.faceDown ? BACK_COLOR : LOADING_COLOR });
    this.fill.stroke({ color: BORDER_COLOR, width: 1 });

    if (opts.faceDown) {
      // Face-down tiles never get art. We're done.
      return;
    }

    // Name label — shown while art loads, and kept as a caption
    // overlay that reads fine even once the sprite paints over it
    // (Scryfall art is mostly concentrated in the top 2/3 of the
    // card; our label sits at the top in the name-bar zone).
    this.label = new Text({ text: card.name, style: nameStyle });
    this.label.anchor.set(0.5, 0);
    this.label.x = 0;
    this.label.y = -TILE_H / 2 + 4;
    this.view.addChild(this.label);

    if (!card.scryfall_id) {
      // Placeholder cards (demo game, or decks that haven't been
      // resolved against Scryfall) never get art. The label + fill
      // is the whole tile.
      return;
    }

    // Kick off the texture load. The sprite is added immediately at
    // full tile dimensions so once the texture resolves the image
    // fits without a relayout.
    const placeholder = new Sprite(Texture.EMPTY);
    placeholder.anchor.set(0.5);
    placeholder.x = 0;
    placeholder.y = 0;
    placeholder.width = TILE_W;
    placeholder.height = TILE_H;
    this.view.addChild(placeholder);
    this.sprite = placeholder;

    void textureFor(card.scryfall_id).then((tex) => {
      // The tile may have been destroyed while we were waiting
      // — skip the assignment in that case.
      if (this.view.destroyed || !this.sprite) return;
      this.sprite.texture = tex;
    });
  }
}
