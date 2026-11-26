package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Might Makes Right — Enchantment {5}{R} (Magic 2015, #5556):
//
//	"At the beginning of combat on your turn, if you control each creature on
//	 the battlefield with the greatest power, gain control of target creature an
//	 opponent controls until end of turn. Untap that creature. It gains haste
//	 until end of turn."
//
// An intervening "if" (CR 603.4), checked as combat begins and again as the
// ability resolves, over every creature's current power: each one tied for the
// greatest is yours. The steal is Take for a Ride's (StdB31Threaten): the
// control change, the untap and the haste in printed order, for the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec73863f-417b-45d7-a6b7-97a69ce0196a",
		Name:         "Might Makes Right",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control each creature on the battlefield with the greatest power",
				effects.S82OldB359YouControlEachCreatureWithTheGreatestPower,
				effects.Targeting(
					effects.AtBeginningOfYourCombat("Might Makes Right — gain control of target creature an opponent controls until end of turn, untap it, it gains haste",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							t, ok := effects.FirstLegalTarget(ctx)
							if !ok {
								return nil
							}
							return effects.StdB31Threaten(ctx, t, "Might Makes Right")
						}),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()))),
		},
	})
}
