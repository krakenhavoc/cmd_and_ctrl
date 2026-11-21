package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mother Bear — Creature — Bear {1}{G}, 2/2 (Modern Horizons, #5519):
//
//	"{3}{G}{G}, Exile this card from your graveyard: Create two 2/2 green
//	 Bear creature tokens. Activate only as a sorcery."
//
// Dauntless Cathar's graveyard ability: the exile is part of the cost,
// and it works only from the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe742976-ef13-4c25-972f-254b1ed436de",
		Name:         "Mother Bear",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{G}{G}, Exile this card from your graveyard: Create two 2/2 green Bear creature tokens. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}{G}{G}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Bear"), N: 2}),
		}},
	})
}
