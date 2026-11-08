package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agatha of the Vile Cauldron — Legendary Creature — Human Warlock {R}{G},
// 1/1 (Wilds of Eldraine, #4471):
//
//	"Activated abilities of creatures you control cost {X} less to
//	 activate, where X is Agatha's power. This effect can't reduce the
//	 mana in that cost to less than one mana.
//	 {4}{R}{G}: Other creatures you control get +1/+1 and gain trample and
//	 haste until end of turn."
//
// The discount is Training Grounds' activation reduction with the
// one-mana floor (KeepsOneMana), its amount read off Agatha's current
// power as Vadrik's is (the cost query's Source is the live permanent),
// so a pumped Agatha discounts more and one at 0 or less discounts
// nothing. It reduces generic mana only (CR 118.7a) and reaches every
// creature you control on the battlefield, Agatha's own ability
// included. The pump is Scurry of Gremlins' shape over the other
// creatures you control as it resolves (CR 611.2c), "other" read by
// object identity as Draconautics Engineer reads it.
//
// No simplification.
func init() {
	const pumpLabel = "{4}{R}{G}: Other creatures you control get +1/+1 and gain trample and haste until end of turn."
	Register(Spec{
		OracleID:     "f398d222-f8e7-4983-9a77-267adef7f551",
		Name:         "Agatha of the Vile Cauldron",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			KeepsOneMana(ActivationCostsLessEach(func(q game.CostQuery) int { return max(q.Source.CurrentPower(), 0) },
				"Activated abilities of creatures you control cost {X} less to activate, where X is Agatha's power. This effect can't reduce the mana in that cost to less than one mana.",
				OfACreatureYouControl())),
		},
		Activated: []ActivatedAbility{{
			Label: pumpLabel,
			// "Other creatures" is not a whole-team pump (the TeamPump guard
			// reads "Creatures you control"), so it declares only what it
			// answers, as Draconautics Engineer's haste grant does.
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    ManaCost("{4}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				others := And(Creature(), YouControl(), trekB11OtherThanSourceObject(item))
				if err := (BoostUntilEOT{Match: others, Power: 1, Toughness: 1, Label: "Agatha of the Vile Cauldron — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: others, Keywords: []string{"trample", "haste"}, Label: "Agatha of the Vile Cauldron — trample and haste"}.Apply(ctx)
			},
		}},
	})
}
