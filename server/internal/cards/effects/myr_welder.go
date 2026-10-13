package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	borrow := HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: ExiledWithLenders(AnyLender),
	})
	borrow.DependsOnExile = true
	Register(Spec{
		OracleID:     "ea2cbf5a-2def-450b-9db8-40a5b6a604d0",
		Name:         "Myr Welder",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Activated: []ActivatedAbility{{
			Label:   "Imprint — {T}: Exile target artifact card from a graveyard.",
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target artifact card from a graveyard", Artifact()),
			Effect:  exileLinkedTargetsEffect,
		}},
	})
}
