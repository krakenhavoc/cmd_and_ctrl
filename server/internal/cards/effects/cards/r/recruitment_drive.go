package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Recruitment Drive — Sorcery {2}{W} (Commander Legends: Battle for Baldur's
// Gate, #5948):
//
//	"Roll a d20.
//	 1—9 | Create two 1/1 white Soldier creature tokens.
//	 10—19 | Create two 2/2 white Knight creature tokens.
//	 20 | Create three 2/2 white Knight creature tokens."
//
// The results table is Contact Other Plane's shape: one d20, then the row it
// landed in. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d37e37d7-ce66-42cb-862e-fc5c414cfeb6",
		Name:         "Recruitment Drive",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.RollDice{Sides: 20, N: 1, Then: recruitmentDriveTable}.Apply(ctx)
		},
	})
}

// recruitmentDriveTable is the results table, after the d20.
var recruitmentDriveTable = effects.DiceTail("dice/recruitment-drive-table", func(ctx *effects.Context, rolls []int) error {
	switch {
	case rolls[0] >= 20:
		return effects.CreateToken{Template: effects.TokenCard("2/2 white Knight"), N: 3}.Apply(ctx)
	case rolls[0] >= 10:
		return effects.CreateToken{Template: effects.TokenCard("2/2 white Knight"), N: 2}.Apply(ctx)
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 2}.Apply(ctx)
})
