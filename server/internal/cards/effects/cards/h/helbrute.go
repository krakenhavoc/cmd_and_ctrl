package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helbrute — Artifact Creature — Astartes Dreadnought {3}{B}{R}, 5/4:
//
//	"Haste
//	 Sarcophagus — You may cast this card from your graveyard by exiling
//	 another creature card from your graveyard in addition to paying its other
//	 costs."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b216b4d-f9cb-4ce2-af70-c91eef3db083",
		Name:            "Helbrute",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		CastableZones:   []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			effects.S82AfrB16CastFromGraveyardByExiling("{3}{B}{R}", 1, "creature", effects.Creature()),
		},
	})
}
