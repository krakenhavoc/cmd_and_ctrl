package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bone Splinters — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Destroy target creature."
//
// The sacrifice is paid while casting (CR 601.2h), after the target is
// chosen, so the creature sacrificed cannot usefully be the target: it is
// gone before the spell resolves and the target is illegal then. With no
// creature to sacrifice the cast is refused.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0c936582-d4c0-4da8-bc7e-49da5a433939",
		Name:           "Bone Splinters",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetPermanent("target creature", effects.Creature()),
		OnResolve:      effects.DestroyTheTargetPermanent,
	})
}
