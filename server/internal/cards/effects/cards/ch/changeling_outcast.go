package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Changeling Outcast — Creature — Shapeshifter {B}, 1/1:
//
//	"Changeling (This card is every creature type.)
//	 This creature can't block and can't be blocked."
//
// Changeling (CR 702.73a) needs no catalog entry — the deck importer
// stamps it from Scryfall's printed keywords like any other one, and
// `Card.HasSubtype` already answers true for every creature type in
// every zone (S26). The catalog entry exists only for the restriction
// line: `RestrictSelf` is the S24 "this permanent prints a
// restriction on itself" primitive (Carrion Feeder's shape), and the
// two bits OR together in one static the way `Characteristic.Restrictions`
// always combines them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a61ef0cc-1da9-49f9-b0dc-01cf9f6205be",
		Name:         "Changeling Outcast",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock | game.CantBeBlocked)},
	})
}
