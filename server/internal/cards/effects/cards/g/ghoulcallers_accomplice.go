package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghoulcaller's Accomplice — Creature — Human Rogue {1}{B}, 2/2 (Jumpstart,
// #5510):
//
//	"{3}{B}, Exile this card from your graveyard: Create a 2/2 black Zombie
//	 creature token. Activate only as a sorcery."
//
// Mother Bear's graveyard ability with one Zombie: the exile is part of the
// cost and it works only from the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16616aa2-5490-4fab-9de4-f37ca8c8be08",
		Name:         "Ghoulcaller's Accomplice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{B}, Exile this card from your graveyard: Create a 2/2 black Zombie creature token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}{B}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}),
		}},
	})
}
