package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ebon Praetor — Creature — Avatar Praetor {4}{B}{B}, 5/5 (Fallen Empires, S82
// OLD b853, #5637):
//
//	"First strike, trample
//	 At the beginning of your upkeep, put a -2/-2 counter on this creature.
//	 Sacrifice a creature: Remove a -2/-2 counter from this creature. If the
//	 sacrificed creature was a Thrull, put a +1/+0 counter on this creature.
//	 Activate only during your upkeep and only once each turn."
//
// "-2/-2" and "+1/+0" are power/toughness counters the layer pass already
// reads (CR 122.1a). The sacrificed creature is read as it last existed
// (CR 608.2h), so a Thrull that an effect made something else is judged as it
// stood. The sacrifice is the cost, so the Praetor may eat itself; the
// counters then go to nothing.
//
// No simplification.
func init() {
	const label = "Sacrifice a creature: Remove a -2/-2 counter from this creature. If the sacrificed creature was a Thrull, put a +1/+0 counter on this creature. Activate only during your upkeep and only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "0501af1f-c67c-4cb1-aced-4be426c83316",
		Name:            "Ebon Praetor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Ebon Praetor — put a -2/-2 counter on this creature", effects.PutACounterOnThis("-2/-2")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerSurvive},
			Cost:      effects.SacrificeACreature(),
			Condition: effects.AllConditions(effects.DuringYourUpkeep(), effects.OncePerTurnActivation(label)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.AddCounter{Target: ctx.Source(), Kind: "-2/-2", N: -1}).Apply(ctx); err != nil {
					return err
				}
				if info, ok := ctx.SacrificedPermanent(); ok && effects.S82OldB853WasThrull(info) {
					return effects.AddCounter{Target: ctx.Source(), Kind: "+1/+0", N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
