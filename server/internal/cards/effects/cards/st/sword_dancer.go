package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sword Dancer — Creature — Human Rebel {1}{W} 1/2 (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"{W}{W}: Target attacking creature gets -1/-0 until end of turn."
//
// Anointer of Champions' attacking-creature target with a shrink.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ae49b45-a38f-4bd6-aa08-b5a5da6e75a1",
		Name:         "Sword Dancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}: Target attacking creature gets -1/-0 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, 0)),
			Cost:    effects.ManaCost("{W}{W}"),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostTheTargetUntilEOT(-1, 0, "Sword Dancer — -1/-0 until end of turn")(item, effects.NewContext(g, item))
			},
		}},
	})
}
