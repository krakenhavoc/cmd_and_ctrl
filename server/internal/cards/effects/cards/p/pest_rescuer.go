package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pest Rescuer — Creature — Dryad Druid {2}{G}, 2/2 (Secrets of Strixhaven
// Commander, #4834):
//
//	"At the beginning of each upkeep, if you don't control a Pest creature
//	 token, create a 1/1 black and green Pest creature token with "When this
//	 token dies, you gain 1 life."
//	 If you would gain life, you gain that much life plus 1 instead."
//
// The trigger is on every player's upkeep and carries an intervening "if"
// (CR 603.4): the Pest check is made as the upkeep begins and again as the
// trigger resolves, so a Pest made in response (or lost) is honoured. The
// replacement is the Rhox Faithmender family (LifeGainBecomes) with a "plus 1"
// in place of the doubling: it applies to a gain of at least 1, so a gain of 0
// is not replaced, and it raises the Pests' own "you gain 1 life" to 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02be0b99-d40b-4287-989b-12eb2701115c",
		Name:         "Pest Rescuer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.LifeGainBecomes{
				Amount: func(n int) int { return n + 1 },
				Scope:  effects.LifeGainByController,
				Label:  "Pest Rescuer: gain that much life plus 1",
			}.Build(),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you don't control a Pest creature token",
				Test: func(ctx *effects.Context) bool {
					return !effects.SocB10PestTokenYouControl(ctx.Game, ctx.Controller(), ctx.TextSubtype("Pest"))
				},
			}, effects.On(game.EventBeginUpkeep,
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return !effects.SocB10PestTokenYouControl(g, source.Controller, source.TextSubtype("Pest"))
				},
				"Pest Rescuer — create a Pest token",
				effects.Do(effects.CreateToken{Template: effects.PestToken(), N: 1}))),
		},
	})
}
