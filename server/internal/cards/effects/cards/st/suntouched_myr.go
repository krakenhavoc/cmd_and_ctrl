package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suntouched Myr — Artifact Creature — Myr {3}, 0/0:
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.)"
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). Cast off colourless mana it enters as a 0/0 and
// dies, as printed. Spec.WantsDistinctColors has the cast gate spread
// the payment across colours.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "75d4f61a-8276-43ad-99e6-0d756a597e4b",
		Name:                "Suntouched Myr",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        effects.CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
	})
}
