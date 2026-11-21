package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Ransacked Lab — Enchantment — Case {2}{U} (Murders at
// Karlov Manor, #4469):
//
//	"Instant and sorcery spells you cast cost {1} less to cast.
//	 To solve — You've cast four or more instant and sorcery spells this
//	 turn. (If unsolved, solve at the beginning of your end step.)
//	 Solved — Whenever you cast an instant or sorcery spell, draw a
//	 card."
//
// Archmage of Runes split across a Case (ADR 0071): the cost reduction
// is always on, and the draw trigger exists only once the Case is solved
// (WhenSolved). The solve condition reads the controller's per-turn cast
// tally (CastTally.InstantSorcery, counted as each spell became cast) at
// the end step and again on resolution (CR 603.4). A copy is not cast
// (CR 707.10), so it counts for neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f85eea8f-5887-4921-a943-d15e0ad1e028",
		Name:         "Case of the Ransacked Lab",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.ToSolve("Case of the Ransacked Lab — to solve: you've cast four or more instant and sorcery spells this turn",
				youCastFourInstantsOrSorceriesThisTurn),
			effects.WhenSolved(effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Case of the Ransacked Lab — solved: draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}

// youCastFourInstantsOrSorceriesThisTurn is the Case's solve condition.
func youCastFourInstantsOrSorceriesThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.CastTallyFor(controller).InstantSorcery >= 4
}
