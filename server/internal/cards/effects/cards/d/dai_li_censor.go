package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dai Li Censor — Creature — Human Soldier Advisor {1}{B}, 2/1 (Avatar:
// The Last Airbender Eternal, #5075):
//
//	"{1}, Sacrifice another creature: This creature gets +2/+2 until end
//	 of turn. Activate only once each turn."
//
// The sacrifice is part of the cost (CR 602.2c), so the Censor itself is
// never a legal payment, and the once-each-turn limit is the activation
// tally (OncePerTurnActivation). A Censor that left the battlefield and
// came back is a new object the pump does not reach (CR 400.7).
//
// No simplification.
func init() {
	const label = "{1}, Sacrifice another creature: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "f13a8621-4d88-4c1f-8b6e-80f8ef383cd2",
		Name:         "Dai Li Censor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:      effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, ctx.Source()) {
					return nil
				}
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Dai Li Censor — +2/+2"}.Apply(ctx)
			},
		}},
	})
}
