package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbed Bloodletter — Artifact — Equipment {1}{B} (Lorwyn Eclipsed,
// #4458):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 That creature gains wither until end of turn. (It deals damage to
//	 creatures in the form of -1/-1 counters.)
//	 Equipped creature gets +1/+2.
//	 Equip {2}"
//
// Celestial Armor's shape with wither for the rider. The entry trigger
// targets on the stack; a target that has gone gets nothing (CR 608.2b).
// "That creature" is the target, not the equipped creature: if the
// Bloodletter is gone in response the attach does nothing (CR 701.3b) and
// the still-legal creature still gains wither. Wither is read off the
// damage source's abilities as the damage is dealt (ADR 0056), so the
// grant changes combat and fight damage for the rest of the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15febce7-8320-4591-9535-ef11c8adecc2",
		Name:            "Barbed Bloodletter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Barbed Bloodletter — attach it to target creature you control; it gains wither until end of turn",
					barbedBloodletterAttachAndWither),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}

// barbedBloodletterAttachAndWither attaches the Bloodletter to the
// still-legal target, then gives that creature wither until end of turn
// whether or not the attach happened.
func barbedBloodletterAttachAndWither(g *game.Game, item *game.StackItem) error {
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
		Keywords: []string{game.KeywordWither},
		Label:    "Barbed Bloodletter — wither until end of turn",
	}.Apply(effects.NewContext(g, item))
}
