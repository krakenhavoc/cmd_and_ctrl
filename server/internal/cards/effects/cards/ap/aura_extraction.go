package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aura Extraction — Instant {1}{W} (Onslaught):
//
//	"Put target enchantment on top of its owner's library.
//	 Cycling {2}"
//
// Set Adrift's tuck (putTargetOnTopOfOwnersLibrary) over an
// enchantment, any player's: it goes on top of its OWNER's library,
// whoever controls it, and a commander gets the CR 903.9 offer on the
// way. A token tucked this way ceases to exist (CR 111.8). Cycling is the
// keyword's own hand ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84100106-1f97-4859-aadd-a35796ee000f",
		Name:         "Aura Extraction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
