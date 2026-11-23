package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galadhrim Bow — Artifact — Equipment {2}{G} (The Lord of the Rings: Tales
// of Middle-earth):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 Untap that creature.
//	 Equipped creature gets +1/+2 and has reach.
//	 Equip {2}"
//
// Biorganic Carapace's enters trigger (AttachSourceToTarget) with the
// second sentence after it: the untap is the same instruction's second
// half, so it happens to the target even if the attach could not (CR
// 608.2c, the Equipment left in response). A target that is no longer
// legal when the ability resolves does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "92225b59-4f39-4175-ad0d-88e96f4d906a",
		Name:            "Galadhrim Bow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 2), effects.GrantToAttached("reach")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Galadhrim Bow — attach it to target creature you control, then untap that creature",
					func(g *game.Game, item *game.StackItem) error {
						if err := effects.AttachSourceToTarget(g, item); err != nil {
							return err
						}
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return effects.UntapTarget{Target: t.ID}.Apply(ctx)
					}),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
