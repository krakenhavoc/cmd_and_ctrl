package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "15febce7-8320-4591-9535-ef11c8adecc2",
		Name:            "Barbed Bloodletter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Barbed Bloodletter — attach it to target creature you control; it gains wither until end of turn",
					barbedBloodletterAttachAndWither),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// barbedBloodletterAttachAndWither attaches the Bloodletter to the
// still-legal target, then gives that creature wither until end of turn
// whether or not the attach happened.
func barbedBloodletterAttachAndWither(g *game.Game, item *game.StackItem) error {
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
		Keywords: []string{game.KeywordWither},
		Label:    "Barbed Bloodletter — wither until end of turn",
	}.Apply(NewContext(g, item))
}
