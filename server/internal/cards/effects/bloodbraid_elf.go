package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodbraid Elf — Creature — Elf Berserker {2}{R}{G}, 3/2:
//
//	"Haste
//	 Cascade (When you cast this spell, exile cards from the top of
//	 your library until you exile a nonland card that costs less. You
//	 may cast it without paying its mana cost. Put the exiled cards on
//	 the bottom in a random order.)"
//
// The plain reading of the keyword, and the one to check the engine
// against: a four-mana spell cascading into anything with mana value
// three or less.
func init() {
	Register(Spec{
		OracleID:        "3f0c9466-5ab9-4205-a84f-b4b27b5a678e",
		Name:            "Bloodbraid Elf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{Cascade()},
	})
}
