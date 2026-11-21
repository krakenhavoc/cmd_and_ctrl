package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodhusk Ritualist — Creature — Vampire Shaman {2}{B}, 2/2 (Worldwake, #5582):
//
//	"Multikicker {B} (You may pay an additional {B} any number of times as
//	 you cast this spell.)
//	 When this creature enters, target opponent discards a card for each
//	 time it was kicked."
//
// The count is Stronghold Arena's: read off the record the resolution
// carried onto the permanent (CR 400.7d) as the trigger is put on the
// stack, and stamped on the item (wwkWhenEntersPerKick), so it survives
// the creature leaving in response. Unkicked, the trigger still goes on
// the stack and does nothing, as printed (there is no intervening if).
// The opponent chooses what to discard (game.QueueDiscardChoiceForEffect),
// and a hand smaller than the count gives up what it has (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "93e1d3a6-e12f-4a0a-be0f-fbf568d89f1a",
		Name:          "Bloodhusk Ritualist",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Multikicker("{B}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.WwkWhenEntersPerKick("Bloodhusk Ritualist — target opponent discards a card for each time it was kicked", 1,
				effects.TargetPlayer("target opponent", effects.Opponent()), bloodhuskRitualistDiscard),
		},
	})
}

// bloodhuskRitualistDiscard has the still-legal opponent discard the
// stamped number of cards of their choice.
func bloodhuskRitualistDiscard(g *game.Game, item *game.StackItem) error {
	p, ok := effects.FirstLegalPlayerTarget(effects.NewContext(g, item))
	if !ok || item.Params.Amount <= 0 {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p, Source: item.SourceCardID, N: item.Params.Amount})
	return nil
}
