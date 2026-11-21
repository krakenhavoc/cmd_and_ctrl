package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Sword of Kahless — Legendary Artifact — Equipment {4} (Star Trek
// Commander, #4974):
//
//	"Equipped creature gets +2/+2 and has "Whenever this creature
//	 attacks, look at the top five cards of your library. You may put a
//	 Klingon or Warrior creature card from among them onto the
//	 battlefield. Put the rest on the bottom of your library in a random
//	 order."
//	 Equip {2}"
//
// The pump is PumpAttached (layer 7c). The trigger is the equipped
// creature's own (ADR 0093, Conformer Shuriken's grant), and its body is
// Ureni of the Unwritten's LookAtTopThenMayPutOntoBattlefield: the five
// cards are looked at, so only the attacking creature's controller sees
// them; the pick enters without being cast; the rest go to the bottom in
// a random order. "Klingon or Warrior creature card" reads the card's
// creature types.
//
// No simplification.
const theSwordOfKahlessGrant = "the-sword-of-kahless/attack-look"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e301803-49ab-4fce-82ae-945a39c458b5",
		Name:         "The Sword of Kahless",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: theSwordOfKahlessGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("The Sword of Kahless — look at the top five cards; you may put a Klingon or Warrior creature card onto the battlefield",
					effects.LookAtTopThenMayPutOntoBattlefield(5, effects.Or(effects.OfCreatureType("Klingon"), effects.OfCreatureType("Warrior")), 1,
						"The Sword of Kahless — you may put a Klingon or Warrior creature card from among them onto the battlefield")),
			},
			Text: "Whenever this creature attacks, look at the top five cards of your library. You may put a Klingon or Warrior creature card from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantAbilitiesToAttached(theSwordOfKahlessGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
