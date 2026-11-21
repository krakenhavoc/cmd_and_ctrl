package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aphetto Alchemist — Creature — Human Wizard {1}{U}, 1/2 (Onslaught):
//
//	"{T}: Untap target artifact or creature.
//	 Morph {U}"
//
// Cat-Owl's untap on a tap ability. The {T} makes it summoning-sick like
// any creature's. Face down it is a 2/2 with no text (CR 708.2a), so the
// ability is not there until it is turned face up for {U}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "867aa0b7-b814-4d33-a677-f6368c8e93c0",
		Name:             "Aphetto Alchemist",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{U}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target artifact or creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Effect:  effects.UntapTheTarget,
		}},
	})
}
