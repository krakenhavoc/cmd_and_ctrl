package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tidehollow Sculler — Artifact Creature — Zombie {W}{B}, 2/2 (Shards of
// Alara, S82 OLD batch 517, #5588):
//
//	"When this creature enters, target opponent reveals their hand and you
//	 choose a nonland card from it. Exile that card.
//	 When this creature leaves the battlefield, return the exiled card to its
//	 owner's hand."
//
// The two-trigger form, not Brain Maggot's "until" (CR 610.3): the exile is
// linked to the Sculler object that did it (CR 607.2a) and the leave trigger
// returns the card still in exile with that link. A Sculler that leaves before
// its entry trigger resolves finds nothing to return, and the card is then
// exiled for good, as the rulings say; one that comes back is a new object
// with no claim on it (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51b55057-8b14-4277-894b-4813b4f7eb39",
		Name:         "Tidehollow Sculler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Tidehollow Sculler — target opponent reveals their hand; you choose a nonland card from it and exile it",
					effects.NullSummonerExileFromHand),
				effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.ReturnTheExiledCardToHand("Tidehollow Sculler"),
		},
	})
}
