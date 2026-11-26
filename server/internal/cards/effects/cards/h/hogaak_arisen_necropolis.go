package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hogaak, Arisen Necropolis — Legendary Creature — Avatar {5}{B/G}{B/G},
// 8/8 (Modern Horizons):
//
//	"You can't spend mana to cast this spell.
//	 Convoke, delve
//	 You may cast this card from your graveyard.
//	 Trample"
//
// Convoke (Spec.TapCost, Convoke()) and delve (Spec.Delve) are the shared
// seams and compose: convoke takes its share of the cost first, delve then
// exiles graveyard cards for generic mana (CR 702.51, CR 702.66). "You
// can't spend mana to cast this spell" is Spec.CantSpendManaToCast: the cast
// is refused while any mana is left in the total cost, so only tapped
// creatures and exiled cards can pay, and the pool is never charged. The
// graveyard is opened by CastableZones. Trample is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "97acd977-164d-49d5-b3a1-72e0d75c9016",
		Name:                "Hogaak, Arisen Necropolis",
		Completeness:        effects.CompletenessFull,
		PrintedKeywords:     []string{"trample"},
		TapCost:             effects.Convoke(),
		Delve:               true,
		CantSpendManaToCast: true,
		CastableZones:       []game.ZoneKind{game.ZoneGraveyard},
	})
}
