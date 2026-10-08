package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glorybringer — Creature — Dragon {3}{R}{R}, 4/4:
//
//	"Flying, haste
//	 You may exert this creature as it attacks. When you do, it deals
//	 4 damage to target non-Dragon creature an opponent controls. (An
//	 exerted creature won't untap during your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a targeted
// linked trigger (CR 607.2h). It may be exerted with no legal target:
// the trigger is then removed from the stack and Glorybringer stays
// exerted (CR 603.3d; the Amonkhet ruling). The target is re-checked
// as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b75c3902-633e-4d24-acde-d7a9cc8f466e",
		Name:            "Glorybringer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		ExertOnAttack:   ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(
					WhenExerted("Glorybringer — 4 damage to target non-Dragon creature an opponent controls",
						sourceDealsDamageToEachLegalTarget(4)),
					TargetCreature("target non-Dragon creature an opponent controls", Not(Subtype("Dragon")), OpponentControls())),
				game.Purpose{DamageToCreature: 4}),
		},
	})
}
