package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manifold Key — Artifact {1} (Core Set 2020, #5518):
//
//	"{1}, {T}: Untap another target artifact.
//	 {3}, {T}: Target creature can't be blocked this turn."
//
// Sonic Screwdriver's untap and can't-be-blocked abilities, sharing one
// {T}. "Another" is object identity (CR 109.1): a second Key can be
// untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "370c37f7-2484-40e8-8fd9-c10cf6f5d5f2",
		Name:         "Manifold Key",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Untap another target artifact.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.Another(effects.TargetPermanent("another target artifact", effects.Artifact())),
				Effect:  effects.UntapFirstLegalTarget,
			},
			{
				Label:   "{3}, {T}: Target creature can't be blocked this turn.",
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TargetCantBeBlockedThisTurn("Manifold Key"),
			},
		},
	})
}
