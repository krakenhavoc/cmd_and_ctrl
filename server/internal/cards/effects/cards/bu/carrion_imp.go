package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carrion Imp — Creature — Imp {3}{B}, 2/3 (Ravnica Allegiance):
//
//	"Flying
//	 When this creature enters, you may exile target creature card from a graveyard.
//	 If you do, you gain 2 life."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d); "you may" is
// asked as it resolves. The 2 life follows only an exile that happened.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bba44fad-b3ca-4892-b7d1-2d03478062af",
		Name:            "Carrion Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Carrion Imp — you may exile target creature card from a graveyard; if you do, you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneExile {
							return nil
						}
						return effects.GainLife{Amount: 2}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature())),
				"Carrion Imp — exile the creature card?"),
		},
	})
}
