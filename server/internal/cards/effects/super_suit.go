package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Super Suit — {1}{U} Artifact — Equipment (Marvel Super Heroes, #4455):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 Untap that creature.
//	 Equipped creature gets +1/+2.
//	 Equip {2}"
//
// Falcon's Wing Harness's entry trigger, with the target untapped once the
// attach has been made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "606092eb-8fe3-4ef7-aaf1-2d3b2687d27a",
		Name:            "Super Suit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{PumpAttached(1, 2)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Super Suit — attach it to target creature you control and untap that creature", superSuitEnters),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// superSuitEnters attaches the Equipment to the target creature, then
// untaps that creature.
func superSuitEnters(g *game.Game, item *game.StackItem) error {
	if err := AttachSourceToTarget(g, item); err != nil {
		return err
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return UntapTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}
