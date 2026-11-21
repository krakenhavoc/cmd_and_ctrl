package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimble Larcenist — Creature — Bird Rogue {W}{U}{B}, 2/1 (Streets of New
// Capenna, #5491):
//
//	"Flying
//	 When this creature enters, target opponent reveals their hand. You choose
//	 an artifact, instant, or sorcery card from it and exile that card."
//
// The revealed-hand pick (ADR 0116) with the exile variant: the whole table
// sees the hand (CR 701.20a), the pick is mandatory when a matching card
// exists, and exiling is not discarding (CR 701.9a). A hand with no matching
// card is revealed and nothing happens (CR 609.3). The exile is permanent, so
// it is not linked to the Larcenist.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters(
		"Nimble Larcenist — target opponent reveals their hand; you choose an artifact, instant, or sorcery card from it and exile it",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Filter: effects.Or(effects.Artifact(), effects.Instant(), effects.Sorcery()),
				Label:  "artifact, instant, or sorcery card",
				Exile:  true,
			}.Apply(ctx)
		})
	enters.Targets = effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:        "abd839ab-8b08-4bc0-900a-3a161c62b06c",
		Name:            "Nimble Larcenist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
