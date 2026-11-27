package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lurking Informant — Creature — Human Rogue {1}{U/B}, 1/2 (Ravnica: City of
// Guilds, S82 OLD batch 574, #5600):
//
//	"({U/B} can be paid with either {U} or {B}.)
//	 {2}, {T}: Look at the top card of target player's library. You may put
//	 that card into that player's graveyard."
//
// Only the controller looks (CR 701.20). The card goes to its owner's
// graveyard if the controller says so, and otherwise stays on top
// (S82OldB574LurkingInformantLook).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ad09b75-6d7a-46c7-95a2-d7f90995fdd1",
		Name:         "Lurking Informant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Look at the top card of target player's library. You may put that card into that player's graveyard.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82OldB574LurkingInformantLook,
		}},
	})
}
