package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Echo of Dusk — Creature — Vampire Spirit {1}{B}, 2/2 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Descend 4 — As long as there are four or more permanent cards in
//	 your graveyard, this creature gets +1/+1 and has lifelink."
//
// Descend 4 is an ability word (CR 207.2c). The one condition gates two
// statics on the Echo alone: a layer 7c +1/+1 (CR 613.4c) and a layer 6
// lifelink grant (CR 613.1f), both on while its controller's graveyard
// holds four or more permanent cards (CR 110.4) and both off the moment
// it holds fewer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd0980ae-eabf-45da-98e7-23d81bfa185a",
		Name:         "Echo of Dusk",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			SelfPumpWhileDescend(4, 1, 1),
			SelfKeywordWhileDescend(4, "lifelink"),
		},
	})
}
