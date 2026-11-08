package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suspicious Detonation — Sorcery {4}{R} (Murders at Karlov Manor,
// #4469):
//
//	"This spell costs {3} less to cast if you've sacrificed an artifact
//	 this turn.
//	 This spell can't be countered. (This includes by the ward ability.)
//	 Suspicious Detonation deals 4 damage to target creature."
//
// The discount reads the caster's card-type sacrifice tally
// (SacrificedAnArtifactThisTurn, #5224) at CR 601.2f, so it is the
// answer standing as the spell is announced, and it comes off the
// generic part only: the mana value stays 5. Spec.CantBeCountered is
// Abrupt Decay's, and the damage is Bombard's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "91f7672c-3090-48f4-8586-0426341c3159",
		Name:            "Suspicious Detonation",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if you've sacrificed an artifact this turn.",
				func(q game.CostQuery) bool {
					return q.Game != nil && SacrificedAnArtifactThisTurn(q.Game, q.Controller)
				}),
		},
		Targets:   TargetCreature("target creature"),
		Purpose:   ForTargets(DamageToTarget(0, 4)),
		OnResolve: damageToFirstTarget(4),
	})
}
