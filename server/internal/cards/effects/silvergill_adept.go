package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silvergill Adept — Creature — Merfolk Wizard {1}{U}, 2/1:
//
//	"As an additional cost to cast this spell, reveal a Merfolk card from your hand or pay {3}.
//	 When this creature enters, draw a card."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the reveal branch added by its 2026-10-07 amendment
// (RevealOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {3}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "cb66c5ec-e7a4-4bd2-b825-6c9b846ba40d",
		Name:           "Silvergill Adept",
		Completeness:   CompletenessFull,
		AdditionalCost: RevealOrPay("a", "Merfolk", "{3}"),
		Purpose:        game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Silvergill Adept — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
