package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temur Banner — Artifact {3} (Khans of Tarkir): "{T}: Add {G}, {U}, or {R}.
// {G}{U}{R}, {T}, Sacrifice this artifact: Draw a card." See KhansBannerSpec.
func init() {
	effects.Register(effects.KhansBannerSpec("8a35f23d-f59a-4677-9736-035c60c22a2b", "Temur Banner", [3]string{"G", "U", "R"}))
}
