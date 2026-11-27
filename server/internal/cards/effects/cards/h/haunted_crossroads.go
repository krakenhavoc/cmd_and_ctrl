package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haunted Crossroads — Enchantment {2}{B} (Mercadian Masques, S82 OLD batch 699, #5619):
//
//	"{B}: Put target creature card from your graveyard on top of your library."
//
// Mortuary Mire's put (PutChosenTargetOnTopOfLibrary) behind an activation: the
// card goes to the top of its owner's library through the shared exit
// primitive, and a target gone by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9aac98a-6fad-463e-89fa-3ad3a7c7f53a",
		Name:         "Haunted Crossroads",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Put target creature card from your graveyard on top of your library.",
			Cost:    effects.ManaCost("{B}"),
			Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			Effect:  effects.PutChosenTargetOnTopOfLibrary,
		}},
	})
}
