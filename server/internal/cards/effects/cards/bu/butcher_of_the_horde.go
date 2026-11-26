package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Butcher of the Horde — Khans of Tarkir (S82 OLD b348, #5554):
//
//	Creature — Demon {1}{R}{W}{B}, 5/4: Flying. Sacrifice another creature: This
//	creature gains your choice of vigilance, lifelink, or haste until end of turn.
//
// The choice is made as the ability resolves (a registered option pick).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ff93e62-6ae0-4438-a88e-ae6fd7d91e41",
		Name:            "Butcher of the Horde",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: This creature gains your choice of vigilance, lifelink, or haste until end of turn.",
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PickOption{
					Question: "Butcher of the Horde — gains which until end of turn?",
					Options:  []game.ChoiceOption{{Label: "Vigilance"}, {Label: "Lifelink"}, {Label: "Haste"}},
					ThenKey:  butcherOfTheHordeKeyword,
					Carry:    []uuid.UUID{item.SourceCardID},
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

var butcherOfTheHordeKeywords = []string{"vigilance", "lifelink", "haste"}

var butcherOfTheHordeKeyword = effects.OptionPickThen("option-pick/s82-old-b348-butcher-of-the-horde-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(butcherOfTheHordeKeywords) || len(r.Carry) == 0 {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{butcherOfTheHordeKeywords[r.Index]}, Label: "Butcher of the Horde — keyword until end of turn"}.Apply(ctx)
})
