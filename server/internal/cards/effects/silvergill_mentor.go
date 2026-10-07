package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silvergill Mentor — Creature — Merfolk Wizard {1}{U}, 2/1:
//
//	"As an additional cost to cast this spell, behold a Merfolk or pay {2}. (To behold a Merfolk, choose a Merfolk you control or reveal a Merfolk card from your hand.)
//	 When this creature enters, create a 1/1 white and blue Merfolk creature token."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the behold branch added by its 2026-10-07 amendment
// (BeholdOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {2}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "dcf07f38-422f-46c3-aee9-50540b9c9115",
		Name:           "Silvergill Mentor",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdOrPay("a", "Merfolk", "{2}"),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Silvergill Mentor — create a 1/1 white and blue Merfolk token", Do(CreateToken{Template: TokenCard("1/1 white and blue Merfolk"), N: 1})),
		},
	})
}
