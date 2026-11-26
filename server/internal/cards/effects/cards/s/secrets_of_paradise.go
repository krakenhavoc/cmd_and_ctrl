package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secrets of Paradise — Conspiracy (Conspiracy, S82 OLD batch 361, #5557):
//
//	"Hidden agenda (Start the game with this conspiracy face down in the
//	 command zone and secretly choose a card name. You may turn this
//	 conspiracy face up any time and reveal that name.)
//	 Creatures you control with the chosen name have "{T}: Add one mana of
//	 any color.""
//
// Hidden agenda is the engine's (ADR 0227); the grant is ADR 0093's layer-6
// grant, applied to the creatures the conspiracy's controller controls that
// have the chosen name, and only while it is face up (CR 315.5).
//
// No simplification.
const secretsOfParadiseGrant = "secrets-of-paradise/any-color"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "f309fe9f-0732-4b81-a342-a394225cbbb1",
		Name:         "Secrets of Paradise",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Grants:       []effects.AbilityGrant{effects.AnyColorManaGrant(secretsOfParadiseGrant)},
		Static:       []game.StaticAbility{effects.GrantAbilities(effects.CreaturesYouControlWithTheChosenName, secretsOfParadiseGrant)},
	})
}
