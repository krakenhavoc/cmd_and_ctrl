package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhet-Tomb Mystic — Creature — Zombie Bird Wizard {1}{U}:
//
//	"Flying
//	 Each creature card in your hand has cycling {1}{U}. ({1}{U}, Discard
//	 that card: Draw a card.)"
//
// A hand grant (#3233, ADR 0093's amendment of 2026-10-13) of the
// ordinary Cycling row: the cycled card is a cycling ability's source in
// every way a printed one is, so "whenever you cycle" watchers see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e445825a-e70e-4796-99be-4a7c4080b8cc",
		Name:            "Rhet-Tomb Mystic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key:       "rhet-tomb-mystic/cycling",
			Zone:      game.ZoneHand,
			Activated: []effects.ActivatedAbility{effects.Cycling("{1}{U}")},
			Text:      "Cycling {1}{U}",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilitiesInZone(game.ZoneHand, effects.CardsYouOwn(effects.CreatureCard), "rhet-tomb-mystic/cycling"),
				"Each creature card in your hand has cycling {1}{U}."),
		},
	})
}
