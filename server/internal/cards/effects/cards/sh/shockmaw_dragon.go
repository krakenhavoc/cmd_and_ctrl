package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shockmaw Dragon — Creature — Dragon {4}{R}{R}, 4/4 (Fate Reforged):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, it deals 1 damage to
//	 each creature that player controls."
//
// The damaged player is the one the event names: each creature they control as
// the trigger resolves takes 1, from the Dragon (CR 603.4 reads the Dragon's last
// known state if it has left).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0754b8e-fd3c-4207-9031-a6323e1b66d5",
		Name:            "Shockmaw Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Shockmaw Dragon — it deals 1 damage to each creature that player controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Trigger().Event.Target
					var victims []game.Card
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == player && c.IsCreature() {
							victims = append(victims, c)
						}
					}
					return g.DamageInstanceForEffect(func() error {
						for _, v := range victims {
							if err := (effects.DealDamage{Source: item.SourceCardID, Target: v.InstanceID, Amount: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				}),
		},
	})
}
