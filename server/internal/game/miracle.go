package game

import (
	"github.com/google/uuid"
)

// miracle.go — miracle (CR 702.94), #1665.
//
//	CR 702.94a  "Miracle is a static ability linked to a triggered
//	             ability. 'Miracle [cost]' means 'You may reveal this
//	             card from your hand as you draw it if it's the first
//	             card you've drawn this turn. When you reveal this card
//	             this way, you may cast it by paying [cost] rather than
//	             its mana cost.'"
//	CR 702.94b  "If a player chooses to reveal a card using its miracle
//	             ability, they play with that card revealed until that
//	             card leaves their hand, that ability resolves, or that
//	             ability otherwise leaves the stack."
//
// Three pieces, and none of them is a new model. It is madness's
// shape (madness.go, ADR 0066's #657 amendment) moved one zone over:
//
//   - THE PRICE is a printed AlternativeCost under the key "miracle"
//     (effects.Miracle), so CR 118.9's machinery — the X rule, the
//     commander tax, cost modifiers, StackItem.AltCost — prices the
//     cast with nothing miracle-specific. What makes it miracle's is
//     one flag, AlternativeCost.RequiresGrant: the offer is claimable
//     only while a live CastPermission for this card OBJECT carries
//     the same key. A Terminus sitting in a hand is not castable for
//     {W}; a Terminus the miracle trigger has just resolved for is.
//
//   - THE TRIGGER watches EventDrawCard from the HAND (#925's zone
//     dimension, which gains ZoneHand here) and applies only to the
//     first card its owner drew this turn — Game.DrawnThisTurn, which
//     actuallyDrawCardLocked appends to before it emits the event, so
//     "the first card" is a fact the harvest reads rather than a count
//     it has to trust. The "you MAY reveal" is the trigger's
//     OptionalPrompt: "Reveal Terminus for its miracle cost {W}?" A
//     "yes" reveals the card to the table (RevealForEffect, CR 701.20)
//     and puts the trigger on the stack; a "no" leaves an ordinary
//     card in an ordinary hand.
//
//   - THE CAST is the trigger's resolution, and it is madness's
//     grant: a per-instance CastPermission{Zone: ZoneHand, AltCostKey:
//     "miracle", Timing: TimingFlash, CastOnly: true} over the one
//     object the trigger was about. The printed offer supplies the
//     price; the permission supplies the right to claim it and CR
//     608.2g's "ignoring timing" — which is what lets a Terminus drawn
//     in an opponent's upkeep be cast there.
//
// # CR 702.94b: the card must still be the card that was revealed
//
// The trigger's item carries the revealed OBJECT (Params.Object, id
// and CR 400.7 epoch). A card that has left the hand since — cast in
// response through some other permission, discarded to a Duress,
// Brainstormed back onto the library — is not that object any more,
// and the resolution grants nothing. A card that left and came BACK is
// a new object with a new epoch, and is refused for the same reason.
//
// # The window, and the one place it is wider than paper
//
// SANDBOX SIMPLIFICATION, madness's and cascade's, for their reason:
// the cast happens with an ordinary cast_spell action AFTER the
// trigger resolves rather than inline in its resolution, because an
// inline cast would have to run the whole announce — targets, modes,
// X, the cost picker — under a paused resolution frame.
//
// Madness bounds its window with an end-step cleanup because its card
// is in exile and would otherwise be stranded. A miracle card is in a
// hand, where an unbounded "cast it for {W} at instant speed" would be
// strictly STRONGER than printed, so its window is shut tighter:
//
//   - by CR 400.7 — the card leaving the hand (the cast itself
//     included) ends the grant, through the epoch it names;
//   - by the holder PASSING PRIORITY (closeMiracleWindowLocked, from
//     PassPriority): in paper the choice is made inside the
//     resolution, and passing is how a player says "no";
//   - by the end of the turn at the latest — the permission's
//     zero Duration, stamped at the grant.
//
// What is left of the widening is CR 117.3b's window: after the
// trigger resolves, the ACTIVE player receives priority first. When the
// miracle's owner is the active player — every draw-step miracle — they
// act first and nothing is wider. When a card was drawn on another
// player's turn, that player may act before the owner casts, and the
// owner may then cast the miracle in response. Madness declares the
// same gap for the same reason.
//
// # The printed cost stays exactly where it was
//
// The permission is scoped to its CLAIM (CastPermission.ForClaim): a
// cast out of a hand that pays the printed cost ignores it entirely,
// so a live miracle grant does not turn a {4}{W}{W} Terminus into an
// instant. The enumerator narrows per offer the same way, so a bot is
// offered the miracle cast at instant speed and the printed one only
// where a sorcery could be cast.

// AltCostKeyMiracle is the alternative-cost key a miracle cast is
// claimed under — the printed offer's (effects.Miracle) and the
// permission's, which must agree for RequiresGrant to open the offer.
const AltCostKeyMiracle = "miracle"

// MiracleTriggerLabel is the stack label and the OncePerBatch key. The
// reveal's log line names the card; the stack overlay shows the source.
const MiracleTriggerLabel = "Miracle — cast it for its miracle cost"

// MiracleTrigger builds CR 702.94a's linked pair for one card: the
// "you may reveal it as you draw it" (the OptionalPrompt) and the
// "when you reveal it this way, you may cast it" (the item). The
// catalog grows it from the card's Miracle alternative cost
// (effects.buildDef), so a card file declares the keyword once.
//
// `name` is only the prompt's copy — "Reveal Terminus for its miracle
// cost {W}?" — and `cost` is only the copy and the log: the price
// itself is the printed offer's.
//
// The trigger is harvested from the HAND, where the card is when the
// draw event fires, with the owner as "you" (CR 108.4; the zone
// harvest hands AppliesTo a source whose Controller is its Owner).
func MiracleTrigger(name, cost string) TriggeredAbility {
	return TriggeredAbility{
		Zones:   []ZoneKind{ZoneHand},
		Watches: []EventKind{EventDrawCard},
		Key:     MiracleTriggerLabel,
		Keyword: AltCostKeyMiracle,
		OptionalPrompt: &TriggerOptionalPrompt{
			Question: "Reveal " + name + " for its miracle cost " + cost + "?",
		},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
			return source != nil && ev.CardID == source.InstanceID &&
				ev.Actor == source.Owner &&
				g.firstCardDrawnThisTurnLocked(ev.Actor) == ev.CardID
		},
		Build: func(_ Event, source *Card, _ Characteristic, g *Game) *StackItem {
			if source == nil {
				return nil
			}
			ref := ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
			// The reveal is the "yes" (CR 702.94a), made of the card as it
			// sits NOW. One that left the hand while the prompt was open
			// is not there to reveal; its trigger still goes on the stack
			// and grants nothing when it resolves (CR 702.94b), exactly as
			// a revealed card Brainstormed away in response would.
			if g.miracleCardInHandLocked(source.Owner, ref) != nil {
				g.RevealForEffect(RevealSpec{
					Player: source.Owner,
					Source: source.InstanceID,
					Reason: "Miracle — " + source.Name,
					Cards:  []uuid.UUID{source.InstanceID},
				})
			}
			return NewKeyedTriggeredItem(source, MiracleTriggerLabel, miracleOfferBody,
				EffectParams{Cost: cost, Object: ref})
		},
	}
}

// miracleOfferBody is "miracle/offer": the trigger's resolution, read
// entirely off the item (ADR 0041 P9) — the revealed object and, for
// the log, the miracle cost.
var miracleOfferBody = DelayedBody("miracle/offer", func(g *Game, item *StackItem, p EffectParams) error {
	g.grantMiracleCastLocked(item.Controller, p.Object)
	return nil
})

// firstCardDrawnThisTurnLocked is the first card `playerID` drew this
// turn, or uuid.Nil. DrawnThisTurn records every draw in order and is
// never pruned when a card leaves the hand — "the first card you drew"
// is a fact about the draw, not about where the card is now.
//
// Caller must hold g.mu.
func (g *Game) firstCardDrawnThisTurnLocked(playerID uuid.UUID) uuid.UUID {
	drawn := g.DrawnThisTurn[playerID]
	if len(drawn) == 0 {
		return uuid.Nil
	}
	return drawn[0]
}

// miracleCardInHandLocked finds the revealed OBJECT in its owner's
// hand: the instance, wearing the epoch it had when it was drawn.
// Nil when it has left (CR 702.94b) or left and come back (CR 400.7).
//
// Caller must hold g.mu.
func (g *Game) miracleCardInHandLocked(owner uuid.UUID, ref ObjectRef) *Card {
	p := g.playerByIDLocked(owner)
	if p == nil || p.Hand == nil {
		return nil
	}
	for i := range p.Hand.Cards {
		c := &p.Hand.Cards[i]
		if c.InstanceID == ref.ID {
			if c.ObjectEpoch != ref.Epoch {
				return nil
			}
			return c
		}
	}
	return nil
}

// grantMiracleCastLocked is CR 702.94a's "you may cast it by paying
// [cost] rather than its mana cost", as a permission over the one
// object the trigger revealed. See the file header for the window and
// ForClaim for why the permission leaves the printed cost alone.
//
// No Cost on the permission: the price is the printed offer's, and a
// permission Cost would reprice every cast of the card that read the
// permission — the printed one included (CastCostFor).
//
// Caller must hold g.mu (write).
func (g *Game) grantMiracleCastLocked(owner uuid.UUID, ref ObjectRef) {
	c := g.miracleCardInHandLocked(owner, ref)
	if c == nil {
		return
	}
	g.GrantCastPermissionForEffect(CastPermission{
		Player:     owner,
		Zone:       ZoneHand,
		Cards:      []PermissionCardRef{{ID: c.InstanceID, Epoch: c.ObjectEpoch}},
		AltCostKey: AltCostKeyMiracle,
		Timing:     TimingFlash,
		CastOnly:   true,
		Source:     c.InstanceID,
		SourceName: c.Name,
		Label:      MiracleTriggerLabel,
	})
}

// closeMiracleWindowLocked drops every miracle permission `playerID`
// holds: they are passing priority, which is the decline (see the file
// header). Called from PassPriority BEFORE the pass runs, because the
// pass that closes one window can be the pass that resolves the next
// miracle trigger and opens another — dropped afterwards, that grant
// would be gone before its owner ever held priority with it.
//
// A no-op for every player with no hand permission, which is every
// player at every pass that does not immediately follow a miracle.
//
// Caller must hold g.mu (write).
func (g *Game) closeMiracleWindowLocked(playerID uuid.UUID) {
	p := g.playerByIDLocked(playerID)
	if p == nil || len(p.CastPermissions) == 0 {
		return
	}
	kept := p.CastPermissions[:0:0]
	for _, perm := range p.CastPermissions {
		if perm.Zone == ZoneHand && perm.AltCostKey == AltCostKeyMiracle {
			continue
		}
		kept = append(kept, perm)
	}
	if len(kept) == len(p.CastPermissions) {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	p.CastPermissions = kept
}
