package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Mechanic — Creature — Human Artificer {2}{W}, 3/3 (Aetherdrift,
// #4464):
//
//	"{3}{W}: Put a +1/+1 counter on target Mount or Vehicle."
//
// High Sentinels of Arashin's repeatable counter with the target any Mount
// or Vehicle on the battlefield, whoever controls it, crewed or not. A
// Vehicle's counter changes its power and toughness only while it is a
// creature (CR 301.7a), but the counter stays on it. A target gone by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b1d79b7-a812-4f48-9be0-789294da3b56",
		Name:         "Daring Mechanic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Put a +1/+1 counter on target Mount or Vehicle.",
			Cost:    effects.ManaCost("{3}{W}"),
			Targets: effects.TargetPermanent("target Mount or Vehicle", effects.Or(effects.HasSubtype("Mount"), effects.HasSubtype("Vehicle"))),
			Effect:  effects.PlusOneCounterOnTheTarget,
		}},
	})
}
