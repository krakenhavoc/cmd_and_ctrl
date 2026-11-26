package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindclaw Shaman — Creature — Lizard Shaman {4}{R}, 2/2 (Magic 2013, S82 OLD, #5568):
//
//	"When this creature enters, target opponent reveals their hand. You may
//	 cast an instant or sorcery spell from among those cards without paying
//	 its mana cost."
//
// The hand is revealed to the table, then the cast is part of the resolution
// (CR 608.2g): Extract Brain's resolution-cast prompt out of the opponent's hand,
// free, at any timing, one spell, instants and sorceries only (the clause is
// judged on the face cast, CR 601.3e). Additional costs are still paid and an X is
// 0 (CR 107.3b). A card not cast stays in its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "018985f5-6a67-4f2f-855c-efa7254ad3f8",
		Name:         "Mindclaw Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Mindclaw Shaman — target opponent reveals their hand; you may cast an instant or sorcery spell from among those cards without paying its mana cost",
				effects.S82OldB429RevealAndFreeCast),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
