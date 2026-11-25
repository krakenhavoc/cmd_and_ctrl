package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctum Spirit — Creature — Spirit {3}{W}, 3/2 (Dominaria, S82 OLD, #5530):
//
//	"Lifelink
//	 Discard a historic card: This creature gains indestructible until end of
//	 turn. (Artifacts, legendaries, and Sagas are historic.)"
//
// The discard is the cost, paid on announcement (CR 602.2b) from a hand card that
// is an artifact, a legendary or a Saga; it can be activated any number of times
// (it costs no mana and no tap). The gain goes to this permanent only.
//
// No simplification.
func init() {
	const label = "Sanctum Spirit — gains indestructible until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "1e640846-8f50-4a94-b7aa-f228124290d3",
		Name:            "Sanctum Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a historic card: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.DiscardCardsMatching(1, "a historic card", effects.B09IsHistoric),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("indestructible", label),
		}},
	})
}
