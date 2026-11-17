package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defiler of Flesh — Creature — Phyrexian Horror {2}{B}{B}, 4/4 (Dominaria United,
// #5659):
//
//	"Menace
//	 As an additional cost to cast black permanent spells, you may pay 2
//	 life. Those spells cost {B} less to cast if you paid life this way.
//	 This effect reduces only the amount of black mana you pay.
//	 Whenever you cast a black permanent spell, target creature you control
//	 gets +1/+1 and gains menace until end of turn."
//
// The offer and its discount are the cycle's linked pair
// (pay_life_for_discount.go): each black permanent spell its controller
// casts may be announced with "pay 2 life", and one that is costs {B}
// less, never less generic mana. The target is
// chosen as the trigger goes on the stack and checked again as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:             "3ad002de-8be3-48d2-a0e0-865538144789",
		Name:                 "Defiler of Flesh",
		Completeness:         CompletenessFull,
		PrintedKeywords:      []string{"menace"},
		GrantedOptionalCosts: []game.GrantedOptionalCost{MayPayLifeToCast("Defiler of Flesh", "B", 2)},
		CostModifiers:        []game.CostModifier{CostsColorLessIfLifePaid("B")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WheneverYouCastAPermanentSpellOfColor("B",
				"Defiler of Flesh — target creature you control gets +1/+1 and gains menace until end of turn",
				defilerOfFleshPump), TargetCreature("target creature you control", YouControl())),
				ForTargets(BenefitUntilEOT(0, 1, 1, "menace"))),
		},
	})
}

// defilerOfFleshPump gives the target +1/+1 and menace until end of turn.
func defilerOfFleshPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Defiler of Flesh — +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"menace"}, Label: "Defiler of Flesh — menace"}.Apply(ctx)
}
