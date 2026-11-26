package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blast of Genius — Sorcery {4}{U}{R} (Dragon's Maze, S82 OLD b406, #5564):
//
//	"Choose any target. Draw three cards, then discard a card. Blast of Genius
//	 deals damage equal to the discarded card's mana value to that permanent or
//	 player."
//
// Hypothesizzle's draw-then-discard: the draw hands the discard to a registered
// body (a draw can pause on a prompt), the discard prompt opens only after every
// draw has finished, so a drawn card may be discarded, and the damage is the
// discard's continuation. The spell's target travels with the prompt as the
// asking item's targets (the resolving spell is read as the draw finishes), and
// is re-checked as the damage is dealt (CR 608.2b). The damage is the discarded
// card's mana value as it was in the hand (a land is 0, X is 0, CR 202.3e);
// nothing is dealt when no card was discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfe82e0f-7ac0-4de3-b1b3-a2d519b00bbd",
		Name:         "Blast of Genius",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
				Ref: blastOfGeniusDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
			})
		},
	})
}

var blastOfGeniusDiscard = game.RegisterDrawThen("s82-old-b406/blast-of-genius-discard", func(g *game.Game, d game.DrawThen) error {
	var asked *game.AskedItem
	if _, item, ok := g.ResolvingSpellForEffect(d.Source); ok {
		asked = game.NewAskedItem(&item)
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   d.Player,
		Source:   d.Source,
		N:        1,
		Question: "Blast of Genius — discard a card",
		Asked:    asked,
	}, blastOfGeniusDiscarded, nil)
})

// blastOfGeniusDiscarded deals the damage once the discard is done. An
// on-disk key: never renamed, never reused.
var blastOfGeniusDiscarded = game.RegisterDiscardThen("choose-cards/s82-old-b406-blast-of-genius-damage", func(g *game.Game, r game.DiscardsDone) error {
	discarded := r.Discarded.Cards()
	if r.Item == nil || len(discarded) == 0 || len(r.Item.Targets) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(discarded[0])
	if !ok {
		return nil
	}
	ctx := effects.NewContext(g, r.Item)
	target := r.Item.Targets[0]
	if !ctx.IsTargetLegal(target) {
		return nil
	}
	n := card.ManaValue()
	if n <= 0 {
		return nil
	}
	return effects.DealDamage{Source: r.Source, Target: target.ID, Amount: n}.Apply(ctx)
})
