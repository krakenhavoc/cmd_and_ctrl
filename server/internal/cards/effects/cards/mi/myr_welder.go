package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Welder — Artifact Creature — Myr {3}, 1/4:
//
//	"Imprint — {T}: Exile target artifact card from a graveyard.
//	 This creature has all activated abilities of all cards exiled
//	 with it."
//
// The imprint exile is linked (ExileTarget.Linked, ADR 0155 §8) and the
// borrow lends every card linked to this Welder object, whatever its type
// now (ExiledWithLenders). An artifact land lends its mana ability too.
// The Welder's own {T} pays for one lent {T} row at a time, as printed.
//
// No simplification.
func init() {
	const clause = "This creature has all activated abilities of all cards exiled with it."
	borrow := effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: effects.ExiledWithLenders(effects.AnyLender),
	})
	borrow.DependsOnExile = true
	effects.Register(effects.Spec{
		OracleID:     "ea2cbf5a-2def-450b-9db8-40a5b6a604d0",
		Name:         "Myr Welder",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Activated: []effects.ActivatedAbility{{
			Label:   "Imprint — {T}: Exile target artifact card from a graveyard.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCardInGraveyard("target artifact card from a graveyard", effects.Artifact()),
			Effect:  effects.ExileLinkedTargetsEffect,
		}},
	})
}
