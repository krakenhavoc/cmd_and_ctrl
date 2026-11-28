package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Mutant — Creature — Goblin Mutant {2}{R}{R}, 5/3 (Ice Age, S82 OLD
// batch 851, #5636):
//
//	"Trample
//	 This creature can't attack if defending player controls an untapped
//	 creature with power 3 or greater.
//	 This creature can't block creatures with power 3 or greater."
//
// Mogg Jailer's attack clause with the power bound turned around (an untapped
// creature with power 3 or greater, asked of each target's defending player at
// declaration, CR 508.1c, CR 508.5) and Cyclops Tyrant's block rule over the
// attacker's power (CR 509.1b), both read live through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "540e1371-baa9-412f-a17a-22ea833ea3e8",
		Name:            "Goblin Mutant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.CantAttackIfDefendingPlayerControls(game.PermanentQuery{Types: []string{"creature"}, Untapped: true}.WithPowerAtLeast(3)),
		},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.PowerGE(3))),
				"it can't block creatures with power 3 or greater"),
		},
	})
}
