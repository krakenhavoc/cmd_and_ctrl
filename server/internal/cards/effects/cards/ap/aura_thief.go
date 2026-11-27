package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aura Thief — Creature — Illusion {3}{U}, 2/2 (Urza's Destiny, S82 #5621):
//
//	"Flying
//	 When this creature dies, you gain control of all enchantments. (You don't get
//	 to move Auras.)"
//
// Control changes for good (CR 611.2a, a layer-2 effect with no end) over every
// enchantment on the battlefield as the trigger resolves, the ones the Thief's
// controller already controls included (nothing changes for them). An Aura whose
// control changes stays attached to what it enchants (CR 303.4e): only its
// controller changes, which is the parenthetical. An enchantment that enters later
// is not taken, and one that left in response is not (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf0936ea-2741-43a4-ad6a-a1258b54945b",
		Name:            "Aura Thief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Aura Thief — you gain control of all enchantments", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item).AsGroupMember()
				var ids []game.Card
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsEnchantment() && c.Controller != item.Controller {
						ids = append(ids, c)
					}
				}
				for _, c := range ids {
					if err := (effects.GainControl{Target: c.InstanceID, Duration: game.IndefiniteDuration(),
						Label: "Aura Thief — you gain control of it"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
