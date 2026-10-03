package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Energybending — Instant — Lesson {2}:
//
//	"Lands you control gain all basic land types until end of turn.
//	 Draw a card."
//
// ADR 0109 §1 decision 6 (#1881): gaining a type takes nothing away (CR
// 205.1b, the last sentence of CR 305.7), so each land you control as it
// resolves (CR 611.2c) is a Plains Island Swamp Mountain Forest in addition
// to its own types until end of turn, keeps its abilities, and has all
// five intrinsic mana abilities (CR 305.6). Then you draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4c1e49d-5f14-486c-acf2-1952512b12b2",
		Name:         "Energybending",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (LandBecomes{
				Match:      And(Land(), YouControl()),
				Types:      game.BasicLandTypes,
				InAddition: true,
				Duration:   DurationUntilEndOfTurn(ctx),
				Label:      "Energybending — lands you control have all basic land types",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
