package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "58bd67a8-1833-4827-aa33-1c141568f481",
		Name:         "Mystifying Maze",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Exile target attacking creature an opponent controls. At the beginning of the next end step, return it to the battlefield tapped under its owner's control.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Targets: TargetCreature("target attacking creature an opponent controls", attackingCreature(), OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return exileTargetsThenScheduleReturnBody(NewContext(g, item),
					"Mystifying Maze — return the exiled creature tapped", returnExiledToOwnersTappedBody)
			},
		}},
	})
}
