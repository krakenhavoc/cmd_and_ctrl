package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "933e4e8a-01ea-4f58-8f5e-4f6c43f4e105",
		Name:            "Platypus-Bear",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			avB18CanAttackAsThoughNoDefenderWhile(
				"As long as there is a Lesson card in your graveyard, this creature can attack as though it didn't have defender.",
				func(g *game.Game, source *game.Card) bool { return avB4LessonsInGraveyard(g, source.Controller) >= 1 }),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Platypus-Bear — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
