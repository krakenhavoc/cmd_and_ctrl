package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kabira Crossroads — Land (Foundations Jumpstart, #5929):
//
//	"This land enters tapped.
//	 When this land enters, you gain 2 life.
//	 {T}: Add {W}."
func init() {
	Register(Spec{
		OracleID:     "b3dbb16f-fa8f-4406-bcf3-e647e4337619",
		Name:         "Kabira Crossroads",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Key:       "Kabira Crossroads — you gain 2 life",
			Effect:    Do(GainLife{Amount: 2}),
		}},
	})
}
