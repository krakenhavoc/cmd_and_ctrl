package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpha Authority — Enchantment — Aura {1}{G} (Gatecrash, S82 OLD b407,
// #5565):
//
//	"Enchant creature
//	 Enchanted creature has hexproof and can't be blocked by more than one
//	 creature."
//
// A layer 6 hexproof grant and Bristling Boar's block cap (MaxBlockers) read
// off the attached creature (OnAttached), so both follow the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "538fa2e0-d225-4232-b243-38868b695e26",
		Name:         "Alpha Authority",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("hexproof")},
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnAttached(), 1)},
	})
}
