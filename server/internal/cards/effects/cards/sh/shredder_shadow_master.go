package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shredder, Shadow Master — {3}{B}{B} Legendary Creature — Human Ninja
// 5/5 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Whenever Shredder attacks a player, for each other opponent, create
//	 a token that's a copy of Shredder tapped and attacking that player,
//	 except it isn't legendary. Sacrifice those tokens at end of combat.
//	 Whenever Shredder deals combat damage to a player, that player loses
//	 half their life, rounded up."
//
// Myriad's shape (CR 702.116a) made mandatory and narrowed. The attack
// trigger fires only for an attack on a player, never one on a
// planeswalker or battle (CR 508.3a). Each other opponent still in the
// game gets one copy attacking them directly, all made by one creation
// (a doubler's extra tokens attack the same player). The copies take
// Shredder's copiable values (CR 707.2), or his last-known ones if he
// left in response (CR 608.2h, the printed ruling), minus the legendary
// supertype (CR 707.9b); tapped, counters, Auras and non-copy effects are
// not copied. They enter attacking without being declared, so they fire
// no attack trigger of their own (CR 508.4). At the beginning of the end
// of combat step his controller sacrifices them (CR 511.2). Each copy
// has the combat-damage ability: the player it hits loses half their life
// as it is then, rounded up (CR 107.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee82947d-1cf0-4c53-84d4-81f81c6201ee",
		Name:         "Shredder, Shadow Master",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.ThisAttacked(ev, source, lki, g) && effects.AttackedPlayer(g, ev) != uuid.Nil
				},
				Key: shredderCopiesLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, shredderCopiesLabel)
					item.Params.Object, _ = g.PermanentRefForEffect(ev.CardID)
					item.Params.Player = effects.AttackedPlayer(g, ev)
					return item
				},
				Effect: shredderTokenCopies,
			},
			effects.WheneverThisDealsCombatDamageToAPlayer("Shredder — that player loses half their life, rounded up",
				effects.TriggeringTargetLosesHalfTheirLife),
		},
	})
}

const shredderCopiesLabel = "Shredder — for each other opponent, a nonlegendary token copy tapped and attacking them"

// shredderTokenCopies makes the copies: one group per other opponent,
// each attacking that player.
func shredderTokenCopies(g *game.Game, item *game.StackItem) error {
	tmpl, ok := g.TokenCopyTemplateForObjectForEffect(item.Params.Object)
	if !ok {
		return nil
	}
	tmpl.TypeLine = effects.RemoveLegendaryFromTypeLine(tmpl.TypeLine)
	var groups []game.TokenGroup
	for _, opponent := range effects.B27OtherOpponents(g, item.Controller, item.Params.Player) {
		groups = append(groups, game.TokenGroup{Template: effects.CopiedToken(tmpl), Count: 1, Entry: game.TokenEntryOptions{Tapped: true}, Attacking: opponent})
	}
	if len(groups) == 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller, Source: ctx.Source(), Groups: groups, Together: true,
	}, ctx.ItemTail(shredderSacrificeTokens))
}

// shredderSacrificeTokens is "sacrifice those tokens at end of combat":
// a delayed trigger at the beginning of the end of combat step naming the
// tokens that landed. An on-disk key.
var shredderSacrificeTokens = effects.CardTail("shredder-shadow-master/sacrifice-tokens-at-end-of-combat", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return (effects.ScheduleDelayedTrigger{At: game.StepEndCombat, Controller: ctx.Item.Controller,
		Label: "Shredder — sacrifice those tokens", Cards: o.IDs, Body: effects.SacrificeListedCardsBody}).Apply(ctx)
})
