package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dregscape Sliver — Creature — Sliver {1}{B}, 2/2 (Secret Lair Drop):
//
//	"Each Sliver creature card in your graveyard has unearth {2}.
//	 Unearth {2} ({2}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the next end
//	 step or if it would leave the battlefield. Unearth only as a sorcery.)"
//
// The printed unearth is the ordinary Unearth row. The grant is Sliver
// Gravemother's graveyard grant of an activated ability, with the unearth row
// at {2} and the recipients every Sliver creature card its controller owns in a
// graveyard. The grant is the card's own, so it ends when the Sliver leaves.
//
// No simplification.
func init() {
	const key = "dregscape-sliver/unearth"
	effects.Register(effects.Spec{
		OracleID:     "d1678eed-a335-454c-9fa1-749317500a36",
		Name:         "Dregscape Sliver",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Unearth("{2}")},
		Grants: []effects.AbilityGrant{{
			Key:       key,
			Zone:      game.ZoneGraveyard,
			Activated: []effects.ActivatedAbility{effects.Unearth("{2}")},
			Text:      "Unearth {2}",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilitiesInZone(game.ZoneGraveyard, effects.OldB013YourSliverCreatureCard, key),
				"Each Sliver creature card in your graveyard has unearth {2}."),
		},
	})
}
