package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Track Down — Sorcery {1}{G} (Core Set 2021, #5511):
//
//	"Scry 3, then reveal the top card of your library. If it's a creature or
//	 land card, draw a card."
//
// The reveal and the draw are the scry's continuation, so they follow the
// player's arrangement of the top three cards: the card revealed is the one
// the player left on top, and a creature or land card is then the card drawn.
// An empty library reveals nothing and draws nothing; a land that is also a
// creature is still one card (the draw is not doubled).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f28b65b-4b73-4080-9cdc-d236ad99f39c",
		Name:         "Track Down",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3, Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{N: 3, ThenTail: ctx.Tail(trackDownReveal)}.Apply(ctx)
		},
	})
}

// trackDownReveal is Track Down's "then reveal the top card of your library.
// If it's a creature or land card, draw a card". An on-disk key.
var trackDownReveal = effects.CardTail("then/s82-old-b105-track-down-reveal", func(ctx *effects.Context, o game.TailOutcome) error {
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{Player: o.Tail.Player, N: 1, Reason: "Track Down — reveal the top card of your library", Revealed: &revealed}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) != 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(revealed[0])
	if !ok || !(c.IsCreature() || c.IsLand()) {
		return nil
	}
	return effects.DrawCards{Player: o.Tail.Player, N: 1}.Apply(ctx)
})
