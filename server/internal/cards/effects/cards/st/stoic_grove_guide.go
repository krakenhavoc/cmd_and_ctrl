package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stoic Grove-Guide — Creature — Elf Druid {4}{B/G} (Lorwyn Eclipsed, #4458):
//
//	"{1}{B/G}, Exile this card from your graveyard: Create a 2/2 black and
//	 green Elf creature token. Activate only as a sorcery."
//
// Renew's shape without the ability word: it works from the graveyard and
// exiles the card as part of its cost.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96c4b51c-8a10-428c-a948-e1ba9c652bfd",
		Name:         "Stoic Grove-Guide",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B/G}, Exile this card from your graveyard: Create a 2/2 black and green Elf creature token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{B/G}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black and green Elf"), N: 1}),
		}},
	})
}
