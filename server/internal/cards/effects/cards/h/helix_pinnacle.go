package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helix Pinnacle — Enchantment {G} (Special Guests):
//
//	"Shroud
//	 {X}: Put X tower counters on this enchantment.
//	 At the beginning of your upkeep, if there are 100 or more tower
//	 counters on this enchantment, you win the game."
//
// The win is an intervening "if" (CR 603.4): checked as the upkeep
// begins and again as the trigger resolves, so an enchantment that has
// lost its counters (or left) in response wins nothing.
func init() {
	const counters = "tower"
	effects.Register(effects.Spec{
		OracleID:        "9ddc5a06-b4e1-459a-9e39-7cd1aa994fe1",
		Name:            "Helix Pinnacle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		XMatters:        true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}: Put X tower counters on this enchantment.",
			Cost:    effects.ManaCost("{X}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.AddCounter{Target: item.SourceCardID, Kind: counters, N: ctx.X()}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "there are 100 or more tower counters on this enchantment",
				Test: func(ctx *effects.Context) bool {
					c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
					return ok && effects.OnBattlefield(ctx.Game, ctx.Source()) && c.Counters[counters] >= 100
				},
			}, effects.AtYourUpkeep("Helix Pinnacle — you win the game", effects.Do(effects.WinTheGame{}))),
		},
	})
}
