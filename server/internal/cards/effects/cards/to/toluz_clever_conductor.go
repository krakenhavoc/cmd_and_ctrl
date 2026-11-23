package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// toluzExileLabel is the label of Toluz's discard trigger, which
// returnCardsExiledWithToOwnersHands reads the exiled cards back by.
const toluzExileLabel = "Toluz, Clever Conductor — exile the discarded cards from your graveyard"

// Toluz, Clever Conductor — Legendary Creature — Human Rogue {W/U}{U}{U/B},
// 3/1 (Streets of New Capenna, #5491):
//
//	"When Toluz enters, it connives.
//	 Whenever you discard one or more cards, exile them from your graveyard.
//	 When Toluz dies, put the cards exiled with it into their owner's hand."
//
// The enters trigger is A.I.M. Scientists' connive. The discard trigger is
// one per batch of discards (Conspiracy Theorist's AtBatchEnd shape, which
// includes the discard Toluz's own connive makes): as it resolves it exiles
// every card you discarded in that batch that is still in your graveyard as
// the same object (one that left since is gone, CR 400.7). The death
// trigger returns what was exiled with Toluz, read off the event log by the
// label above (Bag of Holding's record), to each card's owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "005c181a-05db-4c15-9893-65103cca338e",
		Name:         "Toluz, Clever Conductor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersItConnives("Toluz, Clever Conductor — it connives"),
			effects.WithAtBatchEnd(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, toluzExileLabel, toluzExileDiscarded), effects.ThatManyEvents()),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source) && effects.KeywordSliceContains(lki.Types, "Creature")
			}, "Toluz, Clever Conductor — put the cards exiled with it into their owner's hand",
				effects.ReturnCardsExiledWithToOwnersHands(toluzExileLabel)),
		},
	})
}

// toluzExileDiscarded exiles the batch's discards still in the controller's
// graveyard.
func toluzExileDiscarded(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	for _, id := range effects.S82SncB23DiscardedInBatch(g, item.Controller, item.Trigger.Event) {
		if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
