package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saw It Coming — {1}{U}{U} Instant: "Counter target spell.
// Foretell {1}{U}"
//
// Counterspell with a keyword, and the keyword is the whole card: you
// spend {2} on your own turn to hide it in exile, and from the next
// turn on it answers a spell for {1}{U} with two mana you were never
// holding up. The engine carries every part of that (#658, CR
// 702.143) — the special action, the owner-only face-down exile and
// the per-instance cast permission priced at the foretell cost — so
// the card file is a counter clause and one line.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90edaf33-d0ab-47e0-8f6a-6fba38286e6e",
		Name:         "Saw It Coming",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		SpecialActions: []game.SpecialAction{
			effects.Foretell("{1}{U}"),
		},
		OnResolve: effects.CounterTheTargetSpell,
	})
}
