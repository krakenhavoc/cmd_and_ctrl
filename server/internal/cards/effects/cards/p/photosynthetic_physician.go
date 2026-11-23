package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Photosynthetic Physician — {4}{G} Creature — Plant Doctor, 5/5 (Star
// Trek, #4973):
//
//	"Reach
//	 {G}, Exile this card from your graveyard: Create a 1/1 red Officer
//	 creature token. Activate only as a sorcery."
//
// Eternal Student's graveyard activation: exiling the card is part of the
// cost (CR 602.2b), so it can be activated once, and only at sorcery
// timing (CR 602.5d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "08787942-456f-4c71-a821-1c90b5639ca1",
		Name:            "Photosynthetic Physician",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{G}, Exile this card from your graveyard: Create a 1/1 red Officer creature token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{G}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Officer"), N: 1}),
		}},
	})
}
