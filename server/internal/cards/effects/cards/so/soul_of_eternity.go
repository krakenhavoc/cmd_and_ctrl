package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul of Eternity — {5}{W}{W} Creature — Avatar */*, Commander Legends (#5507):
//
//	"Soul of Eternity's power and toughness are each equal to your life total.
//	 Encore {7}{W}{W}"
//
// A layer 7a characteristic-defining ability (CR 604.3) that reads its
// controller's life total (its owner's off the battlefield, CR 108.4a) on every
// layer pass, and the shared Encore constructor (encore.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d211447-9f18-4d23-8f19-cdf6ad3c68ba",
		Name:         "Soul of Eternity",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:              "Soul of Eternity's power and toughness are each equal to your life total.",
			Layer:              game.Layer7PT,
			SubLayer:           game.SubLayer7A_CDA,
			AppliesTo:          effects.SelfOnly,
			DependsOnLifeTotal: true,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.YourLifeTotal(g, source)
				c.Power, c.Toughness = n, n
			},
		}},
		Activated: []effects.ActivatedAbility{effects.Encore("{7}{W}{W}")},
	})
}
