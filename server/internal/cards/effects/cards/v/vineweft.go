package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vineweft — Enchantment — Aura {G} (Magic 2015, #5556):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 {4}{G}: Return this card from your graveyard to your hand."
//
// Bestial Bloodline at +1/+1: the pump is a layer-7c static on the host,
// and the return is an activated ability that works only from the
// graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4a2d9eb-7f2b-4ee6-b38e-cf6e3a837b49",
		Name:         "Vineweft",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
