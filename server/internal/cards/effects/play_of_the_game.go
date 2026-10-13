package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Play of the Game — Sorcery {6}{W}{W}:
//
//	"Assist (Another player can pay up to {6} of this spell's cost.)
//	 Exile all nonland permanents."
//
// Planar Cleansing's sweep with the exile verb: every nonland permanent
// on the battlefield, the caster's own included, goes to exile. Lands
// stay. The sweep declares its Purpose the way Farewell's modes do.
func init() {
	Register(Spec{
		OracleID:        "337287d6-cd22-4b70-8bbb-318e3cf91bee",
		Name:            "Play of the Game",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ExileAllMatching{Match: Nonland()}.Apply(ctx)
		},
	})
}
