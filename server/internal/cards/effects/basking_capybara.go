package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basking Capybara — Creature — Capybara {1}{G}, 1/3 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Descend 4 — This creature gets +3/+0 as long as there are four or
//	 more permanent cards in your graveyard."
//
// Frilled Cave-Wurm's shape. Descend 4 is an ability word (CR 207.2c).
// The +3/+0 is a layer 7c modification (CR 613.4c) of the Capybara alone,
// on while its controller's graveyard holds four or more permanent cards
// (CR 110.4) and off the moment it holds fewer (SelfPumpWhileDescend).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ce678fa-6495-4857-b178-231c283b4514",
		Name:         "Basking Capybara",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{SelfPumpWhileDescend(4, 3, 0)},
	})
}
