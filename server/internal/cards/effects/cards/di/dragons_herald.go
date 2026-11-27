package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon's Herald — Creature — Goblin Shaman {R}, 1/1 (Shards of Alara):
//
//	"{2}{R}, {T}, Sacrifice a black creature, a red creature, and a green
//	 creature: Search your library for a card named Hellkite Overlord, put it
//	 onto the battlefield, then shuffle."
//
// Sphinx's Herald's ability with the Dragon's colours: three different
// creatures, one of each colour, matched one to one, the Herald itself, a red
// creature, allowed to be the red one. The search is for the named card and
// always shuffles (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85a38b39-5806-4df9-95cc-8d608e2b5e47",
		Name:         "Dragon's Herald",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.OldB508HeraldAbility(
			"{2}{R}, {T}, Sacrifice a black creature, a red creature, and a green creature: Search your library for a card named Hellkite Overlord, put it onto the battlefield, then shuffle.",
			"{2}{R}", "Hellkite Overlord",
			effects.OldB508CreatureOfColor("a black creature", "B"),
			effects.OldB508CreatureOfColor("a red creature", "R"),
			effects.OldB508CreatureOfColor("a green creature", "G"))},
	})
}
