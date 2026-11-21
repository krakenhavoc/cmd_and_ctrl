package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starport Security — Artifact Creature — Robot Soldier {W}, 1/1 (Edge
// of Eternities, #4461):
//
//	"{3}{W}, {T}: Tap another target creature. This ability costs {2}
//	 less to activate if you control a creature with a +1/+1 counter on
//	 it."
//
// Thassa, Deep-Dwelling's "tap another target creature" with Esquire of
// the King's discount: the ability's own cost clause (CR 601.2f via
// 602.2b), {2} off the generic part once, so {1}{W} and {T}. The
// discount reads every creature its activator controls, Starport
// Security itself included (YouControlA would skip the source).
//
// No simplification.
func init() {
	const label = "{3}{W}, {T}: Tap another target creature. This ability costs {2} less to activate if you control a creature with a +1/+1 counter on it."
	effects.Register(effects.Spec{
		OracleID:     "5454dddf-c6e5-4b0d-b96d-4904c13606dc",
		Name:         "Starport Security",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			CostModifiers: []game.CostModifier{
				effects.CostsLess(2, "This ability costs {2} less to activate if you control a creature with a +1/+1 counter on it.",
					starportSecurityYouControlACountered),
			},
			Effect: effects.TapTheFirstLegalTarget,
		}},
	})
}

// starportSecurityYouControlACountered is "if you control a creature with
// a +1/+1 counter on it", over the whole battlefield: the source counts.
func starportSecurityYouControlACountered(q game.CostQuery) bool {
	if q.Game == nil || q.Game.Battlefield == nil {
		return false
	}
	countered := effects.CreatureWithCounter(game.CounterPlusOne)
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Controller && countered(q.Game, q.Controller, c) {
			return true
		}
	}
	return false
}
