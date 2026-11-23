package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ee35ed62-d5a2-4e65-a7eb-fcdcb3665532",
		Name:         "Balefire Liege",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("R"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("W"), 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WheneverYouCast(effects.OfColor("R"), "Balefire Liege — 3 damage to target player or planeswalker",
				effects.SourceDealsDamageToEachLegalTarget(3)), effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.WheneverYouCast(effects.OfColor("W"), "Balefire Liege — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
