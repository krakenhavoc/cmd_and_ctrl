package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "606092eb-8fe3-4ef7-aaf1-2d3b2687d27a",
		Name:            "Super Suit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Super Suit — attach it to target creature you control and untap that creature", superSuitEnters),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// superSuitEnters attaches the Equipment to the target creature, then
// untaps that creature.
func superSuitEnters(g *game.Game, item *game.StackItem) error {
	if err := effects.AttachSourceToTarget(g, item); err != nil {
		return err
	}
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return effects.UntapTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}
