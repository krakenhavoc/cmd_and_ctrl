package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Compelling Deterrence — Instant {1}{U} (Innistrad Remastered):
//
//	"Return target nonland permanent to its owner's hand. Then that player
//	 discards a card if you control a Zombie."
//
// "That player" is the permanent's owner. The Zombie is looked for after the
// bounce, as the sentence is read, and the discard waits for the bounce to
// settle (a commander's command-zone question comes first): it is the bounce's
// registered tail.
func init() {
	Register(Spec{
		OracleID:     "fb0c9048-73ee-434b-a018-dbfbe32e56f5",
		Name:         "Compelling Deterrence",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Not(Land())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BounceToHand{Target: id, ThenTail: ctx.ItemTail(compellingDeterrenceDiscard)}.Apply(ctx)
		},
	})
}

var compellingDeterrenceDiscard = CardTail("compelling-deterrence/discard", func(ctx *Context, _ game.TailOutcome) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		// The target was a legal one when the spell began to resolve; its id is
		// read off the item whether or not it is still legal now.
		if len(ctx.Item.Targets) == 0 {
			return nil
		}
		id = ctx.Item.Targets[0].ID
	}
	if !b05ControlsSubtype(ctx.Game, ctx.Controller(), ctx.TextSubtype("Zombie")) {
		return nil
	}
	card, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   card.Owner,
		Source:   ctx.Source(),
		N:        1,
		Question: "Compelling Deterrence — discard a card",
	})
	return nil
})
