package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glimpse the Impossible — Sorcery {2}{R} (Modern Horizons 3):
//
//	"Exile the top three cards of your library. You may play those cards
//	 this turn. At the beginning of the next end step, if any of those
//	 cards remain exiled, put them into your graveyard, then create a 0/1
//	 colorless Eldrazi Spawn creature token for each card put into your
//	 graveyard this way. Those tokens have "Sacrifice this token: Add {C}.""
//
// Reckless Impulse's impulse exile on a this-turn window (ExileTopWithPermission's
// default), then a delayed trigger over the exiled cards (ScheduleDelayedTrigger).
// The trigger's "if" is read as it resolves: a card that was played has left
// exile (a spell is on the stack, a land on the battlefield) and is not
// binned; every card still exiled goes to its owner's graveyard in one move,
// and the tokens are made from the cards that actually reached the graveyard
// (a keyed tail over that move), one Spawn each. With nothing left the
// trigger does nothing. The trigger is scheduled even when a card was played
// at once, since "if any remain" is only knowable at the end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b70696e-b09e-4bc9-8438-d31075d78c0c",
		Name:         "Glimpse the Impossible",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Tokens: 3, CreatureTokens: 3},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			ids, err := ctx.Game.ExileTopWithPermissionForEffect(me, me, 3, game.CastPermission{})
			if err != nil || len(ids) == 0 {
				return err
			}
			return ScheduleDelayedTrigger{
				Label: "Glimpse the Impossible — put the cards still exiled into your graveyard, then create an Eldrazi Spawn for each",
				Cards: ids,
				Body:  glimpseTheImpossibleBody,
			}.Apply(ctx)
		},
	})
}

// glimpseTheImpossibleBody is the delayed half: the listed cards that are
// still in exile go to the graveyard together. On-disk keys: never renamed,
// never reused.
var glimpseTheImpossibleBody = game.SimpleDelayedBody("exile/glimpse-the-impossible-bin", func(g *game.Game, item *game.StackItem) error {
	var left []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneExile {
			left = append(left, t.ID)
		}
	}
	if len(left) == 0 {
		return nil
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(left, NewContext(g, item).ItemTail(glimpseTheImpossibleSpawns))
})

// glimpseTheImpossibleSpawns makes one Eldrazi Spawn per card that reached
// the graveyard.
var glimpseTheImpossibleSpawns = CardTail("glimpse-the-impossible/spawn-per-card", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CreateToken{Template: EldraziSpawnToken(), N: len(o.IDs)}.Apply(ctx)
})
