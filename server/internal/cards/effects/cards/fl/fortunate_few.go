package cardsfl

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fortunate Few — Sorcery {3}{W}{W} (Commander 2017, #5533):
//
//	"Choose a nonland permanent you don't control, then each other player
//	 chooses a nonland permanent they don't control that hasn't been chosen
//	 this way. Destroy all other nonland permanents."
//
// Sadistic Shell Game's chain, one prompt per player in turn order: the
// caster chooses first, then each other player in turn order from the
// caster (CR 101.4), each from the nonland permanents that player doesn't
// control minus the ones already chosen. A player with nothing legal to
// choose is asked nothing. The destroy is one simultaneous event over every
// other nonland permanent still on the battlefield, so a permanent that
// would be destroyed regenerates or survives as usual (CR 701.8, 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d732d9b8-dea0-41c4-8cff-67b2b3d36ecb",
		Name:         "Fortunate Few",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			order := append([]uuid.UUID{ctx.Controller()}, effects.OpponentsInTurnOrder(ctx.Game, ctx.Controller())...)
			return fortunateFewAsk(ctx, order, nil)
		},
	})
}

func fortunateFewAsk(ctx *effects.Context, players, chosen []uuid.UUID) error {
	if len(players) == 0 {
		var doomed []uuid.UUID
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if !c.IsLand() && !slices.Contains(chosen, c.InstanceID) {
				doomed = append(doomed, c.InstanceID)
			}
		}
		return ctx.Game.DestroyPermanentsKeyedForEffect(doomed, game.KeyedTail{})
	}
	chooser, rest := players[0], players[1:]
	var candidates []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if !c.IsLand() && c.Controller != chooser && !slices.Contains(chosen, c.InstanceID) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return fortunateFewAsk(ctx, rest, chosen)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  chooser,
		Source:   ctx.Source(),
		Question: "Fortunate Few — choose a nonland permanent you don't control that hasn't been chosen (it survives)",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  fortunateFewPick,
		LiveZone: game.ZoneBattlefield,
		Carry:    effects.PackIDLists(rest, chosen),
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

var fortunateFewPick game.ChooseCardsThen

func init() {
	fortunateFewPick = effects.AskedPick("choose-cards/fortunate-few-pick", func(ctx *effects.Context, r game.CardsChosen) error {
		rest, chosen := effects.UnpackIDPair(r.Carry)
		chosen = append([]uuid.UUID(nil), chosen...)
		for _, id := range r.Picked {
			if !slices.Contains(chosen, id) {
				chosen = append(chosen, id)
			}
		}
		return fortunateFewAsk(ctx, rest, chosen)
	})
}
