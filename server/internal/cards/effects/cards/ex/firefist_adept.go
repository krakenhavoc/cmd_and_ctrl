package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firefist Adept — Creature — Human Wizard {4}{R}, 3/3 (Dominaria, S82 OLD,
// #5530):
//
//	"When this creature enters, it deals X damage to target creature an
//	 opponent controls, where X is the number of Wizards you control."
//
// X is counted as the trigger resolves (CR 608.2h), through the layered view,
// so the Adept counts itself (it is a Wizard) and a changeling or a token
// Wizard counts. A target that is gone or no longer an opponent's creature
// is not hit (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f513242-b175-4b4d-a05f-78b1b457a020",
		Name:         "Firefist Adept",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Firefist Adept — damage to target creature an opponent controls equal to the number of Wizards you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						x := len(effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype(ctx.TextSubtype("Wizard")))))
						if x <= 0 {
							return nil
						}
						t, ok := ctx.ClauseTarget(0)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}.Apply(ctx)
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
