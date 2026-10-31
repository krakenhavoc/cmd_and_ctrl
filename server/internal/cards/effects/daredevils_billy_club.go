package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daredevil's Billy Club — Artifact — Equipment {1}{R} (Marvel Super
// Heroes Commander, #4455):
//
//	"When this Equipment enters, it deals 2 damage to any target and 1
//	 damage to you.
//	 Equipped creature gets +1/+0 and has menace.
//	 Equip {2}"
//
// The enters trigger names one any-target; the 2 damage to it and the 1
// damage to the controller are dealt together as one damage event (CR
// 120.2), so a target that left leaves only the damage to you (CR 608.2b).
// The pump and menace are the usual attached statics.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "654db13d-c7e4-4ee0-a87a-3635467f2f19",
		Name:         "Daredevil's Billy Club",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Daredevil's Billy Club — 2 damage to any target and 1 damage to you",
				msc5BillyClubDamage), TargetAny()),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// msc5BillyClubDamage deals 2 damage to each target still legal and 1 to
// the controller, all in one damage event.
func msc5BillyClubDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.DamageInstanceForEffect(func() error {
		for _, t := range ctx.LegalTargets() {
			if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(ctx)
	})
}
