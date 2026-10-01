package game

import "github.com/google/uuid"

// rebound.go — rebound (CR 702.88), #1854, ADR 0107 §3.
//
//	CR 702.88a  "Rebound appears on some instants and sorceries. It
//	             represents a static ability that functions while the
//	             spell is on the stack and may create a delayed
//	             triggered ability. 'Rebound' means 'If this spell was
//	             cast from your hand, instead of putting it into your
//	             graveyard as it resolves, exile it and, at the
//	             beginning of your next upkeep, you may cast this card
//	             from exile without paying its mana cost.'"
//	CR 702.88b  that cast follows the rules for alternative costs
//	             (CR 601.2b, 601.2f–h).
//	CR 702.88c  multiple instances of rebound on the same spell are
//	             redundant.
//
// Three pieces, none of them new. Each is another mechanic's, moved
// one step over:
//
//  1. THE EXILE is a fourth case of the destination switch in
//     routeStackCardToGraveyardLocked, beside flashback, buyback and
//     the Adventure exile. Like buyback and the Adventure it replaces
//     only "put it into its owner's graveyard as it resolves" (CR
//     608.2n), so it is gated on that helper's `resolved` flag: a
//     countered or fizzled rebound spell is an ordinary card going to
//     an ordinary graveyard. A copy never reaches the switch (CR
//     707.10: it was not cast, and it ceases to exist off the stack).
//  2. THE DELAYED TRIGGER is CR 603.7's queue with `At: StepUpkeep`
//     and `ControllerTurnOnly` — "at the beginning of YOUR next
//     upkeep" — created on the route's continuation, so it is created
//     only once the card is really in exile. It is data (ADR 0041
//     phase 3): the body is the registered key "rebound/cast" and the
//     card is Params.Object, the exiled object's id and CR 400.7
//     epoch. A table with the trigger waiting is a restore point.
//  3. THE CAST is suspend's: the "you may" is a PendingChoiceMayCast
//     asked as the trigger resolves (CR 603.5), and "yes" stamps a
//     per-object CastPermission over exile priced "{0}" (without
//     paying its mana cost; CR 107.3b locks X at 0) with flash timing,
//     because the cast happens during a resolution in an upkeep and
//     every rebound sorcery would otherwise be uncastable there (CR
//     608.2g).
//
// # Whose rebound it is
//
// "Your hand", "your graveyard" and "your next upkeep" all name the
// spell's CONTROLLER, and CR 603.7d makes the controller of the
// delayed trigger the player who controlled the spell as it resolved.
// A card is only ever in its owner's hand (CR 400.3), so "cast from
// your hand" is CastFromZone == hand and the owner still controlling
// it. A spell another player gained control of (ADR 0104) was not cast
// from its new controller's hand, and goes to its owner's graveyard.
//
// # The window, and the one place it differs from paper
//
// SANDBOX SIMPLIFICATION, cascade's and discover's (ADR 0099 §4): the
// cast happens with an ordinary cast_spell action AFTER the trigger
// finishes resolving rather than inline in its resolution, because an
// inline cast would have to run the whole announce — targets, modes,
// the cost picker — under a paused resolution frame. The window is
// shut as tightly as theirs, so it is never stronger than printed:
//
//   - by the holder PASSING PRIORITY (CastPermission.LapseOnPass):
//     passing is the decline, and the card stays in exile, which is
//     the printed outcome of a rebound card that was not cast;
//   - by CR 400.7 — the card leaving exile (the cast itself included)
//     ends the grant, through the epoch it names;
//   - by the end of the turn at the latest.
//
// Suspend still holds its free cast until the end of the turn; rebound
// does not, because an upkeep cast that could wait for the main phase
// is a choice the printed card does not give.
//
// # Granted rebound
//
// Whether the spell has rebound is HasKeyword at resolution, which
// reads the card's effective abilities when the layer pass has stamped
// any. That is the door ADR 0107 §3 decision 5 (owner decision 3)
// walks through: once the stack step of the layer pass applies layer-6
// keyword grants to a spell, "that spell gains rebound" and "instant
// and sorcery spells you control have rebound" reach this file with no
// change here.

// KeywordRebound is CR 702.88's canonical token — Scryfall's
// "Rebound", lowercased.
const KeywordRebound = "rebound"

// MayCastKeywordRebound is the PendingChoice.MayCastKeyword the
// upkeep offer is asked under.
const MayCastKeywordRebound = "rebound"

// ReboundFreeCastLabel is what the granted cast is called in the log
// and on the wire.
const ReboundFreeCastLabel = "Rebound — cast it without paying its mana cost"

// LapseStaysInExile is rebound's lapse: a granted cast its holder
// passes on leaves the card where it is, in exile, because a rebound
// card that is not cast "stays exiled" (CR 702.88a says nothing more
// happens to it).
const LapseStaysInExile PermissionLapse = "exile"

// spellRebounds reports whether a spell that has just RESOLVED is
// exiled by its rebound (CR 702.88a) instead of going to its owner's
// graveyard.
//
// Four facts, all of them the rule's:
//
//   - the spell has rebound — printed, or (with ADR 0107 PR 4) granted
//     by a layer-6 effect, which HasKeyword reads through the card's
//     effective abilities. One instance or several, the answer is the
//     same (CR 702.88c);
//   - it is not a copy. A copy was not cast (CR 707.10) and never
//     reaches the graveyard route anyway; the guard is here so the
//     function says what it means;
//   - it was cast from a hand (CR 702.88a), which StackItem.CastFromZone
//     records at announce. A rebound card cast from exile by its own
//     rebound, from a graveyard by flashback, or from anywhere else
//     goes to the graveyard;
//   - that hand was its controller's: the card's owner still controls
//     the spell (see the file header).
func spellRebounds(c Card, item *StackItem) bool {
	if item == nil || item.IsCopy || item.CastFromZone != ZoneHand {
		return false
	}
	if item.Controller != uuid.Nil && item.Controller != c.Owner {
		return false
	}
	return HasKeyword(&c, KeywordRebound)
}

// scheduleReboundLocked creates CR 702.88a's delayed triggered ability
// for a rebound card that has just landed in exile: "at the beginning
// of your next upkeep, you may cast this card from exile without
// paying its mana cost".
//
// Called from the exile route's continuation, so it runs once the
// card is really in exile. A card that is not there — a replacement
// sent it elsewhere — creates nothing, because CR 702.88a's trigger is
// part of the same "instead" and there is no exiled card for it to be
// about.
//
// "Next" is free: the upkeep step-entry hook has already run for any
// upkeep this resolution happens in, so a rebound spell cast in its
// controller's own upkeep comes back a full turn later.
//
// Caller must hold g.mu (write).
func (g *Game) scheduleReboundLocked(cardID, controller uuid.UUID) {
	c := exiledCardByIDLocked(g, cardID)
	if c == nil {
		return
	}
	if controller == uuid.Nil {
		controller = c.Owner
	}
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:         controller,
		SourceCardID:       cardID,
		Label:              "Rebound — cast " + c.Name + " from exile",
		At:                 StepUpkeep,
		ControllerTurnOnly: true,
		Body:               reboundCastBody,
		// The exiled OBJECT (CR 400.7): a card that leaves exile and
		// comes back before the upkeep is a new object the trigger is
		// not about (CR 603.7c).
		Params: EffectParams{Object: ObjectRef{ID: cardID, Epoch: c.ObjectEpoch}},
	})
}

// reboundCastBody is "rebound/cast": the upkeep trigger's resolution.
// Registered in init rather than by a var initialiser, because the
// body reaches the cast-permission store and the may_cast prompt, and
// an initialiser would be an initialisation cycle through the
// scheduler.
var reboundCastBody BodyRef

func init() {
	reboundCastBody = DelayedBody("rebound/cast", reboundCast)
}

// reboundCast offers the cast, against the game as it is now (CR
// 608.2). A card that is no longer the object the trigger was created
// for — it left exile, or left and came back — is offered nothing
// (CR 603.7c, 400.7).
func reboundCast(g *Game, item *StackItem, p EffectParams) error {
	c := exiledCardByIDLocked(g, p.Object.ID)
	if c == nil || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	chooser := item.Controller
	if chooser == uuid.Nil {
		chooser = c.Owner
	}
	ref := p.Object
	return g.QueueMayCastPromptForEffect(MayCastPrompt{
		Chooser:      chooser,
		Source:       c.InstanceID,
		Card:         c.InstanceID,
		Question:     "Rebound — cast " + c.Name + " without paying its mana cost?",
		Keyword:      MayCastKeywordRebound,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in exile",
		OnAccept: func(g *Game) error {
			g.grantReboundCastLocked(chooser, ref)
			return nil
		},
	})
}

// grantReboundCastLocked stamps the free cast on the one exiled object
// the trigger was about. The fields that carry the rules:
//
//   - `Cost: "{0}"` is "without paying its mana cost", an alternative
//     cost (CR 702.88b, 118.9), and what locks X at 0 (CR 107.3b).
//     Empty would mean "pay the printed cost".
//   - `Timing: TimingFlash`, CR 608.2g: the cast happens during a
//     resolution in an upkeep, whatever the card's own timing.
//   - `CastOnly`: "cast this card".
//   - `LapseOnPass: LapseStaysInExile`, the window (see the file
//     header). A cast from exile is not a cast from a hand, so the
//     resolved spell goes to its owner's graveyard.
//
// Caller must hold g.mu (write).
func (g *Game) grantReboundCastLocked(player uuid.UUID, ref ObjectRef) {
	c := exiledCardByIDLocked(g, ref.ID)
	if c == nil || c.ObjectEpoch != ref.Epoch {
		return
	}
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:      player,
		Zone:        ZoneExile,
		Cost:        "{0}",
		Timing:      TimingFlash,
		CastOnly:    true,
		Duration:    g.UntilEndOfTurnDuration(),
		LapseOnPass: LapseStaysInExile,
		Source:      c.InstanceID,
		SourceName:  c.Name,
		Label:       ReboundFreeCastLabel,
	}, []Card{*c})
}
