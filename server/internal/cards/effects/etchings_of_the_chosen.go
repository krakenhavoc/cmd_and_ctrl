package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Etchings of the Chosen — Enchantment {1}{W}{B}:
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+1.
//	 {1}, Sacrifice a creature of the chosen type: Target creature you
//	 control gains indestructible until end of turn."
func init() {
	sacrifice := SacrificeACreature()
	OfTheChosenType(sacrifice.SacrificeOther)
	Register(Spec{
		OracleID:     "0114f97f-b196-4af1-90aa-eba164cb8f22",
		Name:         "Etchings of the Chosen",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Etchings of the Chosen"),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Chosen: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature of the chosen type: Target creature you control gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), sacrifice),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"},
					Label: "Etchings of the Chosen — indestructible until end of turn"}.Apply(ctx)
			},
		}},
	})
}
