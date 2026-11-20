package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alter Reality — Instant {1}{U}:
//
//	"Change the text of target spell or permanent by replacing all
//	 instances of one color word with another. (This effect lasts
//	 indefinitely.)
//	 Flashback {1}{U}"
//
// Sleight of Mind with flashback; the Flashback constructor carries the
// graveyard zone and the exile-on-leaving-the-stack replacement. The
// text change is CR 612.2 (text_words.go, ADR 0160's 2026-10-15
// amendment).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "054c00b4-c940-455a-ad88-b8aa35931ecf",
		Name:             "Alter Reality",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetSpellOrPermanent("target spell or permanent", nil, nil),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{1}{U}")},
		OnResolve:        effects.ChangeTextOfTarget(effects.TextColorWords, false, effects.NothingAfterTheChange),
	})
}
