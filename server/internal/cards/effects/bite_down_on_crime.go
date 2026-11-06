package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bite Down on Crime — Sorcery {3}{G} (Murders at Karlov Manor, #5193):
//
//	"As an additional cost to cast this spell, you may collect evidence 6.
//	 This spell costs {2} less to cast if evidence was collected. (To
//	 collect evidence 6, exile cards with total mana value 6 or greater from
//	 your graveyard.)
//	 Target creature you control gets +2/+0 until end of turn. It deals
//	 damage equal to its power to target creature you don't control."
//
// The collect evidence offer is OptionalCollectEvidence, and the discount
// reads the announcement at CR 601.2f (CostsLessIfEvidenceCollected, off
// game.CostQuery.OptionalCosts), so the {2} comes off the total the caster
// pays and the mana value stays 4. The body is Diplomatic Relations': the
// pump lands first and the bite reads the grown power as it deals
// (oneSidedBite), each clause re-checked on its own (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ff387264-31da-4f6b-88c6-2843f56acbda",
		Name:          "Bite Down on Crime",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalCollectEvidence(6)},
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfEvidenceCollected(2, "This spell costs {2} less to cast if evidence was collected"),
		},
		Purpose: game.Purpose{Answers: game.AnswerRemove, // a one-sided fight, as Bite Down (#4734)
			Targets: game.ForTargets(BenefitUntilEOT(0, 2, 0))},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: mine.ID, Power: 2, Label: "Bite Down on Crime — +2/+0 until end of turn"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return oneSidedBite(ctx.Item, ctx) }),
		),
	})
}
