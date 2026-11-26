package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeskai Banner — Artifact {3} (Khans of Tarkir): "{T}: Add {U}, {R}, or {W}.
// {U}{R}{W}, {T}, Sacrifice this artifact: Draw a card." See KhansBannerSpec.
func init() {
	effects.Register(effects.KhansBannerSpec("fda10ef2-f431-4d9b-96a4-4db4c98d2b0c", "Jeskai Banner", [3]string{"U", "R", "W"}))
}
