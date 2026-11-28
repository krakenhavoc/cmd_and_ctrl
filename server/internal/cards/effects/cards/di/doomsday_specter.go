package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doomsday Specter — Creature — Specter {2}{U}{B}, 2/3 (Planeshift, S82 OLD
// #5615):
//
//	"Flying
//	 When this creature enters, return a blue or black creature you control to
//	 its owner's hand.
//	 Whenever this creature deals combat damage to a player, look at that
//	 player's hand and choose a card from it. The player discards that card."
//
// The enters trigger is mandatory and a choice, not a target (Whitemane Lion's
// ReturnOneYouControl): any blue or black creature you control as it resolves,
// the Specter itself included, which is the usual pick when it is alone. The
// damage trigger is Leshrac's Sigil's look-and-choose against the player the
// damage event names; the hand is shown only to its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5811effb-5010-44b9-bb97-9d7eb358327f",
		Name:            "Doomsday Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Doomsday Specter — return a blue or black creature you control to its owner's hand", effects.Do(effects.ReturnOneYouControl{
				Match:    effects.OldB640BlueOrBlackCreature,
				Question: "Doomsday Specter — return a blue or black creature you control to its owner's hand",
			})),
			effects.WheneverThisDealsCombatDamageToAPlayer("Doomsday Specter — look at that player's hand and choose a card from it; the player discards that card",
				effects.OldB640LookAtTheDamagedPlayersHand),
		},
	})
}
