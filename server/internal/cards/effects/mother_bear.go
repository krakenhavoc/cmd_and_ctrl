package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fe742976-ef13-4c25-972f-254b1ed436de",
		Name:         "Mother Bear",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{3}{G}{G}, Exile this card from your graveyard: Create two 2/2 green Bear creature tokens. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{G}{G}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
			Effect:       Do(CreateToken{Template: TokenCard("2/2 green Bear"), N: 2}),
		}},
	})
}
