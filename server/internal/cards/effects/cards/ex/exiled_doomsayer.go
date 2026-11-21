package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exiled Doomsayer — Creature — Human Cleric {1}{W}, 1/2 (Scourge, #5844):
//
//	"All morph costs cost {2} more. (This doesn't affect the cost to
//	 cast creature spells face down.)"
//
// A special-action cost increase (SpecialActionCostsMore) on the
// turn-face-up action's morph route only (AMorphCost): CR 702.37e's
// morph cost, a megamorph cost (CR 702.37b: "A megamorph cost is a morph
// cost"), and a manifested card's own morph cost (CR 701.40c). "All" is
// every player's, so there is no "you" predicate. A disguise cost and a
// manifested or cloaked card's mana cost are not morph costs, and the {3}
// face-down cast is a spell's alternative cost, which a special-action
// modifier never reaches. A static of the permanent, so it stops the
// moment Doomsayer leaves the battlefield. Two Doomsayers add {4}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a4c2d47-3489-4e56-b38b-de2dddd942ad",
		Name:         "Exiled Doomsayer",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.SpecialActionCostsMore(2, "All morph costs cost {2} more.", effects.AMorphCost()),
		},
	})
}
