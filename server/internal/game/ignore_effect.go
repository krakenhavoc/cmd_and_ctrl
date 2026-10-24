package game

import (
	"slices"

	"github.com/google/uuid"
)

// ignore_effect.go — CR 116.2d: "Some effects from static abilities
// allow a player to take an action to ignore the effect from that
// ability for a duration. Doing so is a special action. A player can
// take such an action any time they have priority." Issue #1986, ADR
// 0109 §4's amendment of 2026-10-23.
//
// The card is Damping Engine: "A player who controls more permanents than
// each other player can't play lands or cast artifact, creature, or
// enchantment spells. That player may sacrifice a permanent of their
// choice for that player to ignore this effect until end of turn." Both
// halves of the first sentence already had a gate (CastRestriction,
// LandPlayRestriction). What was missing was the escape: a record, per
// player and per source, that the gates read.
//
// THREE PARTS, one file:
//
//  1. THE DECLARATION. CardDef.IgnorableEffect says a permanent's static
//     ability can be ignored, by whom (Applies) and what it says on the
//     row (Label). The restrictions that ARE that ability carry
//     Ignorable: true; the card's other restrictions, if it had any,
//     would not be switched off by it.
//  2. THE RECORD. A ModIgnoreEffect ScopedEffect naming the player, whose
//     Source is the permanent as an object (id and epoch, CR 400.7), so a
//     second Engine, or the same Engine blinked back, is a different
//     effect and is not ignored. It lasts until end of turn and is swept
//     at cleanup. It is a ScopedEffect and not a field on the player for
//     ModCantPlayLands' reason: a binary from before this kind refuses a
//     restore point naming it, where it would read an unknown field as no
//     ignore at all and lock the player out again.
//  3. THE ACTION. special_action {kind: "ignore_effect", card_id: the
//     permanent, sacrifice_id: the cost}. It is the CR 116.2 verb with one
//     difference from every other kind: its card is on the BATTLEFIELD but
//     the actor is not its controller, but whoever the effect bears on. So
//     it is not offered through the card (SpecialActionsOfferedByCard) but
//     through IgnoreEffectOffersLocked, which the engine, the legal-move
//     enumerator and the view all ask.
//
// The cost is a sacrifice of a permanent the player controls (CR 701.21a),
// of their choice and so named on the wire. Nothing else in the cost: no
// mana, no tap. It can be paid at any priority, under split second (a
// special action is neither a cast nor an activation, CR 702.61b), on any
// player's turn, with anything on the stack.

// SpecialActionIgnoreEffect is CR 116.2d's special action: sacrifice a
// permanent to ignore a static ability's effect for a duration.
const SpecialActionIgnoreEffect SpecialActionKind = "ignore_effect"

// IgnorableEffect declares that a permanent's static ability may be
// ignored by a player it bears on, for the price of sacrificing a
// permanent. Pure data besides Applies, as the restrictions are.
type IgnorableEffect struct {
	// Label is the row as the player reads it ("Sacrifice a permanent:
	// ignore Damping Engine's effect until end of turn").
	Label string
	// Applies reports whether the ability bears on `player` right now.
	// An ability that does not bear on a player has nothing for them to
	// ignore, so nothing is offered. Read-only, under g.mu.
	Applies func(g *Game, source *Card, player uuid.UUID) bool
}

// CatalogIgnorableEffect is the catalog hook the effects package wires
// at init, mirroring CatalogCastRestrictions. Nil, or a nil return, means
// the card's abilities cannot be ignored, which is every card but one.
var CatalogIgnorableEffect func(oracleID string) *IgnorableEffect

// IgnoreEffectOffer is one ignore a player may take right now.
type IgnoreEffectOffer struct {
	// Source is the permanent whose effect it is.
	Source uuid.UUID
	// SourceName is its name, for the row and the log.
	SourceName string
	// Label is IgnorableEffect.Label.
	Label string
	// Candidates are the permanents the player may sacrifice to pay for
	// it: those they control that can be sacrificed (CR 701.21a, and
	// "can't be sacrificed" stops one). In payment order
	// (SacrificePaymentOrderForEffect), so the first is the one a bot or
	// a "choose for me" button gives up. Never empty: an offer the player
	// cannot pay is not an offer.
	Candidates []uuid.UUID
}

// ignorableEffectOf is the declaration the permanent carries right now:
// none under an ability-removing effect (CR 613.1f), which is
// catalogAbilityKeyOf's whole job.
func ignorableEffectOf(c *Card) *IgnorableEffect {
	if CatalogIgnorableEffect == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return CatalogIgnorableEffect(key)
}

// effectIgnoredLocked reports whether `player` is ignoring the effect of
// the permanent `src` right now: a live ModIgnoreEffect record names
// them and names this object (id and epoch). Asked by the cast gate and
// the land-play gate for every restriction marked Ignorable, so it is
// one scan of the stored records, which is empty in nearly every game.
//
// The duration is asked here rather than left to the sweep, for the
// reason handPlayBanLocked gives.
//
// Caller must hold g.mu (read or write).
func (g *Game) effectIgnoredLocked(player uuid.UUID, src *Card) bool {
	if player == uuid.Nil {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Source.ID != src.InstanceID || e.Source.Epoch != src.ObjectEpoch {
			continue
		}
		for _, m := range e.Mods {
			if m.Kind != ModIgnoreEffect || m.Player != player {
				continue
			}
			if g.durationExpiredLocked(e.Duration, false) {
				continue
			}
			return true
		}
	}
	return false
}

// ignoreEffectCandidatesLocked are the permanents `player` may sacrifice
// to pay for an ignore, in payment order.
//
// Caller must hold g.mu (read or write).
func (g *Game) ignoreEffectCandidatesLocked(player, sourceID uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player || g.permanentCantBeSacrificedLocked(c) {
			continue
		}
		ids = append(ids, c.InstanceID)
	}
	return g.SacrificePaymentOrderForEffect(ids, sourceID)
}

// ignoreEffectOfferLocked is the offer `src` makes to `player`, or nil:
// the permanent declares an ignorable effect, it bears on the player,
// they are not already ignoring it and they have a permanent to give up.
//
// Caller must hold g.mu (read or write).
func (g *Game) ignoreEffectOfferLocked(player uuid.UUID, src *Card) *IgnoreEffectOffer {
	ig := ignorableEffectOf(src)
	if ig == nil || ig.Applies == nil || !ig.Applies(g, src, player) {
		return nil
	}
	if g.effectIgnoredLocked(player, src) {
		return nil
	}
	cands := g.ignoreEffectCandidatesLocked(player, src.InstanceID)
	if len(cands) == 0 {
		return nil
	}
	return &IgnoreEffectOffer{
		Source:     src.InstanceID,
		SourceName: src.Name,
		Label:      ig.Label,
		Candidates: cands,
	}
}

// IgnoreEffectOffersLocked is every CR 116.2d ignore `player` may take
// right now, in battlefield order. THE accessor: the engine's check, the
// legal-move enumerator and the view all ask here, so none of them can
// disagree about what is on offer. A seat that is out of the game is
// offered nothing. Nil for nearly every player in nearly every game: it
// is one battlefield walk whose only cost per permanent is the catalog
// lookup.
//
// Caller must hold g.mu (read or write).
func (g *Game) IgnoreEffectOffersLocked(player uuid.UUID) []IgnoreEffectOffer {
	if CatalogIgnorableEffect == nil || g.Battlefield == nil {
		return nil
	}
	if p := g.playerByIDLocked(player); p == nil || p.Eliminated {
		return nil
	}
	var out []IgnoreEffectOffer
	for i := range g.Battlefield.Cards {
		if offer := g.ignoreEffectOfferLocked(player, &g.Battlefield.Cards[i]); offer != nil {
			out = append(out, *offer)
		}
	}
	return out
}

// performIgnoreEffectLocked is the ignore_effect performer: validate the
// offer and the cost, sacrifice the permanent, record the ignore.
//
// It refuses before paying anything, and the sacrifice is the last thing
// that can fail, so a refused action leaves the board as it was. The
// action uses no stack and grants no window (CR 116.2, 116.3): the
// caller keeps priority.
//
// Caller must hold g.mu (write) and have run the dispatcher's priority
// check.
func (g *Game) performIgnoreEffectLocked(p *Player, sourceID uuid.UUID, params SpecialActionParams) error {
	// The other kinds' pickers mean nothing here; a client that sent one
	// believed it was taking another action.
	if params.Cost != "" || params.Door != DoorNone || params.Route != "" {
		return ErrInvalidParam
	}
	src := findBattlefieldCard(g, sourceID)
	if src == nil {
		return ErrCardNotFound
	}
	offer := g.ignoreEffectOfferLocked(p.ID, src)
	if offer == nil {
		return ErrSpecialActionNotOffered
	}
	if params.SacrificeID == uuid.Nil || !slices.Contains(offer.Candidates, params.SacrificeID) {
		return ErrInvalidParam
	}
	epoch := src.ObjectEpoch
	// Its own CR 116.2 event, as every special action is (#1341).
	g.beginEventBatchLocked()
	// The cost: the sacrifice (CR 701.21a), settled now (a cost is paid
	// before the action takes effect).
	if err := g.payCostSacrificesLocked([]uuid.UUID{params.SacrificeID}, nil); err != nil {
		return err
	}
	// The effect: ignore until end of turn. A permanent that has itself
	// gone (the player gave up the Engine) has nothing left to ignore,
	// and a record naming a dead object would be a stored fact that
	// matches nothing, so none is written.
	if now := findBattlefieldCard(g, sourceID); now != nil && now.ObjectEpoch == epoch {
		g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
			[]Mod{{Kind: ModIgnoreEffect, Player: p.ID}}, g.UntilEndOfTurnDuration(),
			"ignores "+offer.SourceName)
	}
	g.EmitEvent(Event{
		Kind:   EventSpecialAction,
		Actor:  p.ID,
		CardID: sourceID,
		Label:  offer.Label,
	})
	g.noteActionTakenLocked(p.ID)
	g.runStateChecksLocked()
	return nil
}

// ControlsMoreThanEachOtherPlayerForEffect reports whether `player`
// controls strictly more permanents than each other player still in the
// game does — Damping Engine's "a player who controls more permanents
// than each other player". Tokens are permanents and count. A player who
// has left the game (CR 800.4a) is not a player any more. With no other
// player left the comparison holds vacuously, as the words do.
//
// Caller must hold g.mu (read or write).
func (g *Game) ControlsMoreThanEachOtherPlayerForEffect(player uuid.UUID) bool {
	if g.Battlefield == nil || player == uuid.Nil {
		return false
	}
	counts := make(map[uuid.UUID]int, len(g.Seats))
	for i := range g.Battlefield.Cards {
		counts[g.Battlefield.Cards[i].Controller]++
	}
	mine := counts[player]
	for _, p := range g.Seats {
		if p == nil || p.ID == player || p.Eliminated {
			continue
		}
		if counts[p.ID] >= mine {
			return false
		}
	}
	return true
}
