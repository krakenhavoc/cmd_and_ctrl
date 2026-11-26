package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ominous Seas — Enchantment {1}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever you draw a card, put a foreshadow counter on this enchantment.
//	 Remove eight foreshadow counters from this enchantment: Create an 8/8
//	 blue Kraken creature token.
//	 Cycling {2}"
//
// One counter per card drawn, so a three-card draw is three triggers. The
// cost removes eight counters, so it cannot be paid with fewer, and each
// payment makes one token; several can be paid in a row when it holds
// sixteen. It is not a mana ability and uses the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58ea9b0a-1bef-4ccb-acb7-f03c8daf3df3",
		Name:         "Ominous Seas",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Remove eight foreshadow counters from this enchantment: Create an 8/8 blue Kraken creature token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    effects.RemoveCountersFromThis("foreshadow", 8),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("8/8 blue Kraken"), N: 1}),
			},
			effects.Cycling("{2}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Ominous Seas — put a foreshadow counter on it", effects.Do(effects.CounterOnThis{Kind: "foreshadow", N: 1})),
		},
	})
}
