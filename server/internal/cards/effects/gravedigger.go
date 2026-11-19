package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravedigger — Creature — Zombie {3}{B}, 2/2 (Foundations Jumpstart):
//
//	"When this creature enters, you may return target creature card from
//	 your graveyard to your hand."
//
// Archaeomancer's return (returnTargetCardToHand) with Amok Time's creature
// target, behind a CR 603.5 "you may". No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a2030cc-d7ee-4059-b2d7-fb95ea8e267b",
		Name:         "Gravedigger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Gravedigger — return target creature card from your graveyard to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())),
				"Gravedigger — return it to your hand?"),
		},
	})
}
