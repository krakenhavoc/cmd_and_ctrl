package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stone Docent — Creature — Spirit Chimera {1}{W}, 3/1 (Secrets of
// Strixhaven, #4456):
//
//	"{W}, Exile this card from your graveyard: You gain 2 life. Surveil 1.
//	 Activate only as a sorcery."
//
// Cryotheory Adept's graveyard ability: the exile is a cost, so the ability
// is on the stack with the card already gone. The surveil is its own step
// behind the life gain (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dbb1690a-f6f5-459c-9d7f-cb6541fb6688",
		Name:         "Stone Docent",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{W}, Exile this card from your graveyard: You gain 2 life. Surveil 1. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{W}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       Do(GainLife{Amount: 2}, Surveil{N: 1}),
		}},
	})
}
