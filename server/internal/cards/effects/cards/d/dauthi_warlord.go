package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauthi Warlord — Creature — Dauthi Soldier {1}{B}, */1 (Exodus, S82 OLD batch 744, #5626):
//
//	"Shadow
//	 Dauthi Warlord's power is equal to the number of creatures on the
//	 battlefield with shadow."
//
// A characteristic-defining power (CR 604.3, layer 7a), Beast of Burden's shape,
// counting every creature on the battlefield with shadow after layer 6 (its own
// shadow included, and any creature that gains it), whoever controls it.
// Shadow rides PrintedKeywords; the printed toughness is 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "42bb6537-fa54-43b7-84f2-dea41adc35fd",
		Name:            "Dauthi Warlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Static: []game.StaticAbility{{
			Label:     "Dauthi Warlord's power is equal to the number of creatures on the battlefield with shadow.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				c.Power = effects.S82OldB744DauthiWarlordPower(g)
			},
		}},
	})
}
