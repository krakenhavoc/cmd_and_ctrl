package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iridescent Tiger — Creature — Cat {4}{R}, 3/4 (Tarkir: Dragonstorm,
// #4463):
//
//	"When this creature enters, if you cast it, add {W}{U}{B}{R}{G}."
//
// "If you cast it" is Geological Appraiser's intervening if (CR 603.4),
// read off the permanent's cast provenance as it triggers and again as it
// resolves: a reanimated or flickered Tiger adds nothing. The mana is
// Priest of Gix's: a triggered ability that is not a mana ability (CR
// 605.1b), so it uses the stack and the five mana land as it resolves, to
// be spent in that step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce09c5b0-9dd4-4f73-9f3f-dfbd829813e4",
		Name:         "Iridescent Tiger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.GeologicalAppraiserWasCast), "Iridescent Tiger — add {W}{U}{B}{R}{G}",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if ctx.CastProvenance().FromZone == "" {
						return nil
					}
					return effects.AddMana{Produced: "{W}{U}{B}{R}{G}"}.Apply(ctx)
				}),
		},
	})
}
