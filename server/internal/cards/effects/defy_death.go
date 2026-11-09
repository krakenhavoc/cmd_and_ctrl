package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Defy Death — Sorcery {3}{W}{W}:
//
//	"Return target creature card from your graveyard to the
//	 battlefield. If it's an Angel, put two +1/+1 counters on it."
//
// The Angel check reads the permanent as it is on the battlefield, once
// it has landed and nothing could happen in between (the ruling): an
// effect that makes it an Angel only there counts, and "Angel" is a word
// of the spell's text (CR 612.2). The counters are put on it, not
// entered with.
func init() {
	Register(Spec{
		OracleID:     "dbf9c3e8-35dd-4b80-bd23-f43e21cfd991",
		Name:         "Defy Death",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := legalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ReturnFromGraveyardToBattlefieldKeyedForEffect(ids[0], ctx.Controller(), false,
				ctx.ItemTail(defyDeathCounters))
		},
	})
}

// defyDeathCounters is "if it's an Angel, put two +1/+1 counters on it"
// as a keyed tail, told the permanent that entered. An on-disk key.
var defyDeathCounters = EnteredTail("defy-death/two-counters-if-angel", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil || !onBattlefield(ctx.Game, entered) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(entered)
	if !ok || !c.HasSubtype(ctx.TextSubtype("Angel")) {
		return nil
	}
	return ctx.Game.AddCounterForEffect(entered, game.CounterPlusOne, 2)
})
