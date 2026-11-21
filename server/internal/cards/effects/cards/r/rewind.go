package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rewind — Instant {2}{U}{U} (EDHREC rank 1013):
//
//	"Counter target spell. Untap up to four lands."
//
// The free Counterspell: four mana in, four lands back. Counterspell's
// clause, then the refund, in printed order.
//
// "Untap up to four lands" prints no "target" and no "you control",
// so on paper it is a resolution-time choice among every land at the
// table. UntapUpToLands is that choice: a prompt over every tapped
// land at the table, any controller's, queued after the counter
// resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb27bfdf-fe8d-45bd-ad62-8118dce06eda",
		Name:         "Rewind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.B09CounterThenUntapLands(4, "Rewind — untap up to four lands"),
	})
}
