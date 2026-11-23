package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nothic — {4}{B} Creature — Horror 4/3
//
//	"Weird Insight — When this creature dies, roll a d20.
//	 1—9 | You draw a card and you lose 1 life.
//	 10—19 | You draw two cards and you lose 2 life.
//	 20 | You draw seven cards and you lose 7 life."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec8c2927-d857-426e-b81b-c3b1aa81045a",
		Name:         "Nothic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Weird Insight — roll a d20",
				effects.Do(effects.RollDice{Sides: 20, N: 1, Then: nothicTable})),
		},
	})
}

var nothicTable = effects.DiceTail("dice/nothic-table", func(ctx *effects.Context, rolls []int) error {
	n := 1
	switch {
	case rolls[0] == 20:
		n = 7
	case rolls[0] >= 10:
		n = 2
	}
	if err := (effects.DrawCards{N: n}).Apply(ctx); err != nil {
		return err
	}
	return effects.YouLoseLife(n)(ctx)
})
