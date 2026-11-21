package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nazar, the Velvet Fang — Legendary Creature — Vampire Warlock {3}{B},
// 3/3 (Foundations Jumpstart):
//
//	"Menace
//	 Whenever you gain life, put a feeding counter on Nazar.
//	 Whenever Nazar attacks, you may remove three feeding counters from it.
//	 If you do, you draw three cards and you lose 3 life."
//
// A "feeding" counter is a free-form counter kind, as a corpse counter is.
// One counter per life-gain event, however much was gained. The attack
// trigger is optional (asked as it resolves, CR 608.2d); "if you do" is the
// removal really happening: with fewer than three counters on Nazar when it
// resolves nothing is removed, so nothing is drawn or lost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "beeed1c5-2f51-4e40-87b9-9a6f810ee4cb",
		Name:            "Nazar, the Velvet Fang",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Nazar, the Velvet Fang — put a feeding counter on it", effects.Do(effects.CounterOnThis{Kind: "feeding", N: 1})),
			effects.Optional(effects.WheneverThisAttacks("Nazar, the Velvet Fang — remove three feeding counters; draw three cards and lose 3 life",
				effects.S82BnsB51NazarFeast), "Nazar, the Velvet Fang — remove three feeding counters to draw three cards and lose 3 life?"),
		},
	})
}
