package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Skyscout — Creature — Merfolk Scout {2}{U}{U}, 2/3 (Conspiracy:
// Take the Crown, S82 older sets batch 272, #5540):
//
//	"Flying
//	 Whenever this creature attacks or blocks, untap target permanent."
//
// One trigger over both events (Clockwork's ThisAttacksOrBlocks: its own
// attack declaration, or the first of its blocks, CR 509.3a). The target is
// any permanent, chosen as the trigger goes on the stack and untapped as it
// resolves if still legal (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Merfolk Skyscout — untap target permanent"
	effects.Register(effects.Spec{
		OracleID:        "381eb146-879e-473c-b628-65f6c0814a29",
		Name:            "Merfolk Skyscout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks, label, effects.UntapEachLegalTarget),
				effects.TargetPermanent("target permanent", effects.Permanent())),
		},
	})
}
