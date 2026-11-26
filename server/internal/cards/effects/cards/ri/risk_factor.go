package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Risk Factor — Instant {2}{R} (Guilds of Ravnica, #5525):
//
//	"Target opponent may have Risk Factor deal 4 damage to them. If that player
//	 doesn't, you draw three cards.
//	 Jump-start"
//
// Browbeat's offer (AnyPlayerPays), made to the target opponent alone (the
// offer's Only seam) instead of to every player. Taking the damage is the
// opponent's payment: Risk Factor deals it through the ordinary damage path, so
// it can be prevented and taking it still stops the draw. If the opponent
// declines, or has left the game, the caster draws three. A target that has
// become illegal by resolution means the spell does nothing (CR 608.2b).
// Jump-start is the shared alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "7f61a2a6-2fe2-48ec-8e2e-74d147be0a0a",
		Name:             "Risk Factor",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 3},
		Targets:          effects.TargetPlayer("target opponent", effects.Opponent()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Cmdr2S39JumpStart("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			who, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.AnyPlayerPays{
				Pay:          effects.DamagePayment(4),
				Only:         who,
				Question:     "Risk Factor — take 4 damage to stop its controller drawing three cards?",
				AcceptLabel:  "Take 4 damage",
				DeclineLabel: "Don't",
				Then:         riskFactorDraws,
			}.Apply(ctx)
		},
	})
}

// riskFactorDraws is "If that player doesn't, you draw three cards", keyed
// (#4264): the spell's controller is the offer's controller. An on-disk key:
// never renamed, never reused.
var riskFactorDraws = effects.AnyPlayerConsequences("any-player/risk-factor-draw", nil,
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
		return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
	}, nil)
