package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nezumi Prowler — Artifact Creature — Rat Ninja {1}{B}, 3/1 (Neon Dynasty):
//
//	"Ninjutsu {1}{B}
//	 When this creature enters, target creature you control gains
//	 deathtouch and lifelink until end of turn."
//
// One targeted trigger; the target (which may be Nezumi Prowler itself) is
// re-checked as it resolves (CR 608.2b), and both keywords are granted
// until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a93b5956-c11e-434e-bed5-94af9de55382",
		Name:         "Nezumi Prowler",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Nezumi Prowler — target creature you control gains deathtouch and lifelink until end of turn",
				effects.FirstTargetGainsUntilEndOfTurn(0, "Nezumi Prowler — deathtouch and lifelink until end of turn", "deathtouch", "lifelink")),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
