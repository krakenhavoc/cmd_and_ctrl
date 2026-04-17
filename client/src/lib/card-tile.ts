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

function textureFor(scryfallID: string, size: "small" | "normal" = "small"): Promise<Texture> {
  const key = `${scryfallID}:${size}`;
  let pending = textureCache.get(key);
  if (!pending) {
    // Same-origin; the client's Vite proxy / deployed reverse proxy
    // sends /cards/* to the Go server which handles CDN fallback.
    pending = Assets.load<Texture>(`/cards/${scryfallID}/image?size=${size}`);
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

// CardTile owns its own Container; callers add it to a parent scene
// graph via `tile.view`.
export class CardTile {
  readonly view: Container;
  private fill: Graphics;
  private sprite: Sprite | null = null;
  private label: Text | null = null;

  constructor(card: CardView, opts: CardTileOptions) {
    this.view = new Container();
    this.fill = new Graphics();
    this.view.addChild(this.fill);

    this.draw(card, opts);
  }

  // setPosition places the tile's centre at (cx, cy). Mirrors the
  // anchor-at-centre convention used for seat panels in table.ts.
  setPosition(cx: number, cy: number): void {
    this.view.x = cx;
    this.view.y = cy;
  }

  // setTapped rotates the tile 90° clockwise around its centre. The
  // Container's pivot is placed at (0, 0) — tile-local centre —
  // because every visual child is already drawn centred.
  setTapped(tapped: boolean): void {
    this.view.rotation = tapped ? Math.PI / 2 : 0;
  }

  destroy(): void {
    this.view.destroy({ children: true });
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
