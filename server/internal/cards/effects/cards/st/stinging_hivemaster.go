package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stinging Hivemaster — Creature — Phyrexian Warlock {2}{B}, 3/2: "Toxic 1.
// When this creature dies, create a 1/1 colorless Phyrexian Mite artifact
// creature token with toxic 1 and 'This token can't block.'" No simplification.
func init() {
	effects.Register(effects.S82OneB04ToxicDiesMite("6b1f750c-0f01-4b1c-ad8f-260e26fbbfea", "Stinging Hivemaster"))
}
