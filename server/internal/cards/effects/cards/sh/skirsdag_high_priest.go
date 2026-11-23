package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skirsdag High Priest — Creature — Human Cleric {1}{B}, 1/2 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Morbid — {T}, Tap two untapped creatures you control: Create a 5/5
//	 black Demon creature token with flying. Activate only if a creature
//	 died this turn."
//
// The {T} is the symbol (summoning sickness applies, CR 302.6); the two
// creatures are a separate cost that is not the {T} symbol, so a creature
// that arrived this turn may pay it, and the High Priest can't pay for
// itself (CR 118.3). Morbid is an activation condition (any creature,
// anyone's, tokens included; ACreatureDiedThisTurn), not a trigger.
//
// No simplification.
func init() {
	const tapTwo = "two untapped creatures you control"
	effects.Register(effects.Spec{
		OracleID:     "c6064c08-e4e0-49f4-9f0b-55cdecf443af",
		Name:         "Skirsdag High Priest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Morbid — {T}, Tap two untapped creatures you control: Create a 5/5 black Demon creature token with flying. Activate only if a creature died this turn.",
			Cost: game.AbilityCost{
				Tap: true,
				TapOthers: &game.TapOthersCost{
					Count:         2,
					Filter:        effects.TargetPermanent(tapTwo, effects.Creature()),
					ExcludeSource: true,
					Label:         tapTwo,
				},
			},
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Condition: effects.ACreatureDiedThisTurn(),
			Effect:    effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 black Demon with flying"), N: 1}),
		}},
	})
}
