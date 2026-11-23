package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quartzwood Crasher — Creature — Dinosaur Beast {2}{R}{R}{G}, 6/6 (Lost
// Caverns of Ixalan Commander):
//
//	"Trample
//	 Whenever one or more creatures you control with trample deal combat
//	 damage to a player, create an X/X green Dinosaur Beast creature token
//	 with trample, where X is the amount of damage those creatures dealt to
//	 that player."
//
// One trigger per player damaged at the same time (CR 603.2c): the batch
// total is the sum of the combat damage every trampling creature of yours
// dealt to that player together, so two tramplers dealing 3 and 4 make one
// 7/7, and damage to a second player is another trigger. The token's X is
// fixed as it is created (CR 107.3).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "178a79e2-64fd-4777-95fe-98b7834a642a",
		Name:            "Quartzwood Crasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:    []game.EventKind{game.EventDealDamage},
			AppliesTo:  effects.Cmdr2B08TramplerYouControlDealtCombatDamageToAPlayer,
			Key:        "Quartzwood Crasher — create an X/X green Dinosaur Beast with trample, X the damage dealt to that player",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.Cmdr2B08CreateBeastForTheDamage,
		}},
	})
}
