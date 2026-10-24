package game

import (
	"fmt"

	"github.com/google/uuid"
)

// madness.go — madness (CR 702.35), #657.
//
// Madness is two abilities that a card has while it is in a HAND, and
// the whole of the keyword is those two plus the price:
//
//	CR 702.35a  "If a player would discard this card, that player
//	             discards it, but exiles it instead of putting it
//	             into their graveyard" — a replacement over the ONE
//	             discard event (#650's RepEventDiscard).
//	            "When this card is exiled this way, its owner may cast
//	             it by paying [cost] rather than paying its mana
//	             cost" — a triggered ability that functions from
//	             exile (#925's TriggeredAbility.Zones).
//	CR 702.35b  "If that player doesn't, they put this card into
//	             their graveyard."
//
// Nothing here is a new model. The replacement is ADR 0013's, the
// trigger is #925's zone dimension, the offer is the
// resolution_cast prompt cascade and suspend also use (#4250), and the cast
// is ADR 0066's per-instance CastPermission with a key and
// TimingFlash — the timing the ADR reserved for this keyword by name.
// What this file is, is the wiring, declared ONCE: a card file says
// `Madness: "{R}"` and buildDef grows both abilities from it, so two
// madness cards cannot spell the keyword two ways and the third
// cannot forget half of it.
//
// # It does not care WHY the card was discarded
//
// CR 702.35a says "if a player would discard this card", and the
// Fiery Temper ruling (2022-12-08) spells out the consequence: a
// discard as a COST, a discard an effect instructed, and the cleanup
// step's hand-size discard all work. So the replacement reads no
// cause at all — the one thing #650 gave the event that madness does
// not want — and it is the discard event's existence, rather than any
// field on it, that madness keys on. Library of Leng is the contrast
// next door: it replaces DiscardCauseEffect only.
//
// A COST discard settles without asking (CR 601.2h, zoneRoute.
// MustSettleNow), which costs madness nothing, because the
// replacement is not Optional: it is applied rather than offered.
//
// # It fires for a DISCARD and for nothing else
//
// A madness card milled, tutored, sacrificed off the battlefield or
// pitched to Force of Will's alternative cost is not discarded, and
// none of those opens a RepEventDiscard. CR 701.9a defines a discard
// by the move OUT of the hand, so the one event kind is the whole
// test.
//
// # The exile is FACE UP, and carries no marker
//
// CR 702.35a exiles the card plainly. #94's old body said "face down
// with a marker" and was wrong; the trigger needs no marker because
// it reads the discard event and the card's own presence in exile.
//
// LIMITATION, DECLARED, and the only place this is looser than paper:
// if ANOTHER replacement was the one that exiled the discarded card —
// Rest in Peace, Leyline of the Void, both of which also turn "into a
// graveyard" into "into exile" — the trigger cannot tell, so it
// offers the madness cast anyway. Telling them apart needs a record
// of WHICH replacement applied, which the engine does not keep past
// the event. It is unreachable by a player acting in their own
// interest: with both applicable the discarding player is the one
// CR 616.1 asks, madness is strictly better for them, and choosing
// the other one produces the identical board state under this
// engine — so the looseness costs nobody a decision. #657 records it.
//
// # "If that player doesn't, they put it into their graveyard"
//
// The cast happens during the trigger's resolution (CR 608.2g; ADR
// 0066's 2026-10-23 amendment, #4250). The trigger stamps the madness
// cast on the card (a per-instance CastPermission at the madness cost,
// with flash timing, CR 608.2g's "ignoring timing restrictions" as this
// engine spells it) and raises a resolution_cast prompt over it, which
// holds the resolution open while the owner casts the card, with its
// targets, modes and the madness cost paid then, or declines. Nobody
// receives priority in between.
//
// Declining is CR 702.35a's "if that player doesn't": the permission
// lapses with LapseToGraveyard and the card goes into its owner's
// graveyard then and there, inside the trigger's resolution, which is
// the printed outcome. The prompt is plain data, so a table paused on it
// is a restore point.
//
// Outside an open resolution (nothing reaches one: the trigger always
// resolves) the permission alone is the window: it closes on the
// owner's next priority pass with the card put into the graveyard, and
// the end-step cleanup (scheduleMadnessGraveyardLocked) is the backstop
// that keeps an uncast card from being parked in exile forever.

// AltCostKeyMadness is the alternative-cost key a madness cast is
// claimed under. Shared with the printed-keyword keys on purpose
// (ADR 0066 Decision 3): StackItem.AltCost is what a card reads back,
// so CR 702.35b's "if its madness cost was paid" (Avacyn's Judgment,
// #653) will work however the permission arrived.
const AltCostKeyMadness = "madness"

// CatalogMadnessLife is the catalog hook for the "Pay N life" some
// madness costs add to their mana (Shadowgrange Archfiend's "Madness—
// {2}{B}, Pay 8 life"), wired from the CardDef like its neighbours. Nil
// (no catalog) means no card declares one.
var CatalogMadnessLife func(oracleID string) int

// MadnessLifeFor is the life a card's madness cost charges beside its
// mana, or zero.
func MadnessLifeFor(oracleID string) int {
	if CatalogMadnessLife == nil || oracleID == "" {
		return 0
	}
	return CatalogMadnessLife(oracleID)
}

// MadnessReplacementLabel and MadnessTriggerLabel are the CR 616
// prompt header and the stack label. Built once so the log, the
// ordering prompt and the stack overlay cannot disagree.
const (
	MadnessReplacementLabel = "Madness — exile it instead of putting it into your graveyard"
	MadnessTriggerLabel     = "Madness — cast it for its madness cost"
)

// MadnessReplacement builds the CR 702.35a replacement every madness
// card carries while it is in a hand. The catalog attaches it from
// the card's `Madness` declaration rather than the card file writing
// it out (effects.buildDef).
//
// It is found by the SELF-replacement pass in
// gatherActiveReplacementsLocked — the one that consults the card the
// event is about when that card is not on the battlefield — which is
// what lets a card in a hand replace its own exit. The AppliesTo
// still compares ev.CardID to the source's instance, because a second
// copy of the same card sitting on the BATTLEFIELD (Big Game Hunter
// is a creature) is walked by the ordinary catalog pass and must not
// replace a discard of the copy in hand.
//
// `ev.NewZone == ZoneGraveyard` is the CR 614.5 half of that: once
// any replacement has sent the card somewhere else, madness does not
// apply to it any more — which is what keeps madness and Library of
// Leng from both rewriting one discard, and what stops this from
// re-applying to its own output in the CR 616.1 loop.
func MadnessReplacement() ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventDiscardCard},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
			return ev.Kind == RepEventDiscard &&
				src != nil && ev.CardID == src.InstanceID &&
				ev.NewZone == ZoneGraveyard
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			// Exile is a shared zone, so it takes no owner — the
			// route ignores NewZoneOwner for it, and clearing the
			// field keeps the settled event honest for anything that
			// reads it afterwards.
			ev.NewZone = ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return nil
		},
		// CR 616.1: the affected object's controller chooses the
		// order when two replacements apply, and a card in a hand has
		// no controller — its owner, who is the player discarding it.
		Controller: func(ev *ReplacementEvent, _ *Game, _ *Card) uuid.UUID {
			return ev.DiscardPlayer
		},
		Label: MadnessReplacementLabel,
	}
}

// MadnessTrigger builds CR 702.35a's second half: "when this card is
// exiled this way, its owner may cast it by paying its madness cost".
//
// It watches EventDiscardCard from ZoneExile, which is the reflexive
// "when you do" written in the vocabulary the engine has: the discard
// event is emitted once the move has LANDED (CR 701.9a defines a
// discard by the move out of the hand, so it fires wherever the card
// ends up), so by the time the harvest runs the card is sitting in
// exile and the zone walk finds it there.
//
// `ev.NewZone == ZoneExile` is what makes it a MADNESS discard rather
// than any other: a discard that reached a graveyard, a library
// (Library of Leng) or the command zone (CR 903.9) is not one. See
// the file header for the one case this cannot distinguish.
//
// "You" is the card's OWNER (CR 108.4), which the zone harvest
// already supplies by handing the predicate a source whose Controller
// is its Owner.
func MadnessTrigger(cost string) TriggeredAbility {
	return TriggeredAbility{
		Zones:   []ZoneKind{ZoneExile},
		Watches: []EventKind{EventDiscardCard},
		Key:     MadnessTriggerLabel,
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return source != nil && ev.CardID == source.InstanceID &&
				ev.OldZone == ZoneHand && ev.NewZone == ZoneExile
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			// ADR 0041 P9 (#1497, tier 4): madness has no catalog row,
			// so this item is keyed directly, with the printed madness
			// cost carried as Params.Cost rather than captured.
			return NewKeyedTriggeredItem(source, MadnessTriggerLabel, madnessOfferBody, EffectParams{Cost: cost})
		},
	}
}

// madnessOfferBody is "madness/offer": the CR 702.35a "may", reading
// the madness cost back off Params rather than a captured variable
// (ADR 0041 P9, #1497, tier 4).
var madnessOfferBody = DelayedBody("madness/offer", func(g *Game, item *StackItem, p EffectParams) error {
	return g.offerMadnessCastLocked(item.Controller, item.SourceCardID, p.Cost)
})

// offerMadnessCastLocked is the trigger's resolution: the CR 702.35a
// "may", cast during this resolution, and its "if that player doesn't".
//
// A card that has left exile since the trigger went on the stack is
// offered nothing and moved nowhere — CR 608.2 resolves against the
// game as it is, and a madness card answered by a Bojuka Bog is
// somebody else's business now.
//
// Caller must hold g.mu (write).
func (g *Game) offerMadnessCastLocked(owner, cardID uuid.UUID, cost string) error {
	c := exiledCardByIDLocked(g, cardID)
	if c == nil {
		return nil
	}
	if !g.playerInGameLocked(owner) {
		// An owner who has left cannot cast it: "if that player
		// doesn't" is already true.
		return g.madnessToGraveyardLocked(cardID)
	}
	name := c.Name
	priced := cost
	if life := MadnessLifeFor(CatalogKey(*c)); life > 0 {
		priced = fmt.Sprintf("%s and %d life", cost, life)
	}
	g.grantMadnessCastLocked(owner, cardID, cost)
	if g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      owner,
		Source:       cardID,
		Reason:       "Madness",
		Candidates:   []uuid.UUID{cardID},
		Keyword:      MayCastKeywordMadness,
		AcceptLabel:  "Cast it for " + priced,
		DeclineLabel: "Put it into your graveyard",
	}) == uuid.Nil {
		// Outside a resolution: the permission is the window, with the
		// end-step backstop.
		g.scheduleMadnessGraveyardLocked(owner, cardID, name)
	}
	return nil
}

// grantMadnessCastLocked stamps the cast on the one exiled object the
// trigger was about.
//
// Five fields carry the rules:
//
//   - `Cost` is the MADNESS cost, paid "rather than paying its mana
//     cost" (CR 702.35a). A real alternative price, so CR 107.3b
//     locks X at 0 through the same CastCostFor rule suspend's {0}
//     rides.
//
//   - `AltCostKey` is what makes the resulting spell readable as "its
//     madness cost was paid" (CR 702.35b).
//
//   - `Timing: TimingFlash`, because the cast happens during the
//     resolution of a trigger (CR 608.2g) and the trigger can resolve
//     in anyone's turn at any speed. ADR 0066 reserved this value for
//     madness by name.
//
//   - `CastOnly`, because CR 702.35a says "cast it": no printed
//     madness card is a land, and the flag is the rule rather than a
//     guard against one.
//
//   - `LapseOnPass: LapseToGraveyard`, CR 702.35a's "if that player
//     doesn't, they put this card into their graveyard": what the
//     resolution_cast prompt's decline does (#4250), and what a pass
//     does outside a resolution.
//
// The zero `Duration` is #945's "until end of turn, stamped at the
// grant", a backstop only: the prompt's answer ends the window first.
//
// Caller must hold g.mu (write).
func (g *Game) grantMadnessCastLocked(owner, cardID uuid.UUID, cost string) {
	c := exiledCardByIDLocked(g, cardID)
	if c == nil {
		return
	}
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:     owner,
		Zone:       ZoneExile,
		AltCostKey: AltCostKeyMadness,
		Cost:       cost,
		Timing:     TimingFlash,
		CastOnly:   true,
		// CR 702.35a: not cast means the graveyard (#4250).
		LapseOnPass: LapseToGraveyard,
		Label:       MadnessTriggerLabel,
	}, []Card{*c})
}

// scheduleMadnessGraveyardLocked queues the "if you said yes and then
// did not cast it, it goes to the graveyard anyway" cleanup: at the
// beginning of the next end step, a madness card still sitting in
// exile under its madness grant is put into its owner's graveyard.
//
// This is what keeps the grant from being STRONGER than printed. It is
// the shape cascade used before ADR 0099 moved cascade onto a
// pass-closed window (CastPermission.LapseOnPass); madness keeps the
// end-step cleanup.
//
// Deliberately narrow, and narrowed by CR 400.7 rather than by the
// grant: it fires only if the exiled card is still the SAME OBJECT the
// offer was made about, which `Card.ObjectEpoch` answers — the same
// integer a per-instance permission names (ADR 0066 Decision 3). A
// card that was cast, or that left exile and came back by any road, is
// a new object and is left alone.
//
// The epoch rather than the grant, which is what cascade's twin reads.
// The two agree today and the epoch is the one that says what is
// actually meant — "the object this offer was about" — without tying
// the reclamation to when the permission sweep happens to run. That
// matters because the gap can be long: a madness discard in a turn's
// END step or its cleanup (the CR 514.1 hand-size trim is exactly
// that) schedules this for the NEXT turn's end step, a full turn
// after the window the grant was written for.
//
// Caller must hold g.mu (write).
func (g *Game) scheduleMadnessGraveyardLocked(owner, cardID uuid.UUID, name string) {
	c := exiledCardByIDLocked(g, cardID)
	if c == nil {
		return
	}
	epoch := c.ObjectEpoch
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   owner,
		SourceCardID: cardID,
		Label:        "Madness — put " + name + " into its owner's graveyard",
		At:           StepEnd,
		Cards:        []uuid.UUID{cardID},
		Body:         madnessGraveyardBody,
		// The epoch is the object the offer was about (CR 400.7).
		Params: EffectParams{Object: ObjectRef{ID: cardID, Epoch: epoch}},
	})
}

// madnessGraveyardBody is the delayed trigger's body, a registered key
// (ADR 0041 phase 3, #1497): the epoch it used to capture is its
// Params.Object.
// madnessGraveyardBody is assigned in init: a var initialiser would be
// an initialisation cycle through the exit primitives.
var madnessGraveyardBody BodyRef

func init() {
	madnessGraveyardBody = DelayedBody("madness/graveyard-if-not-cast", func(g *Game, _ *StackItem, p EffectParams) error {
		c := exiledCardByIDLocked(g, p.Object.ID)
		if c == nil || c.ObjectEpoch != p.Object.Epoch {
			return nil
		}
		return g.madnessToGraveyardLocked(p.Object.ID)
	})
}

// madnessToGraveyardLocked is CR 702.35b's move: the exiled card goes
// to its OWNER's graveyard, through the shared exit primitive so the
// CR 614 window runs over it and a commander is offered the command
// zone (CR 903.9).
//
// It is NOT a discard: the card left the hand a while ago, and
// CR 701.9a's keyword action is the move out of a hand. So no discard
// payoff fires a second time for one discarded card.
//
// Caller must hold g.mu (write).
func (g *Game) madnessToGraveyardLocked(cardID uuid.UUID) error {
	if exiledCardByIDLocked(g, cardID) == nil {
		return nil
	}
	return g.putIntoOwnersGraveyardLocked(cardID)
}

// putIntoOwnersGraveyardLocked moves a card, from whatever zone holds
// it, into its owner's graveyard through the shared exit primitive (the
// CR 614 window, CR 903.9's command-zone offer). Not a discard. A card
// that is nowhere is left alone. Caller must hold g.mu (write).
func (g *Game) putIntoOwnersGraveyardLocked(cardID uuid.UUID) error {
	c, _ := g.findCardAndZoneLocked(cardID)
	if c == nil {
		return nil
	}
	owner := c.Owner
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:   cardID,
		Dst:      ZoneGraveyard,
		DstOwner: owner,
		Actor:    owner,
	})
	return err
}
