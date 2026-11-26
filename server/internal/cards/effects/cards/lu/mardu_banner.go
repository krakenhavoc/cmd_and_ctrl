package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Banner — Artifact {3} (Khans of Tarkir): "{T}: Add {R}, {W}, or {B}.
// {R}{W}{B}, {T}, Sacrifice this artifact: Draw a card." See KhansBannerSpec.
func init() {
	effects.Register(effects.KhansBannerSpec("9805b7b3-086a-4c91-810f-52c8f8f95f6a", "Mardu Banner", [3]string{"R", "W", "B"}))
}
