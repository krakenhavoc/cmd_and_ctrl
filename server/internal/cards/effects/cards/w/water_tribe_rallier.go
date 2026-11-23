package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Water Tribe Rallier — Creature — Human Soldier Ally {1}{W}, 2/2
// (Avatar: The Last Airbender, #4459):
//
//	"Waterbend {5}: Look at the top four cards of your library. You may
//	 reveal a creature card with power 3 or less from among them and put
//	 it into your hand. Put the rest on the bottom of your library in a
//	 random order. (While paying a waterbend cost, you can tap your
//	 artifacts and creatures to help. Each one pays for {1}.)"
//
// The cost is WaterbendCost("{5}") (CR 701.67a, Katara, Water Tribe's
// Hope's activated waterbend): each untapped artifact or creature you
// control can pay {1} of the five, the Rallier itself included, and it has
// no {T} in its cost, so summoning sickness doesn't stop it. The effect is
// Horn of the Mark's look-and-take (LookAtTopThenMayTakeToHand): a look,
// so only you see the four; the card taken is revealed; the rest go to the
// bottom in a random order. A card's power in the library is its printed
// power.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64acb083-a0a3-46ec-9c93-b2ac36878703",
		Name:         "Water Tribe Rallier",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Waterbend {5}: Look at the top four cards of your library. You may reveal a creature card with power 3 or less " +
				"from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Cost:    effects.WaterbendCost("{5}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: effects.LookAtTopThenMayTakeToHand(4, effects.And(effects.Creature(), effects.PowerLE(3)), 1,
				"Water Tribe Rallier — you may reveal a creature card with power 3 or less and put it into your hand"),
		}},
	})
}
