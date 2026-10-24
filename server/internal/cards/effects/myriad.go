package effects

import (
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
	return myriadChooseOpponents(NewContext(g, item), name, item.Params.Object, tmpl, b27OtherOpponents(g, item.Controller, defender), nil)
}

// myriadGroups is a token copy attacking each of the opponents chosen.
func myriadGroups(tmpl game.Card, chosen []uuid.UUID) []game.TokenGroup {
	var groups []game.TokenGroup
	for _, opponent := range chosen {
		groups = append(groups, game.TokenGroup{Template: copiedToken(tmpl), Count: 1, Entry: game.TokenEntryOptions{Tapped: true}, AttackOpponent: opponent})
	}
	return groups
}

// myriadChooseOpponents asks, for each opponent still to ask, whether to
// create a copy attacking them. `chosen` are the opponents who have said
// yes; the attacker (obj) rides the prompt so the copy's template is read
// again when the answer arrives.
func myriadChooseOpponents(ctx *Context, name string, obj game.ObjectRef, tmpl game.Card, opponents, chosen []uuid.UUID) error {
	if len(opponents) == 0 {
		item := ctx.Item
		return ctx.Game.CreateTokensThenForEffect(game.TokenCreation{
			Controller: ctx.Controller(), Source: ctx.Source(), Groups: myriadGroups(tmpl, chosen), Together: true, AttackChoicesLabel: name + " — myriad",
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
		return myriadChooseOpponents(ctx, name, obj, tmpl, opponents[1:], chosen)
	}
	return (MayChoice{Question: name + " — myriad: create a copy attacking " + p.Name + " or a planeswalker they control?",
		YesLabel: "Create copy", NoLabel: "Skip opponent",
		Then:    myriadAnswer,
		Carry:   joinCarry(chosen, opponents),
		Word:    name,
		Objects: []game.ObjectRef{obj},
	}).Apply(ctx)
}

// myriadAnswer is one opponent's offer of a token copy: a yes adds them to
// the opponents chosen (the carried IDs before the Nil; the one asked is the
// first after it), either way the next is asked. The attacker is the object the prompt
// carries, and Word is the card granting myriad. Registered in init because
// it asks the next opponent through myriadChooseOpponents, which names it.
// Keyed (#4291).
var myriadAnswer game.ConfirmThen

func init() {
	myriadAnswer = MayAnswer("confirm/myriad-copy", func(ctx *Context, r game.ConfirmAnswer) error {
		if len(r.Objects) == 0 {
			return nil
		}
		obj := r.Objects[0]
		tmpl, ok := ctx.Game.TokenCopyTemplateForObjectForEffect(obj)
		if !ok {
			return nil
		}
		// The carry holds the opponents who said yes, then the one asked and
		// the ones still to ask.
		chosen, queue := splitCarry(r)
		if len(queue) == 0 {
			return nil
		}
		asked, rest := queue[0], queue[1:]
		if r.Accepted {
			chosen = append(chosen, asked)
		}
		return myriadChooseOpponents(ctx, r.Word, obj, tmpl, rest, chosen)
	})
}
