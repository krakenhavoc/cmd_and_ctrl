package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lone Revenant — Creature — Spirit {3}{U}{U}, 4/4 (Commander 2015, S82 OLD
// #5546):
//
//	"Hexproof
//	 Whenever this creature deals combat damage to a player, if you control no
//	 other creatures, look at the top four cards of your library. Put one of
//	 them into your hand and the rest on the bottom of your library in any
//	 order."
//
// The "if" is intervening (CR 603.4): checked as the damage is dealt and again
// as the trigger resolves, both against the creatures the controller has
// besides the Revenant. The look is private to the controller; one card is
// taken (not optional: with any card to take, one must be) and the rest go to
// the bottom in an order the controller picks (Beastrider Vanguard's dig).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d814b0b-82e8-4145-87e1-4fed44d16465",
		Name:            "Lone Revenant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.AllOf(effects.ThisDealtCombatDamageToAPlayer, loneRevenantAlone),
				"Lone Revenant — look at the top four cards of your library, put one into your hand and the rest on the bottom",
				loneRevenantLook),
		},
	})
}

// loneRevenantAlone is the intervening "if you control no other creatures",
// read as the damage is dealt.
func loneRevenantAlone(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return controlsNoOtherCreature(g, source.Controller, source.InstanceID)
}

func controlsNoOtherCreature(g *game.Game, player, self uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.InstanceID != self && c.IsCreature() {
			return false
		}
	}
	return true
}

// loneRevenantLook re-checks the condition as it resolves (CR 603.4).
var loneRevenantLook = effects.Do(effects.If{
	Cond: effects.CondBoard{Name: "you control no other creatures", Test: func(ctx *effects.Context) bool {
		return controlsNoOtherCreature(ctx.Game, ctx.Controller(), ctx.Source())
	}},
	Then: []effects.Applier{effects.Instruction(func(ctx *effects.Context) error {
		player := ctx.Controller()
		return effects.TakeFromLibraryToHand{
			Player: player,
			Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
			Max:    1,
			Label:  "Lone Revenant — put one of them into your hand",
			Key:    effects.TakeRestBottomAny,
		}.Apply(ctx)
	})},
})
