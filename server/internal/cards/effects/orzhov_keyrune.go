package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orzhov Keyrune — Artifact {3} (Gatecrash, #5565):
//
//	"{T}: Add {W} or {B}.
//	 {W}{B}: This artifact becomes a 1/4 white and black Thrull artifact
//	 creature with lifelink until end of turn."
//
// Dimir Keyrune's mana ability, and Levitating Statue's animation with
// the colours and lifelink added (gtcKeyruneAnimate). It is still an
// artifact; a Keyrune that came under your control this turn can animate
// but can't attack or tap for mana while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9808c424-0d74-45b8-a092-fd296c6833d2",
		Name:         "Orzhov Keyrune",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{W}{B}: This artifact becomes a 1/4 white and black Thrull artifact creature with lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{W}{B}"),
			Effect: gtcKeyruneAnimate("Orzhov Keyrune — a 1/4 white and black Thrull artifact creature with lifelink until end of turn",
				1, 4, "Thrull", []string{"W", "B"}, "lifelink"),
		}},
	})
}
