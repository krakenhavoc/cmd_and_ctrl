package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitsune Palliator — Creature — Fox Cleric {2}{W}, 0/2 (Betrayers of
// Kamigawa, #5602):
//
//	"{T}: Prevent the next 1 damage that would be dealt to each creature and
//	 each player this turn."
//
// One charged one-point shield (PreventNextDamage, CR 615.7) on each
// creature on the battlefield and each player in the game as the ability
// resolves; each is its own shield, used up by the first damage that
// object takes. A creature that enters later is not covered.
//
// No simplification.
func init() {
	const label = "Kitsune Palliator: prevent the next 1 damage"
	effects.Register(effects.Spec{
		OracleID:     "ed854708-5a65-4293-bebc-d7c639407ba5",
		Name:         "Kitsune Palliator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to each creature and each player this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, c := range effects.MatchingBattlefield(ctx, effects.Creature()) {
					if err := (effects.PreventNextDamage{Target: c.InstanceID, Amount: 1, Label: label}).Apply(ctx); err != nil {
						return err
					}
				}
				for _, p := range g.Seats {
					if p == nil || p.Eliminated {
						continue
					}
					if err := (effects.PreventNextDamage{Target: p.ID, Amount: 1, Label: label}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
