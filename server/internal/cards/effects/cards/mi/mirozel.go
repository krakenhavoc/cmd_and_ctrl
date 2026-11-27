package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirozel — Creature — Illusion {3}{U}, 2/3 (Exodus, S82 OLD batch 745, #5626):
//
//	"Flying
//	 When this creature becomes the target of a spell or ability, return
//	 this creature to its owner's hand."
//
// Angelic Protector's "becomes the target" row (EventBecomesTarget on the
// source, one trigger per targeting spell or ability), returning it to its
// owner's hand as the trigger resolves, above the spell that targets it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c35111ce-4325-4547-b954-08275a268060",
		Name:            "Mirozel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self,
				"Mirozel — return this creature to its owner's hand",
				effects.ReturnThisPermanentToOwnersHand),
		},
	})
}
