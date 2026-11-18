package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geothermal Kami — Creature — Spirit {3}{G}, 4/3 (Kamigawa: Neon Dynasty,
// S82):
//
//	"When this creature enters, you may return an enchantment you control to
//	 its owner's hand. If you do, you gain 3 life."
//
// A resolution-time pick over your own enchantments (no target), Ambrosia
// Whiteheart's shape: zero is a legal answer, and the life is gained only for
// an enchantment that actually went to the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbf8fdf5-0d7d-41a3-a6b9-27bb7a644c20",
		Name:         "Geothermal Kami",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Geothermal Kami — you may return an enchantment you control to its owner's hand; if you do, gain 3 life",
				func(g *game.Game, item *game.StackItem) error {
					return ChoosePermanents{
						Question: "Geothermal Kami — you may return an enchantment you control to its owner's hand",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							var out []uuid.UUID
							for _, c := range g.BattlefieldCardsForEffect() {
								if c.Controller == of && c.IsEnchantment() {
									out = append(out, c.InstanceID)
								}
							}
							return out, 0, 1
						},
						Keyed: geothermalKamiReturned,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// geothermalKamiReturned returns the chosen enchantment and, if it reached
// the hand, gains 3 life. An on-disk key.
var geothermalKamiReturned = RegisterPermanentsPick("prompt-run/geothermal-kami-return", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !onBattlefield(ctx.Game, id) {
				continue
			}
			if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
				return GainLife{Amount: 3}.Apply(ctx)
			}
		}
		return nil
	},
})
