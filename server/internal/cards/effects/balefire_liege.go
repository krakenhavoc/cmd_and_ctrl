package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balefire Liege — Creature — Spirit Horror {2}{R/W}{R/W}{R/W}, 2/4:
//
//	"Other red creatures you control get +1/+1.
//	 Other white creatures you control get +1/+1.
//	 Whenever you cast a red spell, this creature deals 3 damage to target
//	 player or planeswalker.
//	 Whenever you cast a white spell, you gain 3 life."
//
// Glen Elendra Liege's two colour anthems (a red-white creature gets
// +2/+2) and Aragorn's cast triggers: a red-white spell triggers both, a
// colourless one neither. The damage target is chosen as the trigger goes
// on the stack and re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee35ed62-d5a2-4e65-a7eb-fcdcb3665532",
		Name:         "Balefire Liege",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(g2OtherCreaturesYouControlOfColor("R"), 1, 1),
			b16Anthem(g2OtherCreaturesYouControlOfColor("W"), 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WheneverYouCast(OfColor("R"), "Balefire Liege — 3 damage to target player or planeswalker",
				sourceDealsDamageToEachLegalTarget(3)), targetPlayerOrPlaneswalker()), ForTargets(DamageToTarget(0, 3))),
			WheneverYouCast(OfColor("W"), "Balefire Liege — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
