package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bark of Doran — Artifact — Equipment {1}{W} (Lorwyn Eclipsed, #4458):
//
//	"Equipped creature gets +0/+1.
//	 As long as equipped creature's toughness is greater than its power, it
//	 assigns combat damage equal to its toughness rather than its power.
//	 Equip {1}"
//
// The pump is PumpAttached. The damage rule is Ghalta's DamageByToughness
// flag, set by a layer 6 static on the equipped creature only: the combat
// damage pass compares toughness with power per creature at damage time
// (CR 510.1c), so "as long as toughness is greater than power" is exactly
// that test, and nothing but the amount of damage assigned changes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db3ba239-2239-4e9d-ad70-cff32ed24578",
		Name:         "Bark of Doran",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(0, 1),
			{
				Layer:     game.Layer6Ability,
				Label:     "As long as equipped creature's toughness is greater than its power, it assigns combat damage equal to its toughness rather than its power.",
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.DamageByToughness = true
				},
			},
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
