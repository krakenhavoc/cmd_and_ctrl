package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agent of Stromgald — Creature — Human Knight {R}, 1/1 (Alliances,
// #5634):
//
//	"{R}: Add {B}."
//
// A mana ability (CR 605.1a) whose whole cost is {R} from the pool: no
// {T}, so summoning sickness does not stop it (CR 302.6) and it can be
// activated as often as there is red mana to convert.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "099fd4b3-92fd-4b5f-bb4f-eefa6f09700c",
		Name:         "Agent of Stromgald",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Mana: "{R}"},
			Produced: "{B}",
			Label:    "{R}: Add {B}.",
		}},
	})
}
