package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tenacious Tomeseeker — Creature — Human Knight {2}{U}, 3/2 (Wilds of
// Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 When this creature enters, if it was bargained, return target instant
//	 or sorcery card from your graveyard to your hand."
//
// Bargain is the shared optional cost (g2Bargain), and the intervening if
// (CR 603.4) is Diminisher Witch's: it reads the permanent's own record of
// what was paid, so an unbargained Tomeseeker, a reanimated one and a
// token copy put nothing on the stack and ask for no target. The return
// is Zealous Lorecaster's; a card that left the graveyard in response is
// dropped by the legality re-check (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "b07bf08b-d9ec-429f-be6d-a7afa27c60c4",
		Name:          "Tenacious Tomeseeker",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, diminisherWitchWasBargained),
					"Tenacious Tomeseeker — return target instant or sorcery card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return returnLegalGraveyardTargetsToHand(NewContext(g, item))
					}),
				TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery()))),
		},
	})
}
