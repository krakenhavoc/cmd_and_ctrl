package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Varis, Silverymoon Ranger — Legendary Creature — Human Elf Ranger
// {1}{G}{G}, 3/3:
//
//	"Reach, ward {1}
//	 Whenever you cast a creature or planeswalker spell, venture into
//	 the dungeon. This ability triggers only once each turn.
//	 Whenever you complete a dungeon, create a 2/2 green Wolf creature
//	 token."
//
// "Only once each turn" is the turn tally of this object's triggers
// (Welcoming Vampire's guard).
//
// No simplification.
func init() {
	label := ventureLabel("Varis, Silverymoon Ranger")
	Register(Spec{
		OracleID:        "bd751b25-9ea7-4ce4-b532-48be281c6da9",
		Name:            "Varis, Silverymoon Ranger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Varis, Silverymoon Ranger — ward {1}"),
			On(game.EventCast, AllOf(
				YouCast(Or(Creature(), Planeswalker())),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return source != nil && !b11TriggeredThisTurn(g, source.InstanceID, label)
				},
			), label, Venture),
			WheneverYouCompleteADungeon("Varis, Silverymoon Ranger — create a 2/2 green Wolf creature token",
				Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}
