package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Paladin — Creature — Human Knight {2}{W}, 3/3 (Foundations,
// #4465):
//
//	"During your turn, this creature has first strike.
//	 During your turn, creatures you control with +1/+1 counters on them
//	 have first strike."
//
// The first is Bearer of Glory's layer 6 grant on itself while its
// controller is the active player. The second is Badgermole's grant to
// creatures you control with a +1/+1 counter, under the same turn test.
// Both are read at every layer pass, so a creature that gets its first
// counter during your turn has first strike from then on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95efbb81-2418-42ed-ae34-5759efff8253",
		Name:         "Inspiring Paladin",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.B42CreatureYouControlWithAPlusOneCounter(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
		},
	})
}
