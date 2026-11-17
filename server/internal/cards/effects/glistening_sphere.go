package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glistening Sphere — Artifact {3} (Phyrexia: All Will Be One Commander):
//
//	"This artifact enters tapped.
//	 When this artifact enters, proliferate.
//	 {T}: Add one mana of any color.
//	 Corrupted — {T}: Add three mana of any one color. Activate only if an
//	 opponent has three or more poison counters."
//
// Coveted Jewel's OneColorOfAmount(3) behind the corrupted activation
// condition; the proliferate is the controller's choice (#2525).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15679808-baba-4a11-b111-f313ce24197f",
		Name:         "Glistening Sphere",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glistening Sphere — proliferate", Do(Proliferate{})),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost:      ManaAbilityCost{Tap: true},
				Produced:  OneColorOfAmount(3),
				Label:     "Corrupted — Add three mana of any one color",
				Condition: corruptedAnd(nil),
			},
		},
	})
}
