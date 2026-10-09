package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// x_ceiling.go — the printed "X can't be greater than <count>" ceilings
// a spell declares on Spec.XCeiling (#2581; game/x_ceiling.go has the
// rule). Append-only: one var per count, shared by every card that
// prints it.

// XCeilingSnowLandsYouControl is "X can't be greater than the number
// of snow lands you control" (Winter's Chill). Effective types, so a
// land an effect made snow counts; a phased-out land is not on the
// battlefield (CR 702.26b) and does not.
var XCeilingSnowLandsYouControl = &game.XCeiling{
	Label: "the number of snow lands you control",
	Count: func(g *game.Game, caster uuid.UUID) int {
		return g.CountControlledMatchingForEffect(caster, game.PermanentQuery{
			Types:      []string{"land"},
			Supertypes: []string{"snow"},
		})
	},
}

// XCeilingPlayersInGame is "X can't be greater than the number of
// players in the game" (Open the Way). A player who has left the game
// is not in it (CR 800.4a).
var XCeilingPlayersInGame = &game.XCeiling{
	Label: "the number of players in the game",
	Count: func(g *game.Game, _ uuid.UUID) int {
		n := 0
		for _, p := range g.Seats {
			if p != nil && !p.Eliminated {
				n++
			}
		}
		return n
	},
}
