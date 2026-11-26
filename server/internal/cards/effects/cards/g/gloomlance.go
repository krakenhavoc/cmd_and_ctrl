package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gloomlance — Sorcery {3}{B}{B} (Shadowmoor, #5591):
//
//	"Destroy target creature. If that creature was green or white, its
//	 controller discards a card."
//
// The creature's colours and controller are read before it is destroyed
// (CR 608.2h), as it was on the battlefield, so a creature painted green by
// an effect counts and a stolen one's discard is its thief's. The discard
// does not depend on the destruction: an indestructible green creature
// still makes its controller discard. The player chooses the card; an
// empty hand discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2aa040d2-35a1-4b1d-9542-395e1b8a9fd8",
		Name:         "Gloomlance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			tail := ctx.Tail(gloomlanceDiscard)
			if c.HasColor("G") || c.HasColor("W") {
				tail = effects.WithCarry(tail, c.Controller)
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, tail)
		},
	})
}

// gloomlanceDiscard is the second sentence, run once the destruction has
// settled: the carried player (the creature's last controller, only when it
// was green or white) discards a card. An on-disk key: never renamed,
// never reused.
var gloomlanceDiscard = effects.CardTail("card/gloomlance-discard", func(ctx *effects.Context, o game.TailOutcome) error {
	player := effects.TailCarry0(o)
	if player == uuid.Nil || effects.CardsInHand(ctx.Game, player) == 0 {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   ctx.Source(),
		N:        1,
		Question: "Discard a card",
	})
	return nil
})
