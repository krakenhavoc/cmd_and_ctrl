package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fear — Enchantment — Aura {B}{B} (Eighth Edition, S82 older sets batch 617,
// #5607):
//
//	"Enchant creature
//	 Enchanted creature has fear."
//
// Flight's grant with the fear keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "355bbe9b-59bf-470f-8600-410af4c7fe18",
		Name:         "Fear",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("fear")},
	})
}
