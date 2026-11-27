package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortis Dogs — Creature — Phyrexian Dog {3}{B}, 2/2:
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn.
//	 When this creature dies, target player loses life equal to its power."
//
// The death trigger reads the Dog's last-known power (CR 603.10), so the
// attack bonus counts when it died attacking.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a89994d0-2110-4ca0-abfc-5dd232981acf",
		Name:         "Mortis Dogs",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Mortis Dogs — it gets +2/+0 until end of turn",
				effects.ThisCreatureUntilEOT("Mortis Dogs — +2/+0", 2, 0)),
			effects.EerieDiesWithPower("Mortis Dogs — target player loses life equal to its power",
				effects.TargetPlayer("target player"),
				func(g *game.Game, item *game.StackItem) error {
					n := item.Params.Amount
					if n <= 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
						}
					}
					return nil
				}),
		},
	})
}
