package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voldaren Pariah // Abolisher of Bloodlines — Creature — Vampire Horror
// {3}{B}{B}, 3/3 // Creature — Eldrazi Vampire, 6/5 (Secret Lair Drop):
//
//	Voldaren Pariah: "Flying
//	 Sacrifice three other creatures: Transform this creature.
//	 Madness {B}{B}{B}"
//	Abolisher of Bloodlines: "Flying
//	 When this creature transforms into Abolisher of Bloodlines, target
//	 opponent sacrifices three creatures of their choice."
//
// The front face's transform is an activated ability whose cost is three
// OTHER creatures; madness is the engine's. The back face's trigger fires only
// on the transform into it (CR 712.18: a permanent that enters on that face has
// not transformed), targets an opponent, and that player chooses three
// creatures to sacrifice together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4399c284-e3fe-46b1-aa1f-15a2846858f5",
		Name:            "Voldaren Pariah",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Madness:         "{B}{B}{B}",
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice three other creatures: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerProtect | game.AnswerSurvive},
			Cost:    effects.SacrificeAnotherN(3, "three other creatures", effects.Creature()),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        "4399c284-e3fe-46b1-aa1f-15a2846858f5#1",
		Name:            "Abolisher of Bloodlines",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventTransform, effects.ThisTransformedInto(1),
				"Abolisher of Bloodlines — target opponent sacrifices three creatures of their choice",
				effects.S82OldB022TargetOpponentSacrificesThree),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
