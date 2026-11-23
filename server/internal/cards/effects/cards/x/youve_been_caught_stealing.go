package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// You've Been Caught Stealing — Sorcery {1}{R} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Choose one —
//	 • Threaten the Merchant — Each creature blocks this turn if able.
//	 • Bribe the Guards — You create a Treasure token for each opponent who was
//	   dealt damage this turn."
//
// Threaten the Merchant puts the blocks-if-able requirement on every creature
// on the battlefield as the spell resolves, whoever controls it (Culling Mark's
// requirement, CR 509.1c; fixed then, CR 611.2c). Bribe the Guards counts the
// opponents of the caster the damage events of this turn show were dealt damage
// (a damage that was prevented was not dealt), once each however many times,
// and makes that many Treasures.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "750b61a1-7eb1-4ce2-8a56-fa63f8e48755",
		Name:         "You've Been Caught Stealing",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Threaten the Merchant — Each creature blocks this turn if able.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if !c.IsCreature() {
							continue
						}
						if err := (effects.BlockRequirementUntilEOT{
							Target: c.InstanceID,
							Kind:   game.BlockRequirementBlocks,
							Label:  "You've Been Caught Stealing — blocks this turn if able",
						}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Bribe the Guards — You create a Treasure token for each opponent who was dealt damage this turn.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					n := effects.S82BnsB12OpponentsDealtDamage(ctx.Game, item.Controller)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TreasureToken(), N: n}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerValue}),
		),
	})
}
