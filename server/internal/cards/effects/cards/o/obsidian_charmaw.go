package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obsidian Charmaw — Creature — Dragon {3}{R}{R}, 4/4 (Modern Horizons 2):
//
//	"This spell costs {1} less to cast for each land your opponents control that
//	 could produce {C}.
//	 Flying
//	 When this creature enters, destroy target nonbasic land an opponent controls."
//
// The reduction counts, as the spell is priced, the lands the caster's
// opponents control whose mana could include {C} (CR 106.7) and takes it off the
// generic part of the cost. The enters trigger targets a nonbasic land an
// opponent controls as it goes on the stack and destroys it as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c995bc01-f278-4088-9ae2-1b50f87b2a61",
		Name:            "Obsidian Charmaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.OldB035OpponentLandsThatCouldProduceColorless,
				"This spell costs {1} less to cast for each land your opponents control that could produce {C}."),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Obsidian Charmaw — destroy target nonbasic land an opponent controls", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, id := range effects.LegalTargetIDs(ctx) {
						return effects.DestroyTarget{Target: id}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPermanent("target nonbasic land an opponent controls", effects.NonbasicLand(), effects.OpponentControls()),
			),
		},
	})
}
