package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Rising — Enchantment {3}{B}{B}:
//
//	"At the beginning of your end step, if you control exactly one creature,
//	 create a 5/5 black Demon creature token with flying."
//
// An intervening "if" (CR 603.4): checked when the step begins and again on
// resolution.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10064a1d-66f4-4018-94c2-c727e7838850",
		Name:         "Demonic Rising",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control exactly one creature",
				Test: effects.YouControlExactlyOneCreature,
			}, effects.AtYourEndStep("Demonic Rising — create a 5/5 black Demon creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 black Demon with flying"), N: 1}))),
		},
	})
}
