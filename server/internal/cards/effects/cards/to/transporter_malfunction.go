package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transporter Malfunction — Sorcery {1}{R} (Star Trek, #4973):
//
//	"Transporter Malfunction deals 3 damage to any target.
//	 Beam me up {3}{R}{R}{R} (You may cast this card from your graveyard
//	 for {3}{R}{R}{R} if you also return a creature you control to its
//	 owner's hand. Then exile this spell.)"
//
// Lightning Bolt's body (damageToFirstTarget) at sorcery speed, and beam
// me up as the set's shared alternative cost (trekB18BeamMeUp):
// flashback's graveyard cast and exile, with the return of a creature the
// caster controls as part of the price.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6eb9865e-513c-43c6-81e1-6ad000e9c48e",
		Name:             "Transporter Malfunction",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetAny(),
		Purpose:          game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:        effects.DamageToFirstTargetStep(3),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.TrekB18BeamMeUp("{3}{R}{R}{R}")},
	})
}
