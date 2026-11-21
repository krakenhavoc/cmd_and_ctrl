package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "daed1b38-510b-4698-95c3-05c09f5fabe9",
		Name:            "Hidden Footblade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Hidden Footblade — attach it to target creature you control; it gains first strike until end of turn", hiddenFootbladeAttach),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// hiddenFootbladeAttach attaches the Equipment to the still-legal target, then
// gives that creature first strike until end of turn whether or not the attach
// happened.
func hiddenFootbladeAttach(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := effects.AttachSourceToTarget(g, item); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Hidden Footblade — first strike"}.Apply(effects.NewContext(g, item))
}
