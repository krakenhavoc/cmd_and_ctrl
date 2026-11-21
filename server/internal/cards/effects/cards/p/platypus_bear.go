package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Platypus-Bear — Creature — Platypus Bear {1}{G/U}, 2/3 (Avatar: The Last
// Airbender, #4459):
//
//	"Defender
//	 When this creature enters, mill two cards.
//	 As long as there is a Lesson card in your graveyard, this creature
//	 can attack as though it didn't have defender."
//
// The waiver is Starfleet Spanner Jockey's (avB18CanAttackAsThoughNoDefenderWhile),
// covering every target, and on only while its controller's graveyard
// holds a Lesson card (avB4LessonsInGraveyard). The keyword stays. A
// Platypus-Bear that has attacked stays attacking if the Lesson then
// leaves (the ruling): the waiver is asked only at declaration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "933e4e8a-01ea-4f58-8f5e-4f6c43f4e105",
		Name:            "Platypus-Bear",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			effects.AvB18CanAttackAsThoughNoDefenderWhile(
				"As long as there is a Lesson card in your graveyard, this creature can attack as though it didn't have defender.",
				func(g *game.Game, source *game.Card) bool {
					return effects.AvB4LessonsInGraveyard(g, source.Controller) >= 1
				}),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Platypus-Bear — mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
	})
}
