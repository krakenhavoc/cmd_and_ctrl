package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scholar of the Ages — Creature — Human Wizard {5}{U}{U}, 3/3 (Core Set 2020, #5518):
//
//	"When this creature enters, return up to two target instant and/or
//	 sorcery cards from your graveyard to your hand."
//
// Charred Graverobber's targeted enters trigger with an up-to-two clause:
// each pick still in the graveyard as it resolves goes to your hand
// (CR 608.2b), and choosing none is legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1bb1ad32-9ce5-4f14-a408-e142ff30c679",
		Name:         "Scholar of the Ages",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Scholar of the Ages — return up to two target instant and/or sorcery cards from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return returnLegalGraveyardTargetsToHand(NewContext(g, item))
					}),
				TargetCardInGraveyard("up to two target instant and/or sorcery cards from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())).WithCount(0, 2),
			),
		},
	})
}
