package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kabira Crossroads — Land (Foundations Jumpstart, #5929):
//
//	"This land enters tapped.
//	 When this land enters, you gain 2 life.
//	 {T}: Add {W}."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3dbb16f-fa8f-4406-bcf3-e647e4337619",
		Name:         "Kabira Crossroads",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Key:       "Kabira Crossroads — you gain 2 life",
			Effect:    effects.Do(effects.GainLife{Amount: 2}),
		}},
	})
}
