package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doomskar Warrior — Creature — Human Warrior {2}{G}{G}, 4/3 (March of the
// Machine, #5477):
//
//	"Backup 1
//	 Trample
//	 Whenever this creature deals combat damage to a player or battle, look
//	 at that many cards from the top of your library. You may reveal a
//	 creature or land card from among them and put it into your hand. Put
//	 the rest on the bottom of your library in a random order."
//
// Backup (CR 702.165) grants trample and the combat-damage trigger, a
// bundle in Spec.Grants that the Warrior also has itself. "That many" is
// the combat damage the event carried; "your library" is the controller of
// the creature that has the ability. The look-at, take and random-bottom
// are Horn of the Mark's LookAtTopThenMayTakeToHand.
//
// No simplification.
const doomskarWarriorGrant = "doomskar-warrior/damage-look"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ac00312-8963-4fee-8561-1a54e533adbb",
		Name:            "Doomskar Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Grants: []effects.AbilityGrant{{
			Key:       doomskarWarriorGrant,
			Triggered: []game.TriggeredAbility{doomskarWarriorDamage()},
			Text:      "Whenever this creature deals combat damage to a player or battle, look at that many cards from the top of your library. You may reveal a creature or land card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Doomskar Warrior", 1, doomskarWarriorGrant, "trample"),
			doomskarWarriorDamage(),
		},
	})
}

func doomskarWarriorDamage() game.TriggeredAbility {
	return effects.S82MomB04ThisDealtCombatDamage(effects.ThisDealtCombatDamageToAPlayerOrBattle,
		"Doomskar Warrior — look at that many cards from the top of your library; you may reveal a creature or land card and put it into your hand",
		func(g *game.Game, item *game.StackItem) error {
			return effects.LookAtTopThenMayTakeToHand(item.Params.Amount, effects.Or(effects.Creature(), effects.Land()), 1,
				"Doomskar Warrior — reveal a creature or land card and put it into your hand")(g, item)
		})
}
