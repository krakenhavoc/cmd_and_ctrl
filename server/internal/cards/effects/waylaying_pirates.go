package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5453e4fa-49ea-43fc-8de5-6873d29ef627",
		Name:         "Waylaying Pirates",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(InterveningIf(CondBoard{
				Name: "you control an artifact",
				Test: func(ctx *Context) bool { return b03ArtifactsControlled(ctx.Game, ctx.Controller()) > 0 },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && b03ArtifactsControlled(g, source.Controller) > 0
			}, "Waylaying Pirates — tap target artifact or creature an opponent controls and put a stun counter on it", tapTargetAndStunIt)),
				TargetPermanent("target artifact or creature an opponent controls",
					And(Or(Artifact(), Creature()), OpponentControls()))),
		},
	})
}
