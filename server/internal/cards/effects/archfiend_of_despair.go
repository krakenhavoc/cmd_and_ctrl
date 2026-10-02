package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archfiend of Despair — Creature — Demon {6}{B}{B}, 6/6:
//
//	"Flying
//	 Your opponents can't gain life.
//	 At the beginning of each end step, each opponent loses life equal
//	 to the life that player lost this turn. (Damage causes loss of
//	 life.)"
//
// "Your opponents can't gain life" is ADR 0107 §5's battlefield static
// (CR 119.7), read at every life gain before the replacement window,
// so a lifelinking opponent's damage is dealt and gains nothing.
//
// The end-step trigger is Wound Reflection's sentence, on every
// player's end step, and shares its body
// (eachOpponentLosesTheLifeTheyLostThisTurn): every opponent's amount
// is read off the turn tally before any is applied.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c36760b-57c5-488b-afc9-ef141942c6ab",
		Name:            "Archfiend of Despair",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CantGainLife:    OpponentsCantGainLife(),
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AnyPlayer, "Archfiend of Despair — each opponent loses the life they lost this turn",
				eachOpponentLosesTheLifeTheyLostThisTurn),
		},
	})
}
