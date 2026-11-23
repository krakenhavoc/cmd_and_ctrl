package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Queen's Agent — Creature — Vampire Scout {5}{B}, 3/3 (Ixalan, #5532):
//
//	"Lifelink
//	 When this creature enters, it explores."
//
// Lifelink is an engine keyword carried as printed data: damage it
// deals also gains you that much life (CR 702.15b).
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
		OracleID:        "59104038-28c0-4ebd-9201-d7ae60d79c1c",
		Name:            "Queen's Agent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Queen's Agent — it explores", effects.ThatCreatureExplores),
		},
	})
}
