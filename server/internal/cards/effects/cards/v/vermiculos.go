package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vermiculos — Creature — Horror {4}{B}, 1/1 (Mirrodin, S82 OLD, #5606):
//
//	"Whenever an artifact enters, this creature gets +4/+4 until end of
//	 turn."
//
// Battered Golem's trigger: every artifact entering the battlefield under
// any player's control, an artifact creature or a token included, and one
// trigger for each. The Vermiculos is not itself an artifact. Each bonus
// ends in the cleanup step (CR 514.2) and they stack.
//
// No simplification.
func init() {
	const label = "Vermiculos — +4/+4 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "88cadf97-4b24-481a-9136-96c017fd96e9",
		Name:         "Vermiculos",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B20ArtifactEntered(ev, g)
			}, label, effects.ThisGetsUntilEndOfTurn(4, 4, label)),
		},
	})
}
