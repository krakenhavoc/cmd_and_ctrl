package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Year of Hell — Enchantment — Saga {1}{U}{R} (Star Trek, #4973):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II, III — Each player draws a card. This Saga deals 2 damage to
//	 each player."
//
// Three chapter abilities with one text (ChapterTrigger, saga.go). Each
// player draws in turn order from the active player (Faerie Mastermind's
// b05EachPlayerDraws), then the Saga deals 2 damage to each player as one
// simultaneous damage event (Cast Away Doubt's b38DamageEachPlayer). The
// damage source is the Saga, as it last existed if it has gone (CR
// 608.2h).
//
// No simplification.
func init() {
	const name = "Year of Hell"
	chapter := func(g *game.Game, item *game.StackItem) error {
		if err := effects.B05EachPlayerDraws(g, item, 1); err != nil {
			return err
		}
		return effects.B38DamageEachPlayer(effects.NewContext(g, item), 2)
	}
	text := "each player draws a card; this Saga deals 2 damage to each player"
	effects.Register(effects.Spec{
		OracleID:     "83f426db-3b25-4407-afd0-0465e3900c24",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, text), chapter),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, text), chapter),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, text), chapter),
		},
	})
}
