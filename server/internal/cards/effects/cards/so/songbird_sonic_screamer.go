package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Songbird, Sonic Screamer — Legendary Creature — Human Villain {1}{B},
// 2/1 (Marvel Super Heroes Commander, #4455):
//
//	"Lifelink
//	 Discard a card: Songbird gains flying until end of turn."
//
// Lifelink rides PrintedKeywords. The discard is a cost the activator
// picks from hand (Aeromoeba's), and the grant is Endling's layer-6
// keyword pinned to the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "46f78ae1-fa51-4ef9-9ffa-8a3ed71a39e4",
		Name:            "Songbird, Sonic Screamer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: Songbird gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.DiscardACard(),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Songbird — flying until end of turn"),
		}},
	})
}
