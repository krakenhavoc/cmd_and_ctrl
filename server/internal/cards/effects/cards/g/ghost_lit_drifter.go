package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost-Lit Drifter — Creature — Spirit {2}{U}, 2/2 (Modern Horizons 2,
// #5501):
//
//	"Flying
//	 {2}{U}: Another target creature gains flying until end of turn.
//	 Channel — {X}{U}, Discard this card: X target creatures gain flying
//	 until end of turn."
//
// Channel is a hand ability whose X is announced as it is activated and
// whose X targets are chosen with it (CountFromX, Back in Town's shape);
// each target still legal as it resolves gains flying (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8618fec-e124-4f42-b440-0f41a7de3d1f",
		Name:            "Ghost-Lit Drifter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{U}: Another target creature gains flying until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
				Cost:    effects.ManaCost("{2}{U}"),
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect:  effects.S82OldB38GrantFlyingToTargets("Ghost-Lit Drifter — flying until end of turn"),
			},
			{
				Label:   "Channel — {X}{U}, Discard this card: X target creatures gain flying until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
				Cost:    effects.Plus(effects.ManaCost("{X}{U}"), effects.DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Targets: effects.TargetsCountedByX(effects.TargetCreature("X target creatures")),
				Effect:  effects.S82OldB38GrantFlyingToTargets("Ghost-Lit Drifter — channel: flying until end of turn"),
			},
		},
	})
}
