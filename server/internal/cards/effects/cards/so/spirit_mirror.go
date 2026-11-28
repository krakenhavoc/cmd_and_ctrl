package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spirit Mirror — {2}{W}{W} Enchantment (Tempest, S82 OLD batch 769, #5628):
//
//	"At the beginning of your upkeep, if there are no Reflection tokens on the
//	 battlefield, create a 2/2 white Reflection creature token.
//	 {0}: Destroy target Reflection."
//
// The upkeep trigger has an intervening "if" (CR 603.4): no Reflection TOKEN on the
// battlefield, any controller's, checked as the upkeep begins and again as the
// ability resolves. A Reflection that is not a token does not stop it. The {0}
// ability may be activated by this enchantment's controller only, against any
// Reflection, a token or not, and destroys it if it is still one as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42e041c0-ffbc-4bf8-a5aa-f0d14aad7112",
		Name:         "Spirit Mirror",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("there are no Reflection tokens on the battlefield", effects.S82OldB769NoPermanentOfSubtype("Reflection", true),
				effects.AtYourUpkeep("Spirit Mirror — create a 2/2 white Reflection creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Reflection"), N: 1}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: Destroy target Reflection.",
			Cost:    effects.ManaCost("{0}"),
			Targets: effects.TargetPermanent("target Reflection", effects.Subtype("Reflection")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
			},
		}},
	})
}
