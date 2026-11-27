package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shield of Duty and Reason — Enchantment — Aura {W} (Apocalypse, S82 OLD,
// #5614):
//
//	"Enchant creature
//	 Enchanted creature has protection from green and from blue."
//
// Two real protections (CR 702.16) granted to the host, as the Swords grant
// them: it can't be blocked by, damaged by, enchanted or equipped by, or
// targeted by green or blue sources.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c64ab87-1da1-4b07-a3f2-73631ae3d501",
		Name:         "Shield of Duty and Reason",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("protection from green", "protection from blue")},
	})
}
