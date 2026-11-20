package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hanna, Ship's Navigator — Legendary Creature — Human Artificer
// {1}{W}{U}, 1/2 (Invasion, #5616):
//
//	"{1}{W}{U}, {T}: Return target artifact or enchantment card from
//	 your graveyard to your hand."
//
// Adun Oakenshield's shape with an artifact-or-enchantment clause: a
// graveyard target clause (your graveyard) and the shared
// returnTargetCardToHand body. With no such card in your graveyard the
// ability has no legal target and can't be activated; a card that
// leaves the graveyard in response is no longer a legal target, and the
// ability does nothing (CR 608.2b). The tap means summoning sickness
// applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b008c71-e2a0-4f56-83d3-655c5dd58e0e",
		Name:         "Hanna, Ship's Navigator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{W}{U}, {T}: Return target artifact or enchantment card from your graveyard to your hand.",
			Cost:  effects.Plus(effects.ManaCost("{1}{W}{U}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target artifact or enchantment card in your graveyard",
				effects.YouOwn(), effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect: effects.ReturnTargetCardToHand,
		}},
	})
}
