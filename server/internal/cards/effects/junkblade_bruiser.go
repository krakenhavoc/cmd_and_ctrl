package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Junkblade Bruiser — Creature — Raccoon Berserker {3}{R/G}{R/G}, 4/5
// (Bloomburrow, #4467):
//
//	"Trample
//	 Whenever you expend 4, this creature gets +2/+1 until end of turn.
//	 (You expend 4 as you spend your fourth total mana to cast spells
//	 during a turn.)"
//
// Bakersbane Duo's expend trigger (WheneverYouExpend, expend.go) with a
// +2/+1 pump; thisCreatureUntilEOT does nothing if the Bruiser has left
// by the time the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a2fbdd3-4940-4bbf-ae19-f6d00d7ff294",
		Name:            "Junkblade Bruiser",
		ReadsManaSpent:  game.ReadsYourCasts,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouExpend(4, "Junkblade Bruiser — this creature gets +2/+1 until end of turn",
				thisCreatureUntilEOT("Junkblade Bruiser — +2/+1", 2, 1)),
		},
	})
}
