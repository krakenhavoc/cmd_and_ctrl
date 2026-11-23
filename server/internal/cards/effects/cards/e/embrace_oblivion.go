package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embrace Oblivion — Sorcery {B} (Edge of Eternities, #4461):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Destroy target creature or Spacecraft."
//
// Bone Splinters with Deadly Dispute's "an artifact or creature" cost
// and Gravkill's target clause: "Spacecraft" is the artifact subtype, so
// a Spacecraft below its station threshold can be destroyed too. The
// sacrifice is paid as the spell is cast (CR 601.2h), so it is gone even
// if the target becomes illegal and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0fad5713-c09c-43d3-9cc0-10c65aead488",
		Name:           "Embrace Oblivion",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Targets:        effects.TargetPermanent("target creature or Spacecraft", effects.Or(effects.Creature(), effects.HasSubtype("Spacecraft"))),
		OnResolve:      effects.DestroyTheTargetPermanent,
	})
}
