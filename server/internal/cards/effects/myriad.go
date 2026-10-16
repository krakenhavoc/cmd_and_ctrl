package effects

import (
	"slices"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// myriad.go — myriad (CR 702.116) granted by a permanent to the
// creatures it names: Legion Loyalty's "creatures you control have
// myriad" and Cybermen Squadron's "nonlegendary artifact creatures you
// control have myriad".
//
// CR 702.116a: "Whenever this creature attacks, for each opponent other
// than defending player, you may create a token that's a copy of this
// creature that's tapped and attacking that player or a planeswalker
// they control. If one or more tokens are created this way, exile the
// tokens at end of combat."
//
// The keyword is a layer-6 grant on each recipient. Its trigger belongs to
// that attacking creature, so removing the creature's abilities removes
// myriad and each grant contributes a separate instance.
//
// The tokens enter attacking without being declared, so they fire no
// attack trigger and no myriad of their own (CR 508.4).
//
// Each opponent has its own optional decision. After count replacements,
// each resulting token may attack that player or one of their planeswalkers.

// GrantedMyriad declares the recipient's triggered ability. The granting
// card registers it in an AbilityGrant and grants that bundle in layer 6.
func GrantedMyriad(name string) game.TriggeredAbility {
	label := name + " — myriad: token copies attacking each other opponent"
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventAttack},
		AppliesTo: ThisAttacked,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object, _ = g.PermanentRefForEffect(ev.CardID)
			item.Params.Player = b17DefendingPlayer(g, ev)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return myriadTokenCopies(g, item, name)
		},
	}
}

// myriadTokenCopies is the myriad body for one attacking creature: for
// each opponent other than the defending player, a token copy of the
// attacker enters tapped and attacking that player, and the copies are
// exiled at the beginning of the end of combat step through a delayed
// trigger named for the granting card `name`. The attacker is copied
// from its copiable values on the battlefield, or their last-known
// information if it left in response (CR 608.2h, 707.2).
//
// The attacker and the defending player come off the item's Params
// (Object, Player), computed once at trigger time (ADR 0041 P9): the
// defending player is the player the attacker is attacking (CR 508.5).
func myriadTokenCopies(g *game.Game, item *game.StackItem, name string) error {
	attacker, defender := item.Params.Object.ID, item.Params.Player
	if info, ok := g.PermanentForEffect(item.Params.Object); ok && !info.Left {
		if current := g.DefendingPlayerForAttackerForEffect(attacker); current != uuid.Nil {
			defender = current
		}
	}
	tmpl, ok := g.TokenCopyTemplateForObjectForEffect(item.Params.Object)
	if !ok {
		return nil
	}
	return myriadChooseOpponents(NewContext(g, item), name, tmpl, b27OtherOpponents(g, item.Controller, defender), nil)
}

func myriadChooseOpponents(ctx *Context, name string, tmpl game.Card, opponents []uuid.UUID, groups []game.TokenGroup) error {
	if len(opponents) == 0 {
		item := ctx.Item
		return ctx.Game.CreateTokensThenForEffect(game.TokenCreation{
			Controller: ctx.Controller(), Source: ctx.Source(), Groups: groups, Together: true, AttackChoicesLabel: name + " — myriad",
		}, func(g *game.Game, tokens []uuid.UUID) error {
			if len(tokens) == 0 {
				return nil
			}
			return (ScheduleDelayedTrigger{At: game.StepEndCombat, Controller: item.Controller,
				Label: name + " — exile the myriad tokens", Cards: tokens, Body: exileListedCardsBody}).Apply(NewContext(g, item))
		})
	}
	opponent := opponents[0]
	p := ctx.Game.PlayerByIDForEffect(opponent)
	if p == nil || p.Eliminated {
		return myriadChooseOpponents(ctx, name, tmpl, opponents[1:], groups)
	}
	return (MayChoice{Question: name + " — myriad: create a copy attacking " + p.Name + " or a planeswalker they control?",
		YesLabel: "Create copy", NoLabel: "Skip opponent",
		OnYes: func(ctx *Context) error {
			chosen := append(slices.Clone(groups), game.TokenGroup{Template: copiedToken(tmpl), Count: 1, Entry: game.TokenEntryOptions{Tapped: true}, AttackOpponent: opponent})
			return myriadChooseOpponents(ctx, name, tmpl, opponents[1:], chosen)
		},
		OnNo: func(ctx *Context) error { return myriadChooseOpponents(ctx, name, tmpl, opponents[1:], groups) },
	}).Apply(ctx)
}
