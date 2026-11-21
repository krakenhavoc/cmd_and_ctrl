package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elven Farsight — Sorcery {G} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Scry 3, then you may reveal the top card of your library. If a creature
//	 card is revealed this way, draw a card."
//
// Galadriel of Lothlórien's reveal after a scry: the "then" is the scry's tail,
// so it waits until the cards are put back. The reveal is a "may" asked
// afterwards; on a yes the top card is revealed to the table (an empty library
// reveals nothing), and only if it is a creature card (read from the card, not
// a token) do you draw. The revealed card is the one drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04b35e95-2963-4a0a-bd1c-d516cda2dd37",
		Name:         "Elven Farsight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{N: 3, ThenTail: ctx.ItemTail(elvenFarsightAfterScry)}.Apply(ctx)
		},
	})
}

// elvenFarsightAfterScry asks the "you may reveal". An on-disk key: never
// renamed, never reused.
var elvenFarsightAfterScry = effects.CardTail("then/elven-farsight-after-scry", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.MayChoice{
		Question: "Elven Farsight — reveal the top card of your library?",
		Then:     elvenFarsightReveal,
	}.Apply(ctx)
})

var elvenFarsightReveal = effects.MayYes("confirm/elven-farsight-reveal", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{N: 1, Reason: "Elven Farsight — revealed from the top of the library", Revealed: &revealed}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	if c, ok := ctx.Game.LookupCardForEffect(revealed[0]); ok && c.IsCreature() && !c.IsToken() {
		return effects.DrawCards{N: 1}.Apply(ctx)
	}
	return nil
})
