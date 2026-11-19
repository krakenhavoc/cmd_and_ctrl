package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Feud — Sorcery {4}{R}{R} (Foundations Jumpstart, #5929):
//
//	"Target creature fights another target creature."
//
// Two target clauses, the second Distinct from the first ("another"), as
// Pit Fight has them but with no controller restriction on either. The fight
// is slotZeroFightsSlotOne: nothing fights once either creature has left or
// become illegal (CR 701.14b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc401d42-80fa-4a17-937a-c230e9317543",
		Name:         "Blood Feud",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("another target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return slotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
