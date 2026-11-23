package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orochi Leafcaller — Creature — Snake Shaman {G}, 1/1 (Champions of
// Kamigawa, #5603):
//
//	"{G}: Add one mana of any color."
//
// A mana ability (CR 605.1a) whose whole cost is {G} from the pool, as
// Agent of Stromgald's: no {T}, so summoning sickness does not stop it
// (CR 302.6), and it can be activated as often as there is green mana to
// convert.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "081db161-1ad0-4ee7-9d34-c415984080f0",
		Name:         "Orochi Leafcaller",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Mana: "{G}"}, "{G}: Add one mana of any color."),
		},
	})
}
