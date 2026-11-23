package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark-Dweller Oracle — Creature — Goblin Shaman {1}{R}, 2/2:
//
//	"{1}, Sacrifice a creature: Exile the top card of your library. You
//	 may play that card this turn."
//
// The sacrifice is a cost (any creature, itself included); the effect is
// the impulse exile Breeches' bullet shares (exileTopCardYouMayPlayThisTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17917b7a-2b2b-4a96-a21e-10cd91375660",
		Name:         "Dark-Dweller Oracle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Exile the top card of your library. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Impulse: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTopCardYouMayPlayThisTurn(item, effects.NewContext(g, item), 0)
			},
		}},
	})
}
