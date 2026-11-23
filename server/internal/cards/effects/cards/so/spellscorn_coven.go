package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spellscorn Coven // Take It Back — Creature — Faerie Warlock {3}{B},
// 2/3 // Instant — Adventure {2}{U} (Wilds of Eldraine, #4471):
//
//	Spellscorn Coven — "Flying
//	 When this creature enters, each opponent discards a card."
//	Take It Back — "Return target spell to its owner's hand. (Then exile
//	 this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The entry is Burglar Rat's eachOpponentDiscardsOne. Take It
// Back is Reprieve's verb (ReturnSpellToHand, through Venser's shared
// body): the spell is returned, not countered, so a spell that can't be
// countered still goes, and a target that left the stack in response
// is not chased.
//
// No simplification.
const SpellscornCovenOracleID = "9357fc7f-c9d0-4ce2-a6dd-0ff1f24bfc56"

func init() {
	effects.Register(effects.Spec{
		OracleID:        SpellscornCovenOracleID,
		Name:            "Spellscorn Coven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spellscorn Coven — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     SpellscornCovenOracleID + "#1",
		Name:         "Take It Back",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnTargetSpellOrPermanentToHand(ctx.Game, item)
		},
	})
}
