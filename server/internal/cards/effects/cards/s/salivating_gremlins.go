package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salivating Gremlins — Creature — Gremlin {2}{R}, 2/3 (Kaladesh, #5539):
//
//	"Whenever an artifact you control enters, this creature gets +2/+0 and
//	 gains trample until end of turn."
//
// Weldfast Wingsmith's "artifact entered under your control" watcher with
// Hand That Feeds' pump-and-keyword body. The Gremlins is not an artifact, so
// it never triggers itself.
//
// No simplification.
func init() {
	const label = "Salivating Gremlins — this creature gets +2/+0 and gains trample until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "4766f232-8275-4256-b756-8dd5c5e825a3",
		Name:         "Salivating Gremlins",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
			}, label, func(g *game.Game, item *game.StackItem) error {
				return effects.SocB27ThisGetsAndGains(effects.NewContext(g, item), 2, "trample", "Salivating Gremlins")
			}),
		},
	})
}
