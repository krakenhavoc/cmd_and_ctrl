package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauntless Cathar — Creature — Human Soldier {2}{W}, 3/2 (Innistrad
// Remastered):
//
//	"{1}{W}, Exile this card from your graveyard: Create a 1/1 white Spirit
//	 creature token with flying. Activate only as a sorcery."
//
// Works from the graveyard only; the exile is part of the cost.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c14bd133-ff17-48b3-a1c8-56c93d8c6e43",
		Name:         "Dauntless Cathar",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{W}, Exile this card from your graveyard: Create a 1/1 white Spirit creature token with flying. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{W}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}),
		}},
	})
}
