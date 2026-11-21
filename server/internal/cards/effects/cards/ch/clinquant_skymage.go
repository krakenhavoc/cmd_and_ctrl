package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clinquant Skymage — Creature — Bird Wizard {3}{U}, 1/1 (Foundations,
// #4465):
//
//	"Flying
//	 Whenever you draw a card, put a +1/+1 counter on this creature."
//
// Data, New Life's draw trigger, once per card drawn. Flying rides
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d9a3207a-a330-4517-967e-f37c9a9f1355",
		Name:            "Clinquant Skymage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Clinquant Skymage — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
