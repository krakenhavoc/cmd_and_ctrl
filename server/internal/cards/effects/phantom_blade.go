package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phantom Blade — Artifact — Equipment {4}{B} (Assassin's Creed, #5948):
//
//	"When this Equipment enters, attach it to up to one target creature you
//	 control. Destroy up to one other target creature.
//	 Equipped creature gets +1/+1 and has menace.
//	 Equip {2}"
//
// One entry trigger with two "up to one" target clauses (#764): slot 0 the
// creature you control, slot 1 a different creature, which dies (CR 701.8).
// Either may be left empty, and the two are separate picks: the second must be
// another creature than the first (Distinct). A pick that has gone is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58d5d563-cb60-4977-86cd-f32cd811aa77",
		Name:         "Phantom Blade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Phantom Blade — attach it to up to one target creature you control; destroy up to one other target creature", phantomBladeEnters),
				Clauses(
					UpToOneTargetCreature("up to one target creature you control", YouControl()),
					Distinct(UpToOneTargetCreature("up to one other target creature")),
				),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

func phantomBladeEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if mine, ok := ctx.ClauseTarget(0); ok && mine.Kind == game.TargetCard {
		if err := g.AttachSourceForEffect(item, mine); err != nil {
			return err
		}
	}
	if victim, ok := ctx.ClauseTarget(1); ok && victim.Kind == game.TargetCard {
		return DestroyTarget{Target: victim.ID}.Apply(ctx)
	}
	return nil
}
