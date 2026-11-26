package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Signal the Clans — Instant {R}{G} (Gatecrash, #5565):
//
//	"Search your library for three creature cards and reveal them. If you
//	 reveal three cards with different names, choose one of them at random
//	 and put that card into your hand. Shuffle the rest into your library."
//
// The found cards stay in the library, revealed (HoldFound, Verdant
// Mastery's search). A search for a stated quality may find fewer than three
// (CR 701.23b); the random pick is made only when exactly three were
// revealed and no two share a name, through the keyed random stream
// (ChooseAtRandomForEffect). Whatever is not put into the hand is already in
// the library, so "shuffle the rest into your library" is the one shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd68ae3c-52cd-4b20-b015-bd31378d972e",
		Name:         "Signal the Clans",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:    ctx.Controller(),
				Source:    ctx.Source(),
				Pred:      func(c game.Card) bool { return c.IsCreature() },
				Dest:      game.ZoneLibrary,
				HoldFound: true,
				Limit:     3,
				Reveal:    true,
				Reason:    "Signal the Clans — search your library for three creature cards",
				ThenTail:  ctx.ItemTail(signalTheClansFound),
			})
		},
	})
}

// signalTheClansFound is the rest of the spell once the three are revealed.
// An on-disk key: never renamed, never reused.
var signalTheClansFound = effects.CardTail("signal-the-clans/found", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 3 {
		names := map[string]bool{}
		for _, id := range o.IDs {
			if c, ok := ctx.Game.LookupCardForEffect(id); ok {
				names[c.Name] = true
			}
		}
		if len(names) == 3 {
			pick := ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), o.IDs, 1)
			if len(pick) == 1 {
				return ctx.Game.TakeFromLibraryToHandKeyedForEffect(ctx.Controller(), []uuid.UUID{pick[0]}, ctx.Tail(effects.ThenShuffle))
			}
		}
	}
	return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
})
