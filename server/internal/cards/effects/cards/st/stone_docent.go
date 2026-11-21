package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "dbb1690a-f6f5-459c-9d7f-cb6541fb6688",
		Name:         "Stone Docent",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{W}, Exile this card from your graveyard: You gain 2 life. Surveil 1. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{W}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.GainLife{Amount: 2}, effects.Surveil{N: 1}),
		}},
	})
}
