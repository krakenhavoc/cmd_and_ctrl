package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Talons of Wildwood — Enchantment — Aura {1}{G} (Core Set 2019, #5527):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has trample.
//	 {2}{G}: Return this card from your graveyard to your hand."
//
// Bestial Bloodline's graveyard recursion with Frantic Strength's statics:
// the return is an ability of the card in the graveyard (Zones: graveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ff309e3-90e0-40e7-a277-3df057f8f198",
		Name:         "Talons of Wildwood",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
