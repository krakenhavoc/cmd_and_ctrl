package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Piper — Creature — Elf Shaman {3}{G}, 1/1:
//
//	"{G}, {T}: You may put a creature card from your hand onto the
//	 battlefield."
//
// Quicksilver Amulet's put-from-hand on a creature that taps for it: the
// creature is PUT onto the battlefield, never cast, so counterspells and cast
// triggers see nothing and its enters triggers fire. "You may" is a real
// decline, and a player with no creature card in hand is not prompted. The
// {T} makes summoning sickness apply to the Piper itself (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "872ef617-7cdb-4a7a-85f5-efc9e12bece5",
		Name:         "Elvish Piper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: You may put a creature card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.Creature(),
					Optional: true,
					Label:    "Elvish Piper — you may put a creature card from your hand onto the battlefield",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
