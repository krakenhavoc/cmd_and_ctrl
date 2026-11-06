package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Champion of the Clachan — Creature — Kithkin Knight {3}{W}, 4/5
// (Lorwyn Eclipsed):
//
//	"Flash
//	 As an additional cost to cast this spell, behold a Kithkin and exile
//	 it. (Exile a Kithkin you control or a Kithkin card from your hand.)
//	 Other Kithkin you control get +1/+1.
//	 When this creature leaves the battlefield, return the exiled card to
//	 its owner's hand."
//
// The additional cost is BeholdAndExileCost (#4687, CR 701.4a): the caster
// names one Kithkin, a permanent they control or a card in their hand (a
// changeling counts), on reveal_ids, and it is exiled with the spell on
// the stack. With no Kithkin to behold the Champion can't be cast. The
// exiled card is linked to the permanent the spell becomes (CR 607.2q),
// so the leave trigger returns that card and no other: a countered
// Champion leaves its card in exile, a second Champion's card stays with
// the second Champion, and a card that has left exile is not pulled back.
// A token beheld this way ceases to exist in exile and returns nothing.
//
// The anthem is Attuma's: a Layer 7c static over other Kithkin creatures
// the controller controls, read as they are now, so a changeling counts.
//
// No simplification.
func init() {
	yourOtherKithkin := TribeFilter{Tribes: []string{"Kithkin"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "4a220e0c-e463-4264-9a96-658cf842b82e",
		Name:            "Champion of the Clachan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		AdditionalCost:  BeholdAndExileCost("a", "Kithkin"),
		Static: []game.StaticAbility{
			TribalAnthem(yourOtherKithkin, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			ReturnTheExiledCardToHand("Champion of the Clachan"),
		},
	})
}
