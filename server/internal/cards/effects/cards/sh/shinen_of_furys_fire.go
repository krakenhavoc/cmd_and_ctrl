package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shinen of Fury's Fire — Creature — Spirit {2}{R}, 2/1 (Saviors of
// Kamigawa, #5601):
//
//	"Haste
//	 Channel — {R}, Discard this card: Target creature gains haste until
//	 end of turn."
//
// Haste is printed data. The channel row works only from the hand, the
// discard part of its cost; the grant ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b3b512e-6849-4d2e-9352-a36bc1646a92",
		Name:            "Shinen of Fury's Fire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {R}, Discard this card: Target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Shinen of Fury's Fire — haste until end of turn", 0, 0, "haste"),
		}},
	})
}
