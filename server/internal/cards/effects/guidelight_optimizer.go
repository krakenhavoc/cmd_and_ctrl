package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guidelight Optimizer — Artifact Creature — Robot {1}{U}, 2/1
// (Aetherdrift, #4464):
//
//	"{T}: Add {U}. Spend this mana only to cast an artifact spell or
//	 activate an ability."
//
// Isolinear Processor's restriction (CR 106.6) with the second half
// widened: casting an artifact spell, or activating any ability, whatever
// its source. Casting a nonartifact spell with it is refused. The {T} is
// the creature's own, so a summoning-sick Optimizer can't use it
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "045488b3-0206-4dcd-91ab-59ec542f06c0",
		Name:         "Guidelight Optimizer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "{T}: Add {U}. Spend this mana only to cast an artifact spell or activate an ability.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate},
			)},
		}},
	})
}
