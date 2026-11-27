package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost-Lit Nourisher — Creature — Spirit {2}{G}, 2/1 (Saviors of Kamigawa,
// S82 OLD b584, #5601):
//
//	"{2}{G}, {T}: Target creature gets +2/+2 until end of turn.
//	 Channel — {3}{G}, Discard this card: Target creature gets +4/+4 until
//	 end of turn."
//
// Both rows are Shinen of Fury's Fire's boost-until-end-of-turn shape; the
// channel row works only from the hand, the discard part of its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acaed013-2e74-47e3-a687-084388b37a7b",
		Name:         "Ghost-Lit Nourisher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{G}, {T}: Target creature gets +2/+2 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
				Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.NeoB07TargetGetsUntilEOT("Ghost-Lit Nourisher", 2, 2),
			},
			{
				Label:   "Channel — {3}{G}, Discard this card: Target creature gets +4/+4 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
				Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.NeoB07TargetGetsUntilEOT("Ghost-Lit Nourisher", 4, 4),
			},
		},
	})
}
