package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Banner — Artifact {3} (Khans of Tarkir): "{T}: Add {W}, {B}, or {G}.
// {W}{B}{G}, {T}, Sacrifice this artifact: Draw a card." See KhansBannerSpec.
func init() {
	effects.Register(effects.KhansBannerSpec("46535f8e-1bcd-4588-ac6c-a4bc89c379c8", "Abzan Banner", [3]string{"W", "B", "G"}))
}
