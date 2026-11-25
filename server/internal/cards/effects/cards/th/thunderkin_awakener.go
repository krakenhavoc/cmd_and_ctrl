package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderkin Awakener — Creature — Elemental Shaman {1}{R}, 1/2 (Core Set 2020,
// #5518):
//
//	"Haste
//	 Whenever this creature attacks, choose target Elemental creature card in
//	 your graveyard with toughness less than this creature's toughness. Return
//	 that card to the battlefield tapped and attacking. Sacrifice it at the
//	 beginning of the next end step."
//
// The target is chosen as the trigger goes on the stack, its toughness bound
// read live off the Awakener (its last-known information once it has left,
// CR 608.2h) both when it is chosen and when it resolves (CR 608.2b). The card
// returns tapped and attacking, the controller choosing what it attacks
// (CR 508.4); it was never declared as an attacker, so it triggers nothing.
// A delayed trigger sacrifices exactly that object at the next end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c0171343-9c0f-4001-8f4e-a5af9256c4d2",
		Name:            "Thunderkin Awakener",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks(
					"Thunderkin Awakener — return target Elemental creature card with lesser toughness from your graveyard tapped and attacking; sacrifice it at the next end step",
					effects.OldB156AwakenerReturn),
				effects.RelativeToSource(effects.TargetCardInGraveyard("target Elemental creature card in your graveyard with toughness less than this creature's toughness",
					effects.Creature(), effects.OfSubtype("Elemental"), effects.YouOwn()), effects.LesserToughness())),
		},
	})
}
