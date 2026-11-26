package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sultai Banner — Artifact {3} (Khans of Tarkir): "{T}: Add {B}, {G}, or {U}.
// {B}{G}{U}, {T}, Sacrifice this artifact: Draw a card." See KhansBannerSpec.
func init() {
	effects.Register(effects.KhansBannerSpec("93b139a5-2678-46a5-b8eb-45344ee3290a", "Sultai Banner", [3]string{"B", "G", "U"}))
}
