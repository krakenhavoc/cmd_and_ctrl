package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootha, Mastering the Moment — Legendary Creature — Orc Sorcerer
// {2}{U}{R}, 3/4 (Secrets of Strixhaven Commander, #4834):
//
//	"At the beginning of combat on your turn, if you've cast an instant
//	 or sorcery spell this turn, create an X/X blue and red Elemental
//	 creature token with flying and haste, where X is the greatest mana
//	 value among instant and sorcery spells you've cast this turn."
//
// An intervening "if" (CR 603.4): the beginning-of-combat trigger only
// goes on the stack when an instant or sorcery has been cast this turn
// (CastTally.InstantSorcery, bumped as each spell became cast), and the
// condition is checked again as it resolves. X is the greatest mana value
// among those spells (CastTally.InstantSorceryMaxManaValue): read off the
// spell on the stack as it was cast, so an X spell counts the X it was cast
// with and a cost increase, reduction or alternative cost does not change
// it (the rulings, CR 202.3). The tally is per player and per turn, so a
// spell cast by an opponent never counts.
//
// The token is an X/X with its power and toughness fixed at creation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "348c67ef-9ccc-4651-9038-efdf1ad1b36a",
		Name:         "Rootha, Mastering the Moment",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you've cast an instant or sorcery spell this turn",
				Test: func(ctx *effects.Context) bool {
					return ctx.Game.CastTallyFor(ctx.Controller()).InstantSorcery > 0
				},
			}, effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepBeginCombat, true), effects.SocB12YouCastAnInstantOrSorceryThisTurn),
				"Rootha, Mastering the Moment — create an X/X Elemental with flying and haste, X the greatest mana value among your instants and sorceries this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := g.CastTallyFor(ctx.Controller()).InstantSorceryMaxManaValue
					return effects.CreateToken{Template: game.Card{
						Name:           "Elemental",
						TypeLine:       "Token Creature — Elemental",
						Power:          x,
						Toughness:      x,
						Colors:         []string{"U", "R"},
						Keywords:       []string{"flying", "haste"},
						PrintedPTKnown: true,
					}, N: 1}.Apply(ctx)
				})),
		},
	})
}
