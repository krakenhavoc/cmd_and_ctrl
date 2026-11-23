package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystifying Maze — Land:
//
//	"{T}: Add {C}.
//	 {4}, {T}: Exile target attacking creature an opponent controls. At the
//	 beginning of the next end step, return it to the battlefield tapped
//	 under its owner's control."
//
// Morningtide's Light's delayed return (returnExiledToOwnersTappedBody):
// the exile and the scheduled return are one ability; the creature comes
// back as a new object, tapped, under its owner's control (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58bd67a8-1833-4827-aa33-1c141568f481",
		Name:         "Mystifying Maze",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Exile target attacking creature an opponent controls. At the beginning of the next end step, return it to the battlefield tapped under its owner's control.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking creature an opponent controls", effects.IsAttackingCreature(), effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTargetsThenScheduleReturnBody(effects.NewContext(g, item),
					"Mystifying Maze — return the exiled creature tapped", effects.ReturnExiledToOwnersTappedBody)
			},
		}},
	})
}
