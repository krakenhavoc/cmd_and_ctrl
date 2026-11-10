package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Station Monitor — Creature — Lizard Artificer {W}{U}, 2/2 (Edge of
// Eternities, #4461):
//
//	"Whenever you cast your second spell each turn, create a 1/1
//	 colorless Drone artifact creature token with flying and "This token
//	 can block only creatures with flying.""
//
// Brimstone Roundup's second-spell watch (YouCastYourSecondSpellEachTurn):
// the cast tally counts every spell the controller has cast this turn,
// whether it resolved, was countered or is still on the stack, and
// the Monitor itself if it was the first (the rulings). The trigger
// resolves above the spell that caused it. The Drone is a catalog template
// (StdB3DroneToken) because its block restriction is a printed rule the
// token always carries.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94e8791b-e72f-4aed-b761-826779337168",
		Name:         "Station Monitor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Station Monitor — create a 1/1 Drone with flying (second spell this turn)",
				Do(CreateToken{Template: StdB3DroneToken(), N: 1})),
		},
	})
}
