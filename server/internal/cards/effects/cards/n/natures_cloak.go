package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nature's Cloak — {2}{G} Sorcery (Starter 1999, S82 OLD batch 701, #5620):
//
//	"Green creatures you control gain forestwalk until end of turn."
//
// The creatures are fixed as the spell resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb493c95-5f6d-405b-aac6-c56fe6b6f42f",
		Name:         "Nature's Cloak",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl(), effects.OfColor("G")),
				Keywords: []string{"forestwalk"},
				Label:    "Nature's Cloak — forestwalk",
			}),
	})
}
