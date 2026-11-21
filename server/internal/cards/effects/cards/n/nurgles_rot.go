package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nurgle's Rot — Enchantment — Aura {B} (Warhammer 40,000 Commander, S82 GAP
// batch 196, #5948):
//
//	"Enchant creature an opponent controls
//	 When enchanted creature dies, return this card to its owner's hand and you
//	 create a 1/3 black Demon creature token named Plaguebearer of Nurgle."
//
// Angelic Destiny's dies trigger (harvested while the Aura is still attached): it
// returns the card only if it is still the Aura the state-based action put into
// the graveyard, and creates the token for the Aura's controller either way.
// The enchant clause is an opponent's creature, judged again every turn
// (CR 704.5m): if the creature changes hands to its enchanter, the Aura falls off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b01dcff-0639-4ed0-bd9e-f28429cb803d",
		Name:         "Nurgle's Rot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.OpponentControls()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Nurgle's Rot — return this card to its owner's hand and create a 1/3 black Demon creature token named Plaguebearer of Nurgle",
				effects.GapB196NurglesRotDies),
		},
	})
}
