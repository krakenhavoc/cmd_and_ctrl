package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Edge Rover — Artifact Creature — Robot Scout {G}, 2/2 (Edge of
// Eternities, #4461):
//
//	"Reach
//	 When this creature dies, each player creates a Lander token. (It's
//	 an artifact with "{2}, {T}, Sacrifice this token: Search your library
//	 for a basic land card, put it onto the battlefield tapped, then
//	 shuffle.")"
//
// Beamsaw Prospector's dies trigger with Party Dude's "each player
// creates": in turn order from the active player, every player still in
// the game creates one Lander (b16LanderToken), each under their own
// control, as one group of creations. Reach is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b907d375-b715-446b-8966-79959feb9f58",
		Name:            "Edge Rover",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Edge Rover — each player creates a Lander token", edgeRoverEachPlayerCreatesALander),
		},
	})
}

// edgeRoverEachPlayerCreatesALander gives every player a Lander, in turn
// order from the active player.
func edgeRoverEachPlayerCreatesALander(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, p := range effects.SeatsFromActive(g) {
		if err := (effects.CreateToken{Controller: p, Template: effects.B16LanderToken(), N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
