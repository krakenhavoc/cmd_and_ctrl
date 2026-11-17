package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jolrael's Centaur — Creature — Centaur Archer {1}{G}{G}, 2/2 (Mirage,
// #5633):
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)"
//
// Shroud is a printed keyword the engine enforces at the targeting choke
// point for every spell and ability, its controller's too (CR 702.18a).
// Flanking is flanking.go's trigger with the declared keyword (#5757,
// CR 702.25a); the trigger targets nothing, so shroud doesn't stop it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "39b699e1-79c6-4960-97dc-d58574181bbe",
		Name:            "Jolrael's Centaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud", game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{RetiredFlankingRow("Jolrael's Centaur")},
	})
}
