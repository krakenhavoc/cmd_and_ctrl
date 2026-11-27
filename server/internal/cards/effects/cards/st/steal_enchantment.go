package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steal Enchantment — Enchantment — Aura {U}{U} (Tempest, S82 OLD, #5628):
//
//	"Enchant enchantment
//	 You control enchanted enchantment."
//
// Steal Artifact's layer 2 control effect, on an enchantment: it lasts as long as
// the Aura is attached and control goes back when it is not (CR 613.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "567a3fa7-5ca9-4812-b81d-4953c401c83d",
		Name:         "Steal Enchantment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant enchantment", effects.Enchantment()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
	})
}
