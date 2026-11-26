package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fearsome Awakening — {4}{B} Sorcery (Fate Reforged):
//
//	"Return target creature card from your graveyard to the battlefield. If
//	 it's a Dragon, put two +1/+1 counters on it."
//
// The type is read on the permanent after it has entered (CR 205.3), so a
// changeling counts; a creature that did not enter, or has already left,
// gets no counters.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6e87f78-d4dd-44b2-bf70-92436ce583c9",
		Name:         "Fearsome Awakening",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			card, ok := effects.ReanimateSingleTarget(ctx, ctx.Controller())
			if !ok {
				return nil
			}
			if z := ctx.Game.FindCardZoneForEffect(card.InstanceID); z == nil || z.Kind != game.ZoneBattlefield {
				return nil
			}
			now, ok := ctx.Game.LookupCardForEffect(card.InstanceID)
			if !ok || !now.HasSubtype(ctx.TextSubtype("Dragon")) {
				return nil
			}
			return effects.AddCounter{Target: card.InstanceID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
		},
	})
}
