package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mogg Salvage — Instant {2}{R} (Nemesis, #5618):
//
//	"If an opponent controls an Island and you control a Mountain, you
//	 may cast this spell without paying its mana cost.
//	 Destroy target artifact."
//
// Sivvi's Ruse's free cast (the Nemesis cycle's condition, read off
// effective subtypes as you cast it: your own Island doesn't count) on
// Shatter's body. Without the two lands it costs {2}{R}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0923ad85-2ea4-4463-8f71-b7d0a162b486",
		Name:         "Mogg Salvage",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact", Artifact()),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "free",
			Label:     "Cast without paying its mana cost (an opponent controls an Island and you control a Mountain)",
			ManaCost:  "",
			Condition: AnOpponentControlsAAndYouControlA("Island", "Mountain"),
		}},
		OnResolve: destroyTheTargetPermanent,
	})
}
