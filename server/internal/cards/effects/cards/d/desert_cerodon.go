package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desert Cerodon — Creature — Beast {5}{R}, 6/4
// (Amonkhet, #5535):
//
//	"Cycling {R} ({R}, Discard this card: Draw a card.)"
//
// Cycling is the shared constructor; the cycle draws through the ordinary
// activation from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ad18589-ece4-4809-80f6-6764d2822372",
		Name:         "Desert Cerodon",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{R}")},
	})
}
