package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warchief Giant — Creature — Giant Warrior {3}{R}{R}, 5/3 (Commander
// 2015, #5546):
//
//	"Haste
//	 Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)"
//
// Myriad is the shared trigger (myriad.go, CR 702.116); haste is printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "589abb40-04b4-4dbe-9c2a-446b6c3075d1",
		Name:            "Warchief Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{GrantedMyriad("Warchief Giant")},
	})
}
