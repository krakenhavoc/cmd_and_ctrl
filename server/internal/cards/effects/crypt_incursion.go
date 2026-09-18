package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crypt Incursion — Instant {2}{B} (EDHREC rank 4396):
//
//	"Exile all creature cards from target player's graveyard. You
//	 gain 3 life for each card exiled this way."
//
// Graveyard hate that gains life instead of drawing a card, and at
// three mana the life is not a rider — against a reanimator deck in
// the mid-game it is routinely twenty or more, which in Commander is
// the difference between surviving the turn and not. It is in roadmap
// batch 42 (#449) under "no new machinery".
//
// The two halves are one batch, not a sweep followed by a count.
// Exile opens a CR 614 replacement window per card, and a COMMANDER
// in that graveyard can take CR 903.9's offer and go to the command
// zone instead — it left the graveyard, but it was not "exiled this
// way" (CR 400.7), so it must not be paid for. The prompt also pauses
// the resolution, so the number is not knowable on the line after the
// sweep. ExileCardsThenForEffect sequences the pause and hands back
// exactly the cards that reached exile, which is the set the second
// sentence counts.
//
// "All creature cards", not "all cards": lands, instants and
// sorceries in that graveyard are untouched, which is why this is a
// tool against reanimator and not against a storm deck.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce00e0d0-c6f4-4fe8-97c3-7279b06d8fdb",
		Name:         "Crypt Incursion",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			ids := b42CardsInGraveyardOf(ctx.Game, item.Targets[0].ID, func(c game.Card) bool {
				return c.IsCreature()
			})
			if len(ids) == 0 {
				return nil
			}
			you := ctx.Controller()
			source := ctx.Source()
			return ctx.Game.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
				if len(exiled) == 0 {
					return nil
				}
				return g.ChangePlayerLifeForEffect(source, you, 3*len(exiled))
			})
		},
	})
}
