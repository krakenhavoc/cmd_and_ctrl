package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sneaking Guide — Creature — Goblin Rogue {R}, 1/1 (Zendikar Rising,
// #5508):
//
//	"{2}, {T}: Target creature with power 2 or less can't be blocked this
//	 turn."
//
// Access Tunnel's ability at power 2: the power limit is a targeting
// restriction (a target whose power has risen above 2 by resolution is
// illegal, CR 608.2b), and the effect is a block restriction until end of
// turn (CR 509.1b).
//
// No simplification.
func init() {
	const label = "Sneaking Guide — can't be blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "3eabd27c-bd65-479a-9743-e90e4273c016",
		Name:         "Sneaking Guide",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
			Effect:  effects.OldB095TargetCantBeBlocked(label),
		}},
	})
}
