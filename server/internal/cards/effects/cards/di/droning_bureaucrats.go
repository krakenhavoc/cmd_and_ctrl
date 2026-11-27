package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Droning Bureaucrats — Creature — Human Advisor {3}{W}, 1/4 (Guildpact):
//
//	"{X}, {T}: Each creature with mana value X can't attack or block this
//	 turn."
//
// A restriction is not a characteristic, so CR 611.2c does not lock the set:
// the scope is read live until cleanup against a mana-value query, and a
// creature that enters later with that mana value is held too. X is the
// {X} paid. A token's mana value is 0, so X = 0 holds every token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21dc394e-ba82-4f27-8014-9236690d10a4",
		Name:         "Droning Bureaucrats",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Each creature with mana value X can't attack or block this turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.RestrictUntilEOT{
					Scope:        game.ScopeCreaturesMatching,
					Restrictions: game.CantAttackOrBlock,
					Queries:      []game.PermanentQuery{(game.PermanentQuery{Types: []string{"creature"}}).WithManaValue(ctx.X())},
					Label:        "Droning Bureaucrats — can't attack or block",
				}.Apply(ctx)
			},
		}},
	})
}
