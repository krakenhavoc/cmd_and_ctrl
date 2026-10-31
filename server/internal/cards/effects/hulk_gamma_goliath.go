package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hulk, Gamma Goliath — {3}{R}{G} Legendary Creature — Gamma Berserker
// Hero 6/5:
//
//	"Reach, trample
//	 Power-up abilities of other creatures you control cost {3} less to
//	 activate.
//	 Power-up — {6}{R}{G}: Put five +1/+1 counters on Hulk. (Activate
//	 each power-up ability only once. Reduce the cost by his mana cost
//	 if he entered this turn.)"
//
// The discount is an activation-scoped board reduction
// (ActivationCostsLess) that reads the power-up bit on the ability being
// priced and asks that its source be another creature you control, so
// Hulk's own power-up is not discounted by it. That one costs
// {6}{R}{G} less {3}{R}{G}, which is {3}, the turn he entered (CR
// 702.193b: his {R} and {G} take the {R} and {G} off the cost).
//
// No simplification. Marvel Super Heroes (#4455, seam #4527).
func init() {
	Register(Spec{
		OracleID:        "3508fd7d-ca02-47f9-8b0a-525f7bde4f57",
		Name:            "Hulk, Gamma Goliath",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		CostModifiers: []game.CostModifier{
			ActivationCostsLess(3, "Power-up abilities of other creatures you control cost {3} less to activate.",
				APowerUpAbilityCost(), OfAnotherCreatureYouControl()),
		},
		Activated: []ActivatedAbility{
			powerUpAnswering("{6}{R}{G}: Put five +1/+1 counters on Hulk.",
				ManaCost("{6}{R}{G}"), plusOneCountersOnThis(5)),
		},
	})
}
