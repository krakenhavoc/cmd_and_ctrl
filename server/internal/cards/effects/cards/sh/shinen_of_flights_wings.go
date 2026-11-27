package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shinen of Flight's Wings — Creature — Spirit {4}{U}, 3/3 (Saviors of Kamigawa, S82 OLD b584,
// #5601):
//
//	"Channel — {U}, Discard this card: Target creature gains flying until
//	 end of turn."
//
// Shinen of Fury's Fire's channel row with another keyword: it works only
// from the hand, the discard being part of its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5fa542e5-d3a7-45d3-ac1d-32546fbfe784",
		Name:            "Shinen of Flight's Wings",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {U}, Discard this card: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Shinen of Flight's Wings — flying until end of turn", 0, 0, "flying"),
		}},
	})
}
