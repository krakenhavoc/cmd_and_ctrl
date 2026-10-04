package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// The granted keyword is written as the trigger itself, hung on the
// granting permanent and fired once per creature its controller
// declares as an attacker that has myriad from it. Two granting
// permanents give a creature two instances, and each triggers
// (CR 702.116b), which is what one trigger per granting permanent does.
//
// The tokens enter attacking without being declared, so they fire no
// attack trigger and no myriad of their own (CR 508.4).
//
// Two simplifications every granted-myriad card shares, both weaker
// than printed: one yes/no covers every other opponent rather than one
// per opponent, and the copies always attack the player, never a
// planeswalker that player controls.

// GrantedMyriad is the myriad trigger a permanent named `name` grants to
// the attacking creatures `has` accepts. `has` reads the attacker as it
// is when the attack is declared.
func GrantedMyriad(name string, has func(attacker game.Card) bool) game.TriggeredAbility {
	label := name + " — myriad: token copies attacking each other opponent"
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if !b27AttackedAndAnotherOpponentRemains(ev, source, g) {
				return false
			}
			attacker, ok := g.LookupCardForEffect(ev.CardID)
			return ok && has(attacker)
		},
		OptionalPrompt: &game.TriggerOptionalPrompt{Question: name + " — myriad: create token copies of the attacker tapped and attacking each other opponent?"},
		Key:            label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object = game.ObjectRef{ID: ev.CardID}
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
// wherever it now sits: a creature that died in response is copied
// from the graveyard, its last printed values (CR 702.116a makes the
// copies regardless).
//
// The attacker and the defending player come off the item's Params
// (Object, Player), computed once at trigger time (ADR 0041 P9): the
// defending player is the player the attacker is attacking (CR 508.5).
func myriadTokenCopies(g *game.Game, item *game.StackItem, name string) error {
	attacker, defender := item.Params.Object.ID, item.Params.Player
	ctx := NewContext(g, item)
	tmpl, ok := TokenCopyTemplate(g, attacker)
	if !ok {
		return nil
	}
	tmpl.Tapped = true
	cursor := b25LastEventSeq(g)
	for _, opp := range b27OtherOpponents(g, item.Controller, defender) {
		if err := g.CreateTokensAttackingForEffect(item.Controller, tmpl, 1, opp); err != nil {
			return err
		}
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEndCombat,
		Label: name + " — exile the myriad tokens",
		Cards: tokens,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
}
