package cardsr

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rethink — Instant {2}{U} (Prophecy, S82 OLD batch 675, #5617):
//
//	"Counter target spell unless its controller pays {X}, where X is its mana
//	 value."
//
// Condescend's pay-unless with X fixed by the spell rather than announced:
// the target spell's mana value is read as the spell resolves (CR 608.2h) and
// its controller is asked for that much generic mana, the table waiting for
// the answer (CR 118.12). A spell on the stack counts {X} in its mana cost as
// the X chosen for it (CR 202.3e). A mana value of 0 means {0}, paid at once.
// With the target illegal as this resolves, nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ede61683-fa78-4103-8248-e0e64162ed49",
		Name:         "Rethink",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell (#4734)
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			spell := ctx.Game.StackItemForEffect(targets[0].ID)
			if spell == nil {
				return nil
			}
			cost := "{" + strconv.Itoa(effects.SpellManaValueForEffect(ctx.Game, spell.SourceCardID)) + "}"
			return effects.CounterUnlessPaid{
				StackID:  targets[0].ID,
				Cost:     cost,
				Question: "Rethink — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
