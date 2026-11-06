package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frilled Cave-Wurm — Creature — Salamander Wurm {3}{U}, 2/5 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Descend 4 — This creature gets +2/+0 as long as there are four or
//	 more permanent cards in your graveyard."
//
// Descend 4 is an ability word (CR 207.2c). The +2/+0 is a layer 7c
// modification (CR 613.4c) of the Wurm alone, on while its controller's
// graveyard holds four or more permanent cards (CR 110.4) and off the
// moment it holds fewer (SelfPumpWhileDescend).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac5294c6-bc79-4a0f-b8b3-52b554561e38",
		Name:         "Frilled Cave-Wurm",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{SelfPumpWhileDescend(4, 2, 0)},
	})
}
