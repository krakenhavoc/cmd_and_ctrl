package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adun Oakenshield — Legendary Creature — Human Knight {B}{R}{G}, 1/2
// (Legends, #5639):
//
//	"{B}{R}{G}, {T}: Return target creature card from your graveyard to
//	 your hand."
//
// A graveyard target clause (your graveyard, a creature card) and the
// shared returnTargetCardToHand body. With no creature card in your
// graveyard the ability has no legal target and can't be activated; a
// card that leaves the graveyard in response is no longer a legal
// target, and the ability does nothing (CR 608.2b). The tap means
// summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aefd49a4-ecec-42cb-b882-e5db7f81eff8",
		Name:         "Adun Oakenshield",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{R}{G}, {T}: Return target creature card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{B}{R}{G}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
			Effect:  effects.ReturnTargetCardToHand,
		}},
	})
}
