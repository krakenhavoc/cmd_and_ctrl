package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Suicidal Charge — Enchantment {3}{B}{R} (Conflux):
//
//	"Sacrifice this enchantment: Creatures your opponents control get -1/-1
//	 until end of turn. Those creatures attack this turn if able."
//
// "Those creatures" are the ones that got -1/-1: the set is locked as the
// ability resolves (CR 611.2c), so a creature an opponent puts onto the
// battlefield afterwards is neither shrunk nor forced to attack. The
// requirement is one pinned record over that set (Legion Warboss's token
// shape), read each time attackers are declared (CR 508.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b052ce1-ee82-4335-bc1d-9563bc59b948",
		Name:         "Suicidal Charge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Creatures your opponents control get -1/-1 until end of turn. Those creatures attack this turn if able.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				match := effects.And(effects.Creature(), effects.OpponentControls())
				if err := (effects.BoostUntilEOT{Match: match, Power: -1, Toughness: -1, Label: "Suicidal Charge — -1/-1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.UntilEndOfTurn(ctx, uuid.Nil, match, "Suicidal Charge — attack this turn if able",
					game.AddAttackRequirementMod(uuid.Nil))
			},
		}},
	})
}
