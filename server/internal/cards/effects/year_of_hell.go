package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		if err := b05EachPlayerDraws(g, item, 1); err != nil {
			return err
		}
		return b38DamageEachPlayer(NewContext(g, item), 2)
	}
	text := "each player draws a card; this Saga deals 2 damage to each player"
	Register(Spec{
		OracleID:     "83f426db-3b25-4407-afd0-0465e3900c24",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, text), chapter),
			ChapterTrigger(2, SagaChapterLabel(name, 2, text), chapter),
			ChapterTrigger(3, SagaChapterLabel(name, 3, text), chapter),
		},
	})
}
