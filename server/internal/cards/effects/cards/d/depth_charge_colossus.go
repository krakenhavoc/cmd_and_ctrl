package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Depth Charge Colossus — Artifact Creature — Dreadnought {9}, 9/9 (The
// Brothers' War, #5656): "Prototype {4}{U}{U} — 6/6", "This creature
// doesn't untap during your untap step.", "{3}: Untap this creature."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "f376f4c6-4526-4ae3-a84a-e8ea50fe78b6",
		Name:                  "Depth Charge Colossus",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{3}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
