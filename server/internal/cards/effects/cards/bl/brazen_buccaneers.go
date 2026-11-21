package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brazen Buccaneers — Creature — Human Pirate {3}{R}, 2/2 (Ixalan, #5532):
//
//	"Haste
//	 When this creature enters, it explores."
//
// Haste is an engine keyword carried as printed data: it can attack and
// {T} the turn it arrives (CR 702.10b).
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
		OracleID:        "c8afa4b5-01be-400a-836f-aa27e1ae1346",
		Name:            "Brazen Buccaneers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Brazen Buccaneers — it explores", effects.ThatCreatureExplores),
		},
	})
}
