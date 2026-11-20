package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0923ad85-2ea4-4463-8f71-b7d0a162b486",
		Name:         "Mogg Salvage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "free",
			Label:     "Cast without paying its mana cost (an opponent controls an Island and you control a Mountain)",
			ManaCost:  "",
			Condition: effects.AnOpponentControlsAAndYouControlA("Island", "Mountain"),
		}},
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
