package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const topazDragonOracleID = "ca8535f2-1f96-4836-bd32-5e952f9bee72"

// Topaz Dragon // Entropic Cloud — Adventure.
//
//	Topaz Dragon {4}{B}{B} Creature — Dragon 4/4: "Flying, deathtouch"
//	Entropic Cloud {1}{B} Instant — Adventure: "Creatures you control gain
//	deathtouch until end of turn."
//
// The creatures you control as the spell resolves are the ones that gain it
// (CR 611.2c); one that enters later this turn does not.
func init() {
	effects.Register(effects.Spec{
		OracleID:        topazDragonOracleID,
		Name:            "Topaz Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
	effects.Register(effects.Spec{
		OracleID:     topazDragonOracleID + "#1",
		Name:         "Entropic Cloud",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"deathtouch"},
				Label:    "Entropic Cloud — deathtouch",
			}.Apply(ctx)
		},
	})
}
