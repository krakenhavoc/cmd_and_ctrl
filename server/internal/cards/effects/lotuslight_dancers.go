package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lotuslight Dancers — Creature — Zombie Bard {2}{B}{G}{U}, 3/6 (Tarkir:
// Dragonstorm, #4463):
//
//	"Lifelink
//	 When this creature enters, search your library for a black card, a
//	 green card, and a blue card. Put those cards into your graveyard, then
//	 shuffle."
//
// One search for up to three cards, Buried Alive's shape, whose set rule
// (searchRuleOneCardPerColor, std_b11_helpers.go) makes every card found
// stand for its own colour: a black-green-blue card may be the black one
// and the green and blue ones still need cards of their own. Any of the
// three may be left unfound (CR 701.23b, a stated quality).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "83c8f1ca-194e-4cb6-9d5f-029133c1edbc",
		Name:            "Lotuslight Dancers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Purpose:         game.Purpose{SelfMillTutor: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lotuslight Dancers — search for a black card, a green card, and a blue card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.HasColor("B") || c.HasColor("G") || c.HasColor("U")
					},
					Dest:    game.ZoneGraveyard,
					Limit:   3,
					Shuffle: true,
					Rule:    searchRuleOneCardPerColor.WithN(sb11ColorMask("B", "G", "U")),
					Reason:  "Lotuslight Dancers — a black card, a green card, and a blue card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
