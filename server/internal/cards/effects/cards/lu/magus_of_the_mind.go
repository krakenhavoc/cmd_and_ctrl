package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magus of the Mind — Creature — Human Wizard {4}{U}{U}, 4/5 (Commander
// 2017, #5533):
//
//	"{U}, {T}, Sacrifice this creature: Shuffle your library, then exile the
//	 top X cards, where X is one plus the number of spells cast this turn.
//	 Until end of turn, you may play lands and cast spells from among cards
//	 exiled this way without paying their mana costs."
//
// Mind's Desire on a creature, with storm's count folded into one exile:
// shuffle, then exile the top 1 + N cards, where N is every spell cast this
// turn by any player as the ability resolves (the turn tally's casts; a
// copy was not cast, CR 707.10), and grant the controller a free play of
// each until end of turn: a land under the usual land rules, a spell cast
// for {0} (CR 118.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b39d98c3-0173-4d7c-83d0-f80db82fc9ba",
		Name:         "Magus of the Mind",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}, Sacrifice this creature: Shuffle your library, then exile the top X cards, where X is one plus the number of spells cast this turn. Until end of turn, you may play lands and cast spells from among cards exiled this way without paying their mana costs.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{U}", Tap: true, SacrificeSelf: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				me := ctx.Controller()
				if err := (effects.ShuffleLibrary{Player: me}).Apply(ctx); err != nil {
					return err
				}
				_, err := g.ExileTopWithPermissionForEffect(me, me, 1+len(g.TurnTally.Casts), game.CastPermission{
					Cost:     "{0}",
					Duration: g.UntilEndOfTurnDuration(),
				})
				return err
			},
		}},
	})
}
