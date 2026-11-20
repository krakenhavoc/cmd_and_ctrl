package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Secretkeeper // Venture Deeper — Creature — Merfolk Wizard {U}, 0/4 // Sorcery — Adventure {U} (Throne of Eldraine, #5516):
//
//	Merfolk Secretkeeper — no rules text.
//	Venture Deeper — "Target player mills four cards."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const merfolkSecretkeeperOracleID = "4f5096d4-bbb5-4a48-9dd0-9dabdd2ac125"

func init() {
	Register(Spec{
		OracleID:     merfolkSecretkeeperOracleID,
		Name:         "Merfolk Secretkeeper",
		Completeness: CompletenessFull,
	})
	Register(Spec{
		OracleID:     merfolkSecretkeeperOracleID + "#1",
		Name:         "Venture Deeper",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return TargetPlayerMills(4)(ctx.Game, item)
		},
	})
}
