package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minion Reflector — {5} Artifact (Shards of Alara, S82 OLD b516):
//
//	"Whenever a nontoken creature you control enters, you may pay {2}. If you
//	 do, create a token that's a copy of that creature, except it has haste and
//	 'At the beginning of the end step, sacrifice this permanent.'"
//
// Kindle the Inner Flame's copy (haste on the token's copiable values and the
// quoted ability granted as part of the copying process, CR 707.9a), made
// behind an optional payment asked as the trigger resolves. The creature is
// copied as it last existed if it has left by then (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Minion Reflector — you may pay {2}; if you do, create a token that's a copy of that creature"
	effects.Register(effects.Spec{
		OracleID:     "0124977f-06e6-430a-a874-f7d05883a97c",
		Name:         "Minion Reflector",
		Completeness: effects.CompletenessFull,
		Grants:       []effects.AbilityGrant{effects.OldB516SacrificeThisPermanentAtEndStepGrant(minionReflectorGrant, "Minion Reflector — sacrifice this permanent")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && !c.IsToken()
			}, label, func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Minion Reflector — pay {2} to create a hasty token copy of that creature?",
					Then:     minionReflectorPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

const minionReflectorGrant = "minion-reflector/sacrifice-this-permanent"

var minionReflectorPaid = effects.PayUnlessBranches("pay-unless/old-b516-minion-reflector-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || ctx.Item.Trigger == nil {
			return nil
		}
		return effects.CreateTokenCopy{
			Controller: ctx.Item.Controller,
			Copy:       ctx.Item.Trigger.Event.CardID,
			N:          1,
			Except: func(t *game.Card) {
				effects.TokenCopyGainsHaste(t)
				effects.TokenCopyGainsAbility(t, minionReflectorGrant)
			},
		}.Apply(ctx)
	}, nil)
