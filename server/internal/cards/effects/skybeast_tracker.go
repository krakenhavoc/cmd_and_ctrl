package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skybeast Tracker — Creature — Giant Archer {3}{G}, 2/4 (Wilds of
// Eldraine, #4471):
//
//	"Reach
//	 Whenever you cast a spell with mana value 5 or greater, create a Food
//	 token."
//
// Reach is the engine's keyword. The cast trigger is Lurking Lizards'
// (WheneverYouCast with ManaValueGE), which reads the spell's mana value
// on the stack, X included (CR 202.3e). The Food is the token table's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2c53a913-8a99-4a15-9c47-e79af026dc05",
		Name:            "Skybeast Tracker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(5), "Skybeast Tracker — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
