package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of Cloverdell — Creature — Treefolk Shaman {5}{G}{G}, 4/5 (Lorwyn,
// S82 OLD, #5593):
//
//	"When this creature enters, create three 1/1 white Kithkin Soldier creature
//	 tokens.
//	 {G}, Sacrifice a Kithkin: You gain 1 life."
//
// Cloudgoat Ranger's three Kithkin Soldiers and Seething Pathblazer's
// sacrifice over effective creature types, with {G} beside it. Any Kithkin
// you control pays, the Guardian itself only if it is one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27d6b1fb-ce53-45d6-a601-9f5b9ee7ab2f",
		Name:         "Guardian of Cloverdell",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Guardian of Cloverdell — create three 1/1 white Kithkin Soldier tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Kithkin Soldier"), N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice a Kithkin: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				Mana:           "{G}",
				SacrificeOther: effects.SacrificeSpec("a Kithkin", effects.OfCreatureType("Kithkin")),
			},
			Effect: effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}
