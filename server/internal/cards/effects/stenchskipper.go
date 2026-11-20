package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stenchskipper — Creature — Elemental {3}{B}, 6/5 (Morningtide, #5592):
//
//	"Flying
//	 At the beginning of the end step, if you control no Goblins, sacrifice
//	 this creature."
//
// Glimmervoid's end-step upkeep with Goblins for artifacts: every end
// step, anyone's, and an intervening if (CR 603.4) checked as the step
// begins and again on resolution, so a Goblin that arrives in response
// keeps the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0817d902-08c8-4ec3-9a84-2de65bc70e6a",
		Name:            "Stenchskipper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control no Goblins",
				Test: func(ctx *Context) bool { return !controlsAnyCard(ctx.Game, ctx.Controller(), HasSubtype("Goblin")) },
			}, On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b15EndStepBegan(ev) && !controlsAnyCard(g, source.Controller, HasSubtype("Goblin"))
				},
				"Stenchskipper — if you control no Goblins, sacrifice it",
				SacrificeThisIfStillOnBattlefield)),
		},
	})
}
