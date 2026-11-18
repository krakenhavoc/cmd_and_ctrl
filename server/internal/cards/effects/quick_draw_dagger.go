package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quick-Draw Dagger — Artifact — Equipment {3} (Streets of New Capenna,
// #5491):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 That creature gains first strike until end of turn.
//	 Equipped creature gets +1/+1.
//	 Equip {1}"
//
// Celestial Armor's entry trigger: it targets on the stack and a target
// that has gone gets nothing (CR 608.2b). "That creature" is the target,
// so the first strike does not depend on the attach.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "47b4dd69-0b2a-4dc3-99e8-f572fe049327",
		Name:            "Quick-Draw Dagger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Quick-Draw Dagger — attach it to target creature you control; it gains first strike until end of turn",
					quickDrawDaggerAttach),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

// quickDrawDaggerAttach attaches the Dagger to the still-legal target,
// then gives that creature first strike until end of turn.
func quickDrawDaggerAttach(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := AttachSourceToTarget(g, item); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"first strike"},
		Label:    "Quick-Draw Dagger — first strike",
	}.Apply(ctx)
}
