package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dimir Infiltrator — Creature — Spirit {U}{B}, 1/3 (Ravnica: City of
// Guilds, #5600):
//
//	"This creature can't be blocked.
//	 Transmute {1}{U}{B} ({1}{U}{B}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal it,
//	 put it into your hand, then shuffle. Transmute only as a sorcery.)"
//
// Blighted Agent's unblockable restriction (CR 509.1b) with the transmute
// constructor (transmute.go, CR 702.53a), finding a mana value 2 card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d215127b-5f1d-4d3c-935b-c228c9c7f874",
		Name:         "Dimir Infiltrator",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Activated:    []effects.ActivatedAbility{effects.Transmute("{1}{U}{B}")},
	})
}
