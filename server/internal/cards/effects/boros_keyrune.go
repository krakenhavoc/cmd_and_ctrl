package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boros Keyrune — Artifact {3} (Gatecrash, #5565):
//
//	"{T}: Add {R} or {W}.
//	 {R}{W}: This artifact becomes a 1/1 red and white Soldier artifact
//	 creature with double strike until end of turn. (It deals both
//	 first-strike and regular combat damage.)"
//
// Dimir Keyrune's mana ability, and Levitating Statue's animation with
// the colours and double strike added (gtcKeyruneAnimate). It is still an
// artifact; a Keyrune that came under your control this turn can animate
// but can't attack or tap for mana while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "991ae2f7-3f8e-4364-8698-5277786d7dcd",
		Name:         "Boros Keyrune",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}{W}: This artifact becomes a 1/1 red and white Soldier artifact creature with double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{R}{W}"),
			Effect: gtcKeyruneAnimate("Boros Keyrune — a 1/1 red and white Soldier artifact creature with double strike until end of turn",
				1, 1, "Soldier", []string{"R", "W"}, "double strike"),
		}},
	})
}
