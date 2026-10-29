package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Troll of Khazad-dûm — Creature — Troll {5}{B}, 6/5 (The Hobbit Eternal,
// HOC):
//
//	"This creature can't be blocked except by three or more creatures.
//	 Swampcycling {1} ({1}, Discard this card: Search your library for a
//	 Swamp card, reveal it, put it into your hand, then shuffle.)"
//
// Rampaging Ceratops' block rule (MinBlockers 3) and Injector Crocodile's
// swampcycling (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "222fd9fc-7af4-4269-bbb9-cd7b8ac029bd",
		Name:         "Troll of Khazad-dûm",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			MinBlockers(OnSelf(), 3),
		},
		Activated: []ActivatedAbility{
			Typecycling("Swampcycling", "{1}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}
