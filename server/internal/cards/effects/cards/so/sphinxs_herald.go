package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx's Herald — Artifact Creature — Vedalken Wizard {U}, 1/1 (Shards of
// Alara):
//
//	"{2}{U}, {T}, Sacrifice a white creature, a blue creature, and a black
//	 creature: Search your library for a card named Sphinx Sovereign, put it
//	 onto the battlefield, then shuffle."
//
// The cost is three different creatures, one of each colour, matched one to
// one (SacrificeEach with coloured kinds, #2526): a creature of two of the
// colours fills only one of them (CR 601.2h), and the Herald itself, a blue
// creature, may be the blue one. The search is for the named card and always
// shuffles, even when nothing is found (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6452873b-1bb2-44b3-b630-88f15962b6e8",
		Name:         "Sphinx's Herald",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.OldB508HeraldAbility(
			"{2}{U}, {T}, Sacrifice a white creature, a blue creature, and a black creature: Search your library for a card named Sphinx Sovereign, put it onto the battlefield, then shuffle.",
			"{2}{U}", "Sphinx Sovereign",
			effects.OldB508CreatureOfColor("a white creature", "W"),
			effects.OldB508CreatureOfColor("a blue creature", "U"),
			effects.OldB508CreatureOfColor("a black creature", "B"))},
	})
}
