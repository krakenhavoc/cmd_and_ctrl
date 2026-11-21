package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crush of Wurms — Sorcery {6}{G}{G}{G} (Judgment, #5611, #5789):
//
//	"Create three 6/6 green Wurm creature tokens.
//	 Flashback {9}{G}{G}{G} (You may cast this card from your graveyard
//	 for its flashback cost. Then exile it.)"
//
// Roar of the Wurm at three tokens: the table's plain 6/6 green Wurm
// row, and the shared Flashback constructor (graveyard cast, price, and
// CR 702.34a's exile as the spell leaves the stack).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6c12cd05-b673-453c-9f6b-f3aa022467c1",
		Name:             "Crush of Wurms",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tokens: 3, CreatureTokens: 3},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{9}{G}{G}{G}")},
		OnResolve:        effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("6/6 green Wurm"), N: 3}),
	})
}
