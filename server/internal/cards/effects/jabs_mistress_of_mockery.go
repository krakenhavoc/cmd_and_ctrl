package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jabs, Mistress of Mockery — Legendary Creature — Human Rogue {2}{R},
// 2/3 (Mystery Booster Commander Edition):
//
//	"If a triggered ability of a permanent you control but don't own
//	 triggers, that ability triggers an additional time."
//
// Delney's shape (CR 603.2d). The source's controller must be Jabs's
// controller (the doubler's default check) and its owner (immutable, read off the card) must be someone
// else, both read from the source as it was when the ability triggered, so
// a permanent that left the battlefield in the same event still counts. A
// permanent Jabs's controller owns is not doubled, and neither is a
// permanent an opponent controls.
//
// No simplification.
func init() {
	doubler := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{
		SourceMatch: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			return q.SourceLKI.Controller == q.DoublerLKI.Controller && q.Source.Owner != q.SourceLKI.Controller
		},
	})
	doubler.Label = "Jabs, Mistress of Mockery"
	doubler.RowLabel = "Triggered abilities of permanents you control but don't own trigger an additional time."
	Register(Spec{
		OracleID:        "723959e1-8356-4fdb-99e2-818df317dc72",
		Name:            "Jabs, Mistress of Mockery",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
