package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ragamuffin Raptor — {4}{G} Creature — Dinosaur 4/3 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"When this creature enters, return up to one target creature or Food
//	 card from your graveyard to your hand."
//
// General Leo Cristophe's "up to one target … card from your graveyard"
// clause, returned to the hand. A card that left the graveyard in response is
// no longer a legal target and nothing moves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71b81f3f-6092-4819-8529-e27bccfcf088",
		Name:         "Ragamuffin Raptor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Ragamuffin Raptor — return up to one target creature or Food card from your graveyard to your hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("up to one target creature or Food card from your graveyard",
					YouOwn(), Or(Creature(), HasSubtype("Food"))).WithCount(0, 1)),
		},
	})
}
