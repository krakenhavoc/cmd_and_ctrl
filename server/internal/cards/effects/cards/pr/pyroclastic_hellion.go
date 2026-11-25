package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pyroclastic Hellion — Creature — Hellion {4}{R}, 4/5 (Zendikar Rising,
// S82 OLD #5508):
//
//	"When this creature enters, you may return a land you control to its
//	 owner's hand. When you do, this creature deals 2 damage to each
//	 opponent."
//
// The land is chosen as the trigger resolves (zero or one). Returning one
// is what sets off the reflexive trigger (CR 603.12), and the damage is the
// Hellion's, dealt to each opponent at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2151d51-39f6-49cc-a5cc-8202dd028336",
		Name:         "Pyroclastic Hellion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ReturnsLandOnEntry: true},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pyroclastic Hellion — you may return a land you control to its owner's hand; when you do, it deals 2 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChoosePermanents{
						Question: "Pyroclastic Hellion — you may return a land you control to its owner's hand",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							return s82OldB091LandsOf(g, of), 0, 1
						},
						Keyed: pyroclasticHellionReturned,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var pyroclasticHellionReturned = effects.RegisterPermanentsPick("prompt-run/pyroclastic-hellion-return-land", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		if err := effects.BouncePickedToHand(ctx, picked); err != nil {
			return err
		}
		if picked.Count() == 0 {
			return nil
		}
		return effects.S82OldB091DamageEachOpponent(ctx, ctx.Source(), 2)
	},
})

// s82OldB091LandsOf is every land `of` controls.
func s82OldB091LandsOf(g *game.Game, of uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && c.IsLand() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
