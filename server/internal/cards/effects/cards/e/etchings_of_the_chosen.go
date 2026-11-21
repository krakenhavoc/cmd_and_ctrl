package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etchings of the Chosen — Enchantment {1}{W}{B}:
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+1.
//	 {1}, Sacrifice a creature of the chosen type: Target creature you
//	 control gains indestructible until end of turn."
func init() {
	sacrifice := effects.SacrificeACreature()
	effects.OfTheChosenType(sacrifice.SacrificeOther)
	effects.Register(effects.Spec{
		OracleID:     "0114f97f-b196-4af1-90aa-eba164cb8f22",
		Name:         "Etchings of the Chosen",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Etchings of the Chosen"),
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Chosen: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Sacrifice a creature of the chosen type: Target creature you control gains indestructible until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerSurvive | game.AnswerSacOutlet,
				Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
			},
			Cost:    effects.Plus(effects.ManaCost("{1}"), sacrifice),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"},
					Label: "Etchings of the Chosen — indestructible until end of turn"}.Apply(ctx)
			},
		}},
	})
}
