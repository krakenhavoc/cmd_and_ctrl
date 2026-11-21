package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "654db13d-c7e4-4ee0-a87a-3635467f2f19",
		Name:         "Daredevil's Billy Club",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Daredevil's Billy Club — 2 damage to any target and 1 damage to you",
				msc5BillyClubDamage), effects.TargetAny()),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// msc5BillyClubDamage deals 2 damage to each target still legal and 1 to
// the controller, all in one damage event.
func msc5BillyClubDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return g.DamageInstanceForEffect(func() error {
		for _, t := range ctx.LegalTargets() {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(ctx)
	})
}
