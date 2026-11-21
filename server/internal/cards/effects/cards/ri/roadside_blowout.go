package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roadside Blowout — Sorcery {2}{U} (Aetherdrift, #4464):
//
//	"This spell costs {2} less to cast if it targets a permanent with mana
//	 value 1.
//	 Return target creature or Vehicle an opponent controls to its
//	 owner's hand.
//	 Draw a card."
//
// The reduction is CostsLessIfItTargets over a permanent whose mana
// value is exactly 1 (CR 202.3), judged when the cost is determined
// (CR 601.2f). The bounce is Bounce Off's body limited to an opponent's
// permanent, then a draw. With its only target illegal on resolution the
// spell does nothing at all, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03008c7d-c738-473d-a968-1922b2899d1a",
		Name:         "Roadside Blowout",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets: effects.TargetPermanent("target creature or Vehicle an opponent controls",
			effects.Or(effects.Creature(), effects.HasSubtype("Vehicle")), effects.OpponentControls()),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a permanent with mana value 1.",
				effects.And(effects.ManaValueGE(1), effects.ManaValueLE(1))),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BounceTheTarget(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
