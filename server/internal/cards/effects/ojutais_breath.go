package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ojutai's Breath — Instant {2}{U}:
//
//	"Tap target creature. It doesn't untap during its controller's next
//	 untap step.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Frost Breath's tap-and-freeze for one creature. Rebound is the
// engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2c09d296-5d15-449a-9b98-cf5056a17b91",
		Name:            "Ojutai's Breath",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b751TapFreezeTargets(ctx, uuid.Nil, "Ojutai's Breath")
		},
	})
}
