package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illvoi Light Jammer — Artifact — Equipment {1}{U} (Edge of Eternities,
// #4461):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you
//	 control. That creature gains hexproof until end of turn.
//	 Equipped creature gets +1/+2.
//	 Equip {3}"
//
// Barbed Bloodletter's body with hexproof in place of wither: flash, an
// entry trigger that attaches the Equipment to the chosen creature and
// grants the keyword until end of turn, a +1/+2 static and an equip
// ability. The creature gains hexproof even if the Equipment has left by
// then and cannot attach, since "that creature" is the target, not the
// equipped creature. A target gone by resolution does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cbc0b136-544d-45a6-be3e-cd21a573bf19",
		Name:            "Illvoi Light Jammer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Illvoi Light Jammer — attach it to target creature you control; it gains hexproof until end of turn",
					illvoiLightJammerAttachAndHexproof),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

func illvoiLightJammerAttachAndHexproof(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := effects.AttachSourceToTarget(g, item); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Illvoi Light Jammer — hexproof until end of turn"}.Apply(ctx)
}
