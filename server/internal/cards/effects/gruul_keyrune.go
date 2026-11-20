package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gruul Keyrune — Artifact {3} (Gatecrash, #5565):
//
//	"{T}: Add {R} or {G}.
//	 {R}{G}: This artifact becomes a 3/2 red and green Beast artifact
//	 creature with trample until end of turn."
//
// Dimir Keyrune's mana ability, and Levitating Statue's animation with
// the colours and trample added (gtcKeyruneAnimate). It is still an
// artifact; a Keyrune that came under your control this turn can animate
// but can't attack or tap for mana while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98bdb1e1-7550-47e8-aa06-f81802b257a0",
		Name:         "Gruul Keyrune",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}{G}: This artifact becomes a 3/2 red and green Beast artifact creature with trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{R}{G}"),
			Effect: gtcKeyruneAnimate("Gruul Keyrune — a 3/2 red and green Beast artifact creature with trample until end of turn",
				3, 2, "Beast", []string{"R", "G"}, "trample"),
		}},
	})
}
