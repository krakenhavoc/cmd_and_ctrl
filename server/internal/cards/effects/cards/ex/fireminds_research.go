package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Firemind's Research — Enchantment {U}{R} (Guilds of Ravnica):
//
//	"Whenever you cast an instant or sorcery spell, put a charge counter on
//	 this enchantment.
//	 {1}{U}, Remove two charge counters from this enchantment: Draw a card.
//	 {1}{R}, Remove five charge counters from this enchantment: It deals 5
//	 damage to any target."
//
// The counters are put on as the cast trigger resolves, above the spell. Each
// removal is a cost paid at activation (Arcane Spyglass), so a Research without
// the counters cannot be activated, and the damage is Borborygmos Enraged's
// source-dealt damage at the announced target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edac6c23-cff5-4ee2-9227-756b1dc6ab0c",
		Name:         "Firemind's Research",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), "Firemind's Research — put a charge counter on it", effects.PutACounterOnThis(game.CounterCharge)),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}, Remove two charge counters from this enchantment: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.RemoveCountersFromThis(game.CounterCharge, 2)),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{1}{R}, Remove five charge counters from this enchantment: It deals 5 damage to any target.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 5)),
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.RemoveCountersFromThis(game.CounterCharge, 5)),
				Targets: effects.TargetAny(),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(5),
			},
		},
	})
}
