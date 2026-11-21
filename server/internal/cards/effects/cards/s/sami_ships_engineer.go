package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sami, Ship's Engineer — Legendary Creature — Human Artificer
// {2}{R}{W}, 2/4 (Edge of Eternities, #4461):
//
//	"At the beginning of your end step, if you control two or more
//	 tapped creatures, create a tapped 2/2 colorless Robot artifact
//	 creature token."
//
// Dawnstrike Vanguard's end-step intervening "if" (CR 603.4): checked as
// the end step begins and again as the ability resolves, so untapping a
// creature in response makes nothing. Sami counts toward the two when
// tapped. The Robot is Secluded Starforge's 2/2 row, made with a
// TokenSpec, so it enters tapped rather than being tapped afterwards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "470c16c5-b078-47f5-89b0-42991e1f7157",
		Name:         "Sami, Ship's Engineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control two or more tapped creatures",
				Test: func(ctx *effects.Context) bool {
					return effects.B24TappedCreaturesControlled(ctx.Game, ctx.Controller()) >= 2
				},
			}, effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B24TappedCreaturesControlled(g, source.Controller) >= 2
			}), "Sami, Ship's Engineer — create a tapped 2/2 colorless Robot artifact creature token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("2/2 colorless Robot artifact")).EntersTapped(), N: 1}))),
		},
	})
}
