package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emissary of Sunrise — Creature — Human Cleric {2}{W}, 2/1 (Ixalan, #5532):
//
//	"First strike
//	 When this creature enters, it explores."
//
// First strike is an engine keyword carried as printed data: it deals
// combat damage in the first combat damage step (CR 702.7b).
//
// The explorer is the object that entered, read off the trigger
// (thatCreatureExplores, Cenote Scout's body): a land on top goes to
// the hand, anything else gives this creature a +1/+1 counter and asks
// whether to put the card into the graveyard (CR 701.44a). One that has
// left by the time the trigger resolves still explores by its last
// known information: the card is revealed, and no counter is put on
// anything (CR 701.44c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e2155e61-f389-4b1d-9598-9a79cbccaaa2",
		Name:            "Emissary of Sunrise",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Emissary of Sunrise — it explores", effects.ThatCreatureExplores),
		},
	})
}
