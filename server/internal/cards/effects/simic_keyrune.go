package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simic Keyrune — Artifact {3} (Gatecrash, #5565):
//
//	"{T}: Add {G} or {U}.
//	 {G}{U}: This artifact becomes a 2/3 green and blue Crab artifact
//	 creature with hexproof until end of turn. (It can't be the target of
//	 spells or abilities your opponents control.)"
//
// Dimir Keyrune's mana ability, and Levitating Statue's animation with
// the colours and hexproof added (gtcKeyruneAnimate). It is still an
// artifact; a Keyrune that came under your control this turn can animate
// but can't attack or tap for mana while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e737a22-a352-411d-8412-96770c7f6cce",
		Name:         "Simic Keyrune",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G}{U}: This artifact becomes a 2/3 green and blue Crab artifact creature with hexproof until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{G}{U}"),
			Effect: gtcKeyruneAnimate("Simic Keyrune — a 2/3 green and blue Crab artifact creature with hexproof until end of turn",
				2, 3, "Crab", []string{"G", "U"}, "hexproof"),
		}},
	})
}
