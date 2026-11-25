package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fall of the Thran — Enchantment — Saga {5}{W} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Destroy all lands.
//	 II, III — Each player returns two land cards from their graveyard to the
//	 battlefield."
//
// I is Armageddon's sweep of every player's lands (an indestructible land
// survives). II and III are Exhume's shape: each player in turn order from the
// controller chooses two land cards from their own graveyard (all they have
// when fewer, none when none), seeing the earlier choices, and the chosen
// lands then enter together under their owners' control, so each sees the
// others enter (CR 603.6a). Chapter III runs the same body.
//
// No simplification.
func init() {
	const name = "Fall of the Thran"
	each := effects.OldB207EachPlayerReturnsTwoLands(name)
	effects.Register(effects.Spec{
		OracleID:     "341803d9-fd29-4721-81bf-ae9b7f1c01d2",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "destroy all lands"),
				effects.Do(effects.DestroyAllMatching{Match: effects.Land()})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "each player returns two land cards from their graveyard to the battlefield"), each),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "each player returns two land cards from their graveyard to the battlefield"), each),
		},
	})
}
