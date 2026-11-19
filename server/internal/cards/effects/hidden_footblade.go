package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidden Footblade — Artifact — Equipment {R} (Assassin's Creed, #5948):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 That creature gains first strike until end of turn.
//	 Equipped creature gets +1/+0 and has haste.
//	 Equip {2}"
//
// Celestial Armor's shape: the entry trigger targets on the stack, and "that
// creature" is the target, not the equipped creature, so the first strike does
// not depend on the attach; a target that has gone gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "daed1b38-510b-4698-95c3-05c09f5fabe9",
		Name:            "Hidden Footblade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Hidden Footblade — attach it to target creature you control; it gains first strike until end of turn", hiddenFootbladeAttach),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// hiddenFootbladeAttach attaches the Equipment to the still-legal target, then
// gives that creature first strike until end of turn whether or not the attach
// happened.
func hiddenFootbladeAttach(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := AttachSourceToTarget(g, item); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Hidden Footblade — first strike"}.Apply(NewContext(g, item))
}
