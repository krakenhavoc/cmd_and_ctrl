package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kindred Boon — Enchantment {2}{W}{W}:
//
//	"As this enchantment enters, choose a creature type.
//	 {1}{W}: Put a divinity counter on target creature you control of the
//	 chosen type.
//	 Each creature you control with a divinity counter on it has indestructible."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f04441b0-755c-4e7b-a9e4-4e03b8175ab6",
		Name:         "Kindred Boon",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Kindred Boon"),
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && target.Counters["divinity"] > 0
			}, "indestructible"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Put a divinity counter on target creature you control of the chosen type.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.OfTheChosenType(effects.TargetCreature("target creature you control of the chosen type", effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: id, Kind: "divinity", N: 1}.Apply(ctx)
			},
		}},
	})
}
