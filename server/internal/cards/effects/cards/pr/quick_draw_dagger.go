package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "47b4dd69-0b2a-4dc3-99e8-f572fe049327",
		Name:            "Quick-Draw Dagger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Quick-Draw Dagger — attach it to target creature you control; it gains first strike until end of turn",
					quickDrawDaggerAttach),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}

// quickDrawDaggerAttach attaches the Dagger to the still-legal target,
// then gives that creature first strike until end of turn.
func quickDrawDaggerAttach(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := effects.AttachSourceToTarget(g, item); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"first strike"},
		Label:    "Quick-Draw Dagger — first strike",
	}.Apply(ctx)
}
