package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "46f78ae1-fa51-4ef9-9ffa-8a3ed71a39e4",
		Name:            "Songbird, Sonic Screamer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: Songbird gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    DiscardACard(),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Songbird — flying until end of turn"),
		}},
	})
}
