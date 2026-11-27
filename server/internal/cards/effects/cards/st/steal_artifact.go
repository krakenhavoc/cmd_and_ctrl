package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steal Artifact — Enchantment — Aura {2}{U}{U} (Eighth Edition, S82 older sets
// batch 617, #5607):
//
//	"Enchant artifact
//	 You control enchanted artifact."
//
// Mind Control on an artifact: the layer 2 control effect lasts as long as the
// Aura is attached (Domineer's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd8ae9f2-edac-473a-8846-c08219e617c3",
		Name:         "Steal Artifact",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact", effects.Artifact()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
	})
}
