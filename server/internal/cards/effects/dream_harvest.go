package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dream Harvest — Sorcery {5}{U/B}{U/B} (Lorwyn Eclipsed, #4458):
//
//	"Each opponent exiles cards from the top of their library until they
//	 have exiled cards with total mana value 5 or greater this way. Until
//	 end of turn, you may cast cards exiled this way without paying their
//	 mana costs."
//
// Tasha's Hideous Laughter's chain with a 5 threshold: one opponent at a
// time in APNAP order, each run ending when its total reaches 5 (a library
// that runs out is exiled whole and nobody loses for it, CR 701.17b). Once
// every opponent is done the chain's `done` tail is told every card exiled
// and grants the controller the free cast of the nonland ones until end
// of turn, from exile. Cards cast this way follow normal timing (a
// sorcery's controller can cast creatures in the same main phase) and X is
// 0 for a spell with {X} (the ruling).
func init() {
	Register(Spec{
		OracleID:     "70b9497f-0dd6-4b66-9675-57c7041fe3c0",
		Name:         "Dream Harvest",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return eachPlayerInTurn(ctx.Game, ctx.ItemTail(dreamHarvestExileStep), ctx.Opponents(), nil,
				ctx.ItemTail(dreamHarvestDone), dreamHarvestExileAct)
		},
	})
}

var dreamHarvestExileStep = registerEachPlayerStep("dream-harvest/exile-step", dreamHarvestExileAct)

func dreamHarvestExileAct(ctx *Context, opp uuid.UUID, _ int, step game.KeyedTail) error {
	return b27ExileTopUntilTotalManaValueThen(ctx, opp, 5, step)
}

// dreamHarvestDone grants the free cast of every card exiled this way. An
// on-disk key: never renamed, never reused.
var dreamHarvestDone = CardTail("dream-harvest/after-exile", func(ctx *Context, o game.TailOutcome) error {
	var cards []game.Card
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		c, ok := ctx.Game.LookupCardForEffect(id)
		if ok && z != nil && z.Kind == game.ZoneExile && !c.IsLand() {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   ctx.Game.UntilEndOfTurnDuration(),
		Cost:       "{0}",
		AltCostKey: game.AltCostKeyGrantedCost,
		CastOnly:   true,
		Source:     ctx.Source(),
		SourceName: "Dream Harvest",
		Label:      "Dream Harvest — cast it without paying its mana cost",
	}, cards)
	return nil
})
