package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Buried Treasure — Artifact — Treasure {2}:
//
//	"{T}, Sacrifice this artifact: Add one mana of any color.
//	 {5}, Exile this card from your graveyard: Discover 5. Activate only
//	 as a sorcery."
//
// The first ability is a Treasure's. The second works from the
// GRAVEYARD (CR 113.6, embalm's shape): the card exiles itself as the
// cost. Discover is ADR 0099's (game/discover.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56f5ed8d-ef53-4766-9ef3-1e24a10e267b",
		Name:         "Buried Treasure",
		Completeness: effects.CompletenessFull,
		Discovers:    true,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}, Exile this card from your graveyard: Discover 5",
			Purpose:      game.Purpose{Discover: 5},
			Cost:         effects.Plus(effects.ManaCost("{5}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.DiscoverN(5),
		}},
	})
}
