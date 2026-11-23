package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goldmeadow Nomad — Creature — Kithkin Scout {W}, 1/2 (Lorwyn Eclipsed,
// #4458):
//
//	"{W}, Exile this card from your graveyard: Create a 1/1 green and white
//	 Kithkin creature token. Activate only as a sorcery."
//
// Photosynthetic Physician's graveyard activation: exiling the card is
// part of the cost (CR 602.2b), so it can be activated once, and only at
// sorcery timing (CR 602.5d). The token is Catharsis's row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79dfe2eb-82b9-4dae-95c8-58fdf9006bc7",
		Name:         "Goldmeadow Nomad",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{W}, Exile this card from your graveyard: Create a 1/1 green and white Kithkin creature token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{W}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Kithkin"), N: 1}),
		}},
	})
}
