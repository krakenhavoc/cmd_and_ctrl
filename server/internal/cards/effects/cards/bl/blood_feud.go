package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fc401d42-80fa-4a17-937a-c230e9317543",
		Name:         "Blood Feud",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SlotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
