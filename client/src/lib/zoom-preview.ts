// ZoomPreview renders a large, always-upright image of the card
// currently under the cursor. It lives on app.stage (not the scene
// root), so the per-snapshot rebuild in TableRenderer.render leaves
// it untouched — only show()/hide() drive its visibility.
//
// The preview is positioned at a fixed canvas-relative anchor in the
// top-right quadrant so players always know where to look; there's
// no cursor tracking. Texture fetch is lazy (on first show() for a
// given card) and reuses the same Assets-backed cache CardTile uses,
// so the second hover of the same card paints instantly.

import { Container, Graphics, Sprite, Text, Texture, TextStyle } from "pixi.js";
import { TILE_H, TILE_W, textureFor } from "./card-tile";
import type { CardView } from "./protocol";

// ASPECT matches the TILE constants; keeping the preview proportional
// to a real Magic card (63mm × 88mm) so Scryfall art fits without
// letterboxing.
const ASPECT = TILE_W / TILE_H;

// MAX_PREVIEW_H caps the preview height so it doesn't blow past the
// "normal" Scryfall image's native ~680px and start to look soft.
const MAX_PREVIEW_H = 560;

// PREVIEW_HEIGHT_RATIO sizes the preview as a fraction of the canvas
// height. 0.7 reads well across 720p–1440p without dominating the
// scene; the MAX clamp handles ultra-tall viewports.
const PREVIEW_HEIGHT_RATIO = 0.7;

// ANCHOR_* place the preview's centre at 4/5 across horizontally
// and 1/5 down vertically — well into the top-right quadrant but
// not pinned to the corner. The resize step clamps these against
// the preview's own footprint so smaller viewports don't push the
// plate off-canvas.
const ANCHOR_X_RATIO = 0.8;
const ANCHOR_Y_RATIO = 0.2;

// EDGE_MARGIN keeps the preview off the very edge when the anchor
// clamps against a small canvas. Matches the plate's visual inset so
// the border stays a consistent distance from the canvas boundary.
const EDGE_MARGIN = 16;

// PLATE_MARGIN is the outer padding between the card image and the
// plate border. Reused both when drawing the plate (in resize) and
// when computing the effective footprint for anchor clamping.
const PLATE_MARGIN = 6;

// BG_COLOR / BG_ALPHA back the preview sprite with a dark plate so a
// partly-transparent Scryfall JPG still reads against a bright table.
// Alpha of 0.9 is deliberate — enough to dim the card behind without
// fully occluding the scene (you can still sense the battlefield
// state through the preview, which helps spatial memory).
const BG_COLOR = 0x0b1220;
const BG_ALPHA = 0.92;
const BORDER_COLOR = 0x4a5270;

const labelStyle = new TextStyle({
  fontFamily: "system-ui, -apple-system, sans-serif",
  fontSize: 14,
  fill: 0xe0e6f5,
  wordWrap: true,
  align: "center",
});

export class ZoomPreview {
  readonly view: Container;
  private plate: Graphics;
  private sprite: Sprite;
  private nameLabel: Text;
  // currentScryfallID lets us ignore stale texture loads: if the
  // cursor moves to card B before card A's fetch resolves, A's .then
  // must not overwrite B's sprite texture.
  private currentScryfallID: string | null = null;

  constructor() {
    this.view = new Container();
    this.view.visible = false;
    // eventMode "none" keeps the preview from eating pointer events —
    // it's pure chrome; the underlying tile still receives pointerout
    // when the cursor leaves it, even if the preview is rendered on
    // top.
    this.view.eventMode = "none";

    this.plate = new Graphics();
    this.view.addChild(this.plate);

    this.sprite = new Sprite(Texture.EMPTY);
    this.sprite.anchor.set(0.5);
    this.view.addChild(this.sprite);

    this.nameLabel = new Text({ text: "", style: labelStyle });
    this.nameLabel.anchor.set(0.5);
    this.view.addChild(this.nameLabel);
  }

  // resize recomputes the preview's dimensions + anchored position
  // against the current canvas size. Called by TableRenderer on every
  // render() so a window resize is picked up on the next frame.
  resize(canvasWidth: number, canvasHeight: number): void {
    const h = Math.min(canvasHeight * PREVIEW_HEIGHT_RATIO, MAX_PREVIEW_H);
    const w = h * ASPECT;

    // Anchor towards the top-right quadrant, then clamp against the
    // preview's own footprint (plate included) so the plate + border
    // never cross the canvas edge. On tiny viewports where the
    // preview is larger than half the canvas, the clamp collapses
    // toward the canvas centre, which is the graceful fallback.
    const halfW = w / 2 + PLATE_MARGIN;
    const halfH = h / 2 + PLATE_MARGIN;
    const targetX = canvasWidth * ANCHOR_X_RATIO;
    const targetY = canvasHeight * ANCHOR_Y_RATIO;
    const minX = halfW + EDGE_MARGIN;
    const maxX = canvasWidth - halfW - EDGE_MARGIN;
    const minY = halfH + EDGE_MARGIN;
    const maxY = canvasHeight - halfH - EDGE_MARGIN;
    this.view.x = Math.max(minX, Math.min(maxX, targetX));
    this.view.y = Math.max(minY, Math.min(maxY, targetY));

    // Plate sits behind the sprite — draw it with a small outer
    // margin so the border reads cleanly against the table.
    this.plate.clear();
    this.plate.roundRect(
      -w / 2 - PLATE_MARGIN,
      -h / 2 - PLATE_MARGIN,
      w + PLATE_MARGIN * 2,
      h + PLATE_MARGIN * 2,
      10,
    );
    this.plate.fill({ color: BG_COLOR, alpha: BG_ALPHA });
    this.plate.stroke({ color: BORDER_COLOR, width: 1 });

    this.sprite.width = w;
    this.sprite.height = h;

    // Name label is a fallback for placeholder cards that have no
    // scryfall_id (and therefore no image). When a real texture is
    // present the label sits on the card — harmless since it also
    // hides via visibility toggle in show().
    this.nameLabel.x = 0;
    this.nameLabel.y = 0;
    this.nameLabel.style.wordWrapWidth = w - 20;
  }

  // show paints the preview for the given card and makes it visible.
  // If the card's image is already cached the transition is instant;
  // otherwise the plate + name appear first and the sprite populates
  // once the texture resolves.
  show(card: CardView): void {
    // Face-down cards and placeholder cards without a scryfall_id
    // don't get a preview — nothing useful to read.
    if (!card.scryfall_id) {
      this.currentScryfallID = null;
      this.sprite.texture = Texture.EMPTY;
      this.sprite.visible = false;
      this.nameLabel.text = card.name;
      this.nameLabel.visible = true;
      this.view.visible = true;
      return;
    }

    this.currentScryfallID = card.scryfall_id;
    this.nameLabel.text = card.name;
    // Show the name as a placeholder while the normal-size texture
    // is fetched. It's replaced by the sprite the moment the texture
    // resolves (below).
    this.nameLabel.visible = true;
    this.sprite.visible = true;
    this.sprite.texture = Texture.EMPTY;
    this.view.visible = true;

    const expectedID = card.scryfall_id;
    void textureFor(expectedID, "normal").then((tex) => {
      // If the pointer has moved to another card (or off any card)
      // since we kicked off this load, bail — the UI already reflects
      // the newer intent.
      if (this.currentScryfallID !== expectedID) return;
      if (this.view.destroyed) return;
      this.sprite.texture = tex;
      this.nameLabel.visible = false;
    });
  }

  // hide clears the preview. Safe to call when already hidden.
  hide(): void {
    this.currentScryfallID = null;
    this.view.visible = false;
  }

  destroy(): void {
    this.view.destroy({ children: true });
  }
}
