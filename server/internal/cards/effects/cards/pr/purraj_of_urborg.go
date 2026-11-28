package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Purraj of Urborg — Legendary Creature — Cat Warrior {3}{B}{B}, 2/3 (Mirage,
// S82 OLD b805):
//
//	"Purraj has first strike as long as it's attacking.
//	 Whenever a player casts a black spell, you may pay {B}. If you do, put a
//	 +1/+1 counter on Purraj."
//
// The first strike is Kor Scythemaster's attacking-status static. The trigger
// is Iron Star's shape: any player's black spell (Purraj's controller
// included), the {B} asked as the trigger resolves, and the counter put on
// only if it is paid and Purraj is still on the battlefield.
//
// No simplification.
func init() {
	firstStrike := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "first strike")
	firstStrike.Label = "Purraj has first strike as long as it's attacking."
	firstStrike.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:     "ae86965c-3330-48bf-abf7-3142c61099ea",
		Name:         "Purraj of Urborg",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{firstStrike},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("B")),
				"Purraj of Urborg — you may pay {B}: put a +1/+1 counter on Purraj",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{B}",
						Question: "Purraj of Urborg — pay {B} to put a +1/+1 counter on Purraj?",
						Then:     purrajPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// purrajPaid is the paid branch: the counter, if Purraj is still there. An
// on-disk key: never renamed, never reused.
var purrajPaid = effects.PayUnlessBranches("pay-unless/s82-old-b805-purraj-counter",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !effects.OnBattlefield(ctx.Game, ctx.Item.SourceCardID) {
			return nil
		}
		return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}, nil)
