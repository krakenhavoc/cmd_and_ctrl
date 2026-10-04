package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inherited Envelope — Artifact {3}:
//
//	"When this artifact enters, the Ring tempts you.
//	 {T}: Add one mana of any color."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f1ae834-f6b8-458c-ac0d-3510ddb85193",
		Name:         "Inherited Envelope",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Inherited Envelope — the Ring tempts you", Do(TheRingTemptsYou{})),
		},
	})
}
