package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mtenda Griffin — Creature — Griffin {3}{W}, 2/2 (Mirage, S82 OLD batch 812,
// #5633):
//
//	"Flying
//	 {W}, {T}: Return this creature to its owner's hand and return target
//	 Griffin card from your graveyard to your hand. Activate only during your
//	 upkeep."
//
// Shard Phoenix's upkeep-only gate (DuringYourUpkeep) on a tap ability that
// bounces the Griffin itself and returns a Griffin card. Each half is done if
// it can be (CR 608.2b): the Griffin is not bounced if it is a new object
// (CR 400.7), the card is not returned if it left the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "273f0353-4936-491b-82da-6d4d7f3259a6",
		Name:            "Mtenda Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{W}, {T}: Return this creature to its owner's hand and return target Griffin card from your graveyard to your hand. Activate only during your upkeep.",
			Cost:      effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Condition: effects.DuringYourUpkeep(),
			Targets:   effects.TargetCardInGraveyard("target Griffin card from your graveyard", effects.OfSubtype("Griffin"), effects.YouOwn()),
			Effect:    effects.S82OldB812ReturnThisAndAGriffin,
		}},
	})
}
