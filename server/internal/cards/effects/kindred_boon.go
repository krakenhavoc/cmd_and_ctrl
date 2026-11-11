package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kindred Boon — Enchantment {2}{W}{W}:
//
//	"As this enchantment enters, choose a creature type.
//	 {1}{W}: Put a divinity counter on target creature you control of the
//	 chosen type.
//	 Each creature you control with a divinity counter on it has indestructible."
func init() {
	Register(Spec{
		OracleID:     "f04441b0-755c-4e7b-a9e4-4e03b8175ab6",
		Name:         "Kindred Boon",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Kindred Boon"),
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && target.Counters["divinity"] > 0
			}, "indestructible"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Put a divinity counter on target creature you control of the chosen type.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{1}{W}"),
			Targets: OfTheChosenType(TargetCreature("target creature you control of the chosen type", YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return AddCounter{Target: id, Kind: "divinity", N: 1}.Apply(ctx)
			},
		}},
	})
}
