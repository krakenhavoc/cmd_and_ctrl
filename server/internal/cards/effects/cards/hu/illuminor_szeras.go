package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illuminor Szeras — Legendary Artifact Creature — Necron {2}{B}, 3/3
// (Warhammer 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"Secrets of the Soul — {T}, Sacrifice another creature: Add an amount
//	 of {B} equal to the sacrificed creature's mana value."
//
// A mana ability (CR 605.1a) whose output is read from the payment, Slobad's
// shape: the sacrificed creature's mana value as it last existed on the
// battlefield (CR 608.2h), X zero. A token, with no mana cost, adds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10f274f9-9044-447b-81c9-d4906ad9e6be",
		Name:         "Illuminor Szeras",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				SacrificeOther: effects.SacrificeAnotherN(1, "another creature", effects.Creature()).SacrificeOther,
			},
			Answers:         game.AnswerSacOutlet,
			ProducedForPaid: gapB188SzerasProduced,
			Label:           "{T}, Sacrifice another creature: Add an amount of {B} equal to the sacrificed creature's mana value.",
		}},
	})
}

func gapB188SzerasProduced(g *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	if len(paid.SacrificedObjects) == 0 {
		return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"B"}, N: 1}}
	}
	mv := 0
	if info, ok := g.PermanentForEffect(paid.SacrificedObjects[0]); ok {
		mv = info.ManaValue
	}
	return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"B"}, N: mv}}
}
