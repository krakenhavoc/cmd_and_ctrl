package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silver Shroud Costume — Artifact — Equipment {2} (Fallout, #5948):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 That creature gains shroud until end of turn.
//	 Equipped creature can't be blocked.
//	 Equip {3}"
//
// Biorganic Carapace's enter-and-attach trigger, then the creature that
// was targeted gains shroud until end of turn whether or not the attach
// went through (it is "that creature", CR 608.2b, and only a target that is
// no longer legal gets nothing). Whispersilk Cloak's restriction on the
// equipped creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b40138c3-48bc-4825-ae48-5fea95304088",
		Name:            "Silver Shroud Costume",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Silver Shroud Costume — attach it to target creature you control; that creature gains shroud until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						target, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						if err := AttachSourceToTarget(g, item); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{Target: target.ID, Keywords: []string{"shroud"}, Label: "Silver Shroud Costume — shroud"}.Apply(ctx)
					}),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Static:    []game.StaticAbility{RestrictAttached(game.CantBeBlocked)},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
