package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bestial Bloodline — Enchantment — Aura {1}{G} (Wilds of Eldraine,
// #4471):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 {4}{G}: Return this card from your graveyard to your hand."
//
// Convenient Target without the suspect trigger: the pump is the
// attached-creature layer 7c modify, and the recursion is an ability of
// the card in the graveyard (Zones: graveyard), so it can be activated
// only while the card is there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3ae6694-6314-42e2-bc12-d0e32d49b347",
		Name:         "Bestial Bloodline",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
