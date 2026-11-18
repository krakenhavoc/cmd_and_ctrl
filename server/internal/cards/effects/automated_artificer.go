package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Automated Artificer — Artifact Creature — Artificer {2}, 1/3 (Kamigawa:
// Neon Dynasty, S82):
//
//	"{T}: Add {C}. Spend this mana only to activate an ability or cast an
//	 artifact spell."
//
// Guidelight Optimizer's restriction (CR 106.6) with {C}: an artifact spell
// or any activated ability, whatever its source. A nonartifact spell is
// refused. The {T} is the creature's own, so a summoning-sick Artificer can't
// use it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "600e1093-5619-49b3-b193-bd263e4fde33",
		Name:         "Automated Artificer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}. Spend this mana only to activate an ability or cast an artifact spell.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate},
			)},
		}},
	})
}
