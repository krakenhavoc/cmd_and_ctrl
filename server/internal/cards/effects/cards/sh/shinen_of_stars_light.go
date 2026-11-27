package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shinen of Stars' Light — Creature — Spirit {2}{W}, 2/1 (Saviors of Kamigawa, S82 OLD b584,
// #5601):
//
//	"Channel — {1}{W}, Discard this card: Target creature gains first strike until
//	 end of turn."
//
// Shinen of Fury's Fire's channel row with another keyword: it works only
// from the hand, the discard being part of its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d43f813b-3607-4bc3-b996-3dbad6977bd8",
		Name:            "Shinen of Stars' Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {1}{W}, Discard this card: Target creature gains first strike until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike")),
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Shinen of Stars' Light — first strike until end of turn", 0, 0, "first strike"),
		}},
	})
}
