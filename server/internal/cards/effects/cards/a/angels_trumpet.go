package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angel's Trumpet — Artifact {3}:
//
//	"All creatures have vigilance.
//	 At the beginning of each player's end step, tap all untapped creatures
//	 that player controls that didn't attack this turn. This artifact deals
//	 damage to the player equal to the number of creatures tapped this way."
//
// The first sentence is a layer-6 grant to every creature. The trigger reads
// the active player (whose end step it is), taps their untapped creatures that
// did not attack, and deals damage equal to those that really became tapped.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f187f766-3696-419e-9865-58023a46446b",
		Name:         "Angel's Trumpet",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			}, "vigilance"), "All creatures have vigilance."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"Angel's Trumpet — tap their creatures that didn't attack; it deals damage to them equal to the number tapped",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					if player == uuid.Nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == player && c.IsCreature() && !c.Tapped && !g.AttackedThisTurn(c.InstanceID) {
							ids = append(ids, c.InstanceID)
						}
					}
					tapped := 0
					for _, id := range ids {
						if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
						if c, ok := g.LookupCardForEffect(id); ok && c.Tapped {
							tapped++
						}
					}
					if tapped == 0 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: tapped}.Apply(ctx)
				}),
		},
	})
}
