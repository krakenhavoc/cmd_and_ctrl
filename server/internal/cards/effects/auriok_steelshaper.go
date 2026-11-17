package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auriok Steelshaper — Creature — Human Soldier {1}{W}, 1/1 (Mirrodin,
// #5606):
//
//	"Equip costs you pay cost {1} less.
//	 As long as this creature is equipped, each creature you control
//	 that's a Soldier or a Knight gets +1/+1."
//
// The first line is Bureau Headmaster's equip-ability modifier: it
// prices the equip abilities its controller activates, and only their
// generic part. The second is a Soldier-or-Knight lord (TribalAnthem,
// effective subtypes, the Steelshaper itself included as a Soldier)
// that is on only while an Equipment is attached to the Steelshaper
// (b41SourceIsEquipped, read on every recompute, so it switches off the
// moment the Equipment moves). An Aura is not Equipment.
//
// No simplification.
func init() {
	lord := TribalAnthem(TribeFilter{Tribes: []string{"Soldier", "Knight"}, YoursOnly: true}, 1, 1)
	applies := lord.AppliesTo
	lord.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return b41SourceIsEquipped(g, source) && applies(target, g, source)
	}
	lord.Label = "As long as this creature is equipped, each creature you control that's a Soldier or a Knight gets +1/+1."
	Register(Spec{
		OracleID:     "e3ee8b0e-9bdc-4b9a-b4f9-59189dd6a80d",
		Name:         "Auriok Steelshaper",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			EquipAbilitiesYouActivateCostLess(1, "Equip costs you pay cost {1} less."),
		},
		Static: []game.StaticAbility{lord},
	})
}
