package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Case of the Shifting Visage — Enchantment — Case {1}{U}{U} (Murders at
// Karlov Manor Commander):
//
//	"At the beginning of your upkeep, surveil 1.
//	 To solve — There are fifteen or more cards in your graveyard. (If
//	 unsolved, solve at the beginning of your end step.)
//	 Solved — Whenever you cast a nonlegendary creature spell, copy that
//	 spell. (The copy becomes a token.)"
//
// The upkeep surveil is unconditional. The solve is ToSolve over the
// graveyard count, an intervening if checked as the end-step trigger fires
// and resolves (CR 719.3b, CR 603.4). The copy trigger exists only while the
// Case is solved (ActiveWhen: Solved()): it is Double Down's shape over
// nonlegendary creature spells. "Copy that spell" names the spell without
// targeting it, so it is copied from last-known information if countered
// first (CR 608.2h); the copy is not cast and becomes a token as it resolves
// (CR 707.10a).
//
// No simplification.
func init() {
	copyTrigger := effects.WheneverYouCast(effects.And(effects.Creature(), effects.Not(effects.Legendary())),
		"Case of the Shifting Visage — copy that nonlegendary creature spell",
		effects.Cmdr2S41CopyTheSpellJustCast)
	copyTrigger.ActiveWhen = effects.Solved()
	effects.Register(effects.Spec{
		OracleID:     "409697d5-ce3c-46aa-a97d-969d143a2a65",
		Name:         "Case of the Shifting Visage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Case of the Shifting Visage — surveil 1", effects.Do(effects.Surveil{N: 1})),
			effects.ToSolve("Case of the Shifting Visage — to solve: there are fifteen or more cards in your graveyard", effects.Cmdr2S41FifteenOrMoreCardsInGraveyard),
			copyTrigger,
		},
	})
}
