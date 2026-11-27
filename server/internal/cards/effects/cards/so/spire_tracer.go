package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spire Tracer — Creature — Elf Scout {G}, 1/1 (Gatecrash, #5565):
//
//	"This creature can't be blocked except by creatures with flying or reach."
//
// Flying's own block rule spelled with its printed words (Prowler's Helm's
// CantBeBlockedExceptBy on itself): a blocker with neither keyword, read
// through the layered view as blockers are declared (CR 509.1b), is refused.
// It is not the flying keyword, so it grants nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b38ab81-7ee7-4d89-878b-3507e44836ca",
		Name:         "Spire Tracer",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.Or(effects.HasKeyword("flying"), effects.HasKeyword("reach")), "creatures with flying or reach"),
		},
	})
}
