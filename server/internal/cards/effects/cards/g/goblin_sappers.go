package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Sappers — Creature — Goblin {1}{R}, 1/1 (Ice Age, S82 OLD #5636):
//
//	"{R}{R}, {T}: Target creature you control can't be blocked this turn.
//	 Destroy it and this creature at end of combat.
//	 {R}{R}{R}{R}, {T}: Target creature you control can't be blocked this turn.
//	 Destroy it at end of combat."
//
// Access Tunnel's unblockability and Silent Assassin's delayed destruction at the
// end of combat (CR 603.7), one delayed trigger per creature destroyed; the
// Sappers' own is made only while it is still the object that activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e095761-cdae-4ab2-b2d3-099b0b39600a",
		Name:         "Goblin Sappers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}{R}, {T}: Target creature you control can't be blocked this turn. Destroy it and this creature at end of combat.",
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    effects.Plus(effects.ManaCost("{R}{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.SappersUnblockableB832(true),
			},
			{
				Label:   "{R}{R}{R}{R}, {T}: Target creature you control can't be blocked this turn. Destroy it at end of combat.",
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    effects.Plus(effects.ManaCost("{R}{R}{R}{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.SappersUnblockableB832(false),
			},
		},
	})
}
