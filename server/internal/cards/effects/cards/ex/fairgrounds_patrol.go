package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fairgrounds Patrol — Creature — Human Soldier {1}{W}, 2/1 (Modern Horizons
// 2, #5501):
//
//	"{1}{W}, Exile this card from your graveyard: Create a 1/1 colorless
//	 Thopter artifact creature token with flying. Activate only as a
//	 sorcery."
//
// A graveyard ability (Morbius's shape): the exile is a cost, so the card is
// gone before the ability resolves and nothing can return it in response.
// Sorcery speed means its controller's main phase with an empty stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b0689b5-7c68-46b3-a72a-4f97ef1d0a15",
		Name:         "Fairgrounds Patrol",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{W}, Exile this card from your graveyard: Create a 1/1 colorless Thopter artifact creature token with flying. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:         effects.Plus(effects.ManaCost("{1}{W}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
		}},
	})
}
