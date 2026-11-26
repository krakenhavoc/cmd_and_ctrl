package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vish Kal, Blood Arbiter — Legendary Creature — Vampire {4}{W}{B}{B}, 5/5
// (Commander 2011, S82 older sets batch 454, #5574):
//
//	"Flying, lifelink
//	 Sacrifice a creature: Put X +1/+1 counters on Vish Kal, where X is the
//	 sacrificed creature's power.
//	 Remove all +1/+1 counters from Vish Kal: Target creature gets -1/-1 until
//	 end of turn for each +1/+1 counter removed this way."
//
// The first reads the sacrificed creature's power as it last existed (Greater
// Good's shape); sacrificing Vish Kal itself puts nothing anywhere. The second
// pays "all" as a cost (the number removed reaches the effect, zero allowed).
//
// No simplification.
// vishKalShrinkPurpose declares the second ability as removal whose target is harmed
// until end of turn (#4302).
var vishKalShrinkPurpose = func() game.Purpose {
	p := effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetHarm})
	p.Answers = game.AnswerRemove
	return p
}()

func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba3bfe0c-2bd6-469c-802c-440b7caf7fb5",
		Name:            "Vish Kal, Blood Arbiter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: Put X +1/+1 counters on Vish Kal, where X is the sacrificed creature's power.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.S82OldB454VishKalFeed,
		}, {
			Label:   "Remove all +1/+1 counters from Vish Kal: Target creature gets -1/-1 until end of turn for each +1/+1 counter removed this way.",
			Purpose: vishKalShrinkPurpose,
			Cost:    effects.RemoveAllCountersFromThis(game.CounterPlusOne),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB454VishKalShrink,
		}},
	})
}
