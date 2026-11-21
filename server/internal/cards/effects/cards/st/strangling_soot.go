package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strangling Soot — Instant {2}{B} (Time Spiral, #5596):
//
//	"Destroy target creature with toughness 3 or less.
//	 Flashback {5}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The target's toughness is read after the layers, at announce and again
// at resolution (CR 608.2b), so a creature pumped past 3 in response is no
// longer a legal target. Flashback (CR 702.34a) costs {5}{R} from the
// graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6c9d3ad0-1ede-41de-a1b4-15de9c728216",
		Name:             "Strangling Soot",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature with toughness 3 or less", effects.ToughnessLE(3)),
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{R}")},
		OnResolve:        effects.DestroyTheTargetPermanent,
	})
}
