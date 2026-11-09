package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agatha of the Vile Cauldron — Legendary Creature — Human Warlock {R}{G},
// 2/2 (Wilds of Eldraine, #4471):
//
//	"Activated abilities of creatures you control cost {X} less to
//	 activate, where X is Agatha's power. This effect can't reduce the
//	 mana in that cost to less than one mana.
//	 {4}{R}{G}: Other creatures you control get +1/+1 and gain trample
//	 and haste until end of turn."
//
// Training Grounds' reduction (KeepsOneMana over an activation discount on
// creatures you control, mana abilities included) with X read from
// Agatha's power as the cost is determined, layers applied, so a pump
// deepens the discount. The pump affects the other creatures you control
// as the ability resolves and locks that set (CR 611.2c): a creature
// arriving later gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f398d222-f8e7-4983-9a77-267adef7f551",
		Name:         "Agatha of the Vile Cauldron",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			KeepsOneMana(ActivationCostsLessEach(
				func(q game.CostQuery) int { return max(0, q.Source.CurrentPower()) },
				"Activated abilities of creatures you control cost {X} less to activate, where X is Agatha's power. This effect can't reduce the mana in that cost to less than one mana.",
				OfACreatureYouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}{G}: Other creatures you control get +1/+1 and gain trample and haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{4}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				match := And(Creature(), YouControl(), OtherThan(item.SourceCardID))
				if err := (BoostUntilEOT{Match: match, Power: 1, Toughness: 1, Label: "Agatha of the Vile Cauldron — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample", "haste"},
					Label: "Agatha of the Vile Cauldron — trample and haste"}.Apply(ctx)
			},
		}},
	})
}
