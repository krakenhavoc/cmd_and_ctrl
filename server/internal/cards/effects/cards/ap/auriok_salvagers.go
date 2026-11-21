package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auriok Salvagers — Creature — Human Soldier {3}{W}, 2/4 (Fifth Dawn,
// #5604):
//
//	"{1}{W}: Return target artifact card with mana value 1 or less from
//	 your graveyard to your hand."
//
// Argivian Archaeologist's return with no tap and a mana value bound.
// A card with {X} in its cost has X = 0 in the graveyard, so only its
// other symbols count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db2f3a06-f75e-4689-b147-9820aebb30e2",
		Name:         "Auriok Salvagers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{W}: Return target artifact card with mana value 1 or less from your graveyard to your hand.",
			Cost:  effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCardInGraveyard("target artifact card with mana value 1 or less in your graveyard",
				effects.Artifact(), effects.ManaValueLE(1), effects.YouOwn()),
			Effect: effects.ReturnTargetedCardToHand,
		}},
	})
}
