package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cylian Sunsinger — Creature — Elf Shaman {1}{G}, 2/2 (Conflux):
//
//	"{R}{G}{W}: This creature and each other creature with the same name as
//	 it get +3/+3 until end of turn."
//
// The set is the creatures on the battlefield with the Sunsinger's name as
// the ability resolves, whoever controls them, locked now (CR 611.2c); a
// card has the name when any face has it (CR 201.2b). The name is the
// Sunsinger's as it last existed if it has left, so the others are still
// pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcc82c14-d632-4333-be35-4a206a10b184",
		Name:         "Cylian Sunsinger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{G}{W}: This creature and each other creature with the same name as it get +3/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}{G}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok || info.Characteristic.Name == "" {
					return nil
				}
				name := info.Characteristic.Name
				return effects.BoostUntilEOT{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsCreature() && game.CardNameMatches(c, name)
					},
					Power: 3, Toughness: 3, Label: "Cylian Sunsinger — +3/+3",
				}.Apply(ctx)
			},
		}},
	})
}
