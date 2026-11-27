package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tethered Skirge — Creature — Phyrexian Imp {2}{B}, 2/2 (Urza's Legacy, S82 OLD b716, #5623):
//
//	"Flying
//	Whenever this creature becomes the target of a spell or ability, you lose 1 life."
//
// Frost Walker's watch (EventBecomesTarget on itself, any spell or ability,
// whoever controls it), with a life loss for the Skirge's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "42bbef66-967b-4922-8b2e-e57d1b8244bb",
		Name:            "Tethered Skirge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Tethered Skirge — you lose 1 life", effects.Do(effects.YouLoseLife(1))),
		},
	})
}
