package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spitting Hydra — Creature — Hydra {3}{R}{R}, 0/0 (Stronghold, S82 OLD batch
// 748, #5627):
//
//	"This creature enters with four +1/+1 counters on it.
//	 {1}{R}, Remove a +1/+1 counter from this creature: It deals 1 damage to
//	 target creature."
//
// Monoskelion's entry replacement (CR 614.1c, so the Hydra is never a 0/0 on
// the battlefield) and counter-removal cost (paid as the ability is announced,
// CR 602.2b, so a response cannot spend it twice), aimed at a creature. The
// Hydra is the damage source as it last existed (CR 608.2h), and the target is
// read again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67a7bfa4-5004-4838-874d-03fa11f051d0",
		Name:         "Spitting Hydra",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB031SelfEntersWithCounters(game.CounterPlusOne, 4)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Remove a +1/+1 counter from this creature: It deals 1 damage to target creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.B33DamageChosenTargetFromSource(1),
		}},
	})
}
