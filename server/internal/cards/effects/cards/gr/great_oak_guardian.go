package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Oak Guardian — Creature — Treefolk {5}{G}, 4/5 (Commander 2015,
// #5546):
//
//	"Flash
//	 Reach
//	 When this creature enters, creatures target player controls get +2/+2
//	 until end of turn. Untap them."
//
// The enters trigger targets a player (any player, the controller
// included) and is Trystan's Command's last bullet with +2/+2: the creatures
// that player controls as it resolves get the pump (CR 611.2c) and are
// untapped. A target that has left the game resolves to nothing.
//
// No simplification.
func init() {
	etb := effects.WhenThisEnters("Great Oak Guardian — creatures target player controls get +2/+2 and untap",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if err := (effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.ControlledByPlayer(t.ID)), Power: 2, Toughness: 2, Label: "Great Oak Guardian — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				return effects.B16UntapAllYouControlMatching(ctx, t.ID, func(c game.Card) bool { return c.IsCreature() })
			}
			return nil
		})
	etb.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:        "5dba7746-c087-4ec1-92ca-ab9882dea24d",
		Name:            "Great Oak Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
		Triggered:       []game.TriggeredAbility{etb},
	})
}
