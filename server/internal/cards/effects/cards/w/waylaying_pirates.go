package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waylaying Pirates — Creature — Human Pirate {3}{U}, 3/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"When this creature enters, if you control an artifact, tap target
//	 artifact or creature an opponent controls and put a stun counter on
//	 it."
//
// Donatello, Turtle Techie's intervening if (CR 603.4): the trigger needs
// an artifact you control as the Pirates enter, and again as it resolves.
// The body is Shipwreck Patrol's tap-and-stun (tapTargetAndStunIt) with
// Vault Guardsman's "artifact or creature an opponent controls" target. A
// stun counter stops the next untap (CR 122.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5453e4fa-49ea-43fc-8de5-6873d29ef627",
		Name:         "Waylaying Pirates",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(effects.CondBoard{
				Name: "you control an artifact",
				Test: func(ctx *effects.Context) bool { return effects.B03ArtifactsControlled(ctx.Game, ctx.Controller()) > 0 },
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.B03ArtifactsControlled(g, source.Controller) > 0
			}, "Waylaying Pirates — tap target artifact or creature an opponent controls and put a stun counter on it", effects.TapTargetAndStunIt)),
				effects.TargetPermanent("target artifact or creature an opponent controls",
					effects.And(effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls()))),
		},
	})
}
