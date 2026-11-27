package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myr Prototype — Artifact Creature — Myr {5}, 2/2 (Mirrodin, S82 OLD,
// #5606):
//
//	"At the beginning of your upkeep, put a +1/+1 counter on this
//	 creature.
//	 This creature can't attack or block unless you pay {1} for each +1/+1
//	 counter on it."
//
// The upkeep trigger is Passageway Seer's counter on itself. The second
// sentence is one tax on each side (#6191): {N} as it is declared an
// attacker (CR 508.1h) and {N} as it is declared a blocker (CR 509.1d), N
// the +1/+1 counters on it as the declaration is made. With no counters,
// no tax.
//
// No simplification.
func init() {
	const label = "This creature can't attack or block unless you pay {1} for each +1/+1 counter on it."
	effects.Register(effects.Spec{
		OracleID:     "496e0fbb-cb01-41d6-9c56-23e3ae2bb4f7",
		Name:         "Myr Prototype",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Myr Prototype — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: "+1/+1", N: 1})),
		},
		AttackTaxes: []game.AttackTax{effects.CantAttackUnlessYouPayPerCounter("+1/+1", label)},
		BlockTaxes:  []game.BlockTax{effects.CantBlockUnlessYouPayPerCounter("+1/+1", label)},
	})
}
