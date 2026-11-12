package game

import "github.com/google/uuid"

// conspire.go — CR 702.78, conspire (#4690):
//
//	702.78a Conspire is a keyword that represents two abilities. The
//	        first is a static ability that functions while the spell
//	        with conspire is on the stack. The second is a triggered
//	        ability that functions while the spell with conspire is on
//	        the stack. "Conspire" means "As an additional cost to cast
//	        this spell, you may tap two untapped creatures you control
//	        that each share a color with it" and "When you cast this
//	        spell, if its conspire cost was paid, copy it. If the spell
//	        has any targets, you may choose new targets for the copy."
//
// Casualty's shape (casualty.go) with a different cost. The cost is an
// ordinary optional cost keyed ConspireKey — a card's own, or one a
// permanent grants (granted_optional_cost.go, Raiding Schemes) — whose
// component is #3868's TapOthersCost: two creatures, tapped with the
// spell on the stack, each sharing a colour with the spell
// (TapOthersCost.SharesColorWithSource). So it is offered and validated
// by the walk the replicate cost that taps already uses (#4092), and the
// view and the bot read the same candidates.
//
// The two creatures ride a list of their own, CastSpellParams.ConspireIDs
// (#5688), not the teamwork_ids list a spell's own tap cost uses. Conspire
// is an additional cost beside whatever else the spell costs (CR 601.2b
// announces it, CR 601.2f adds it to the total, CR 601.2h pays it), so a
// spell that already taps permanents for its own cost — New Blood,
// teamwork, escalate's taps, a replicate cost that taps — is conspired by
// tapping two MORE creatures: the two lists are validated apart and may
// share no permanent, since a tapped creature isn't an untapped creature
// to tap. castTapLists keeps the announcement #4690 shipped: a cast that
// pays conspire and nothing else on teamwork_ids is read as naming them
// there.
//
// The trigger is the ENGINE's, like casualty's: it belongs to the spell,
// not to whatever granted the cost, so the copy is made even when the
// granting permanent has left by the time the trigger is put on the
// stack. It is read off the paid record on the spell's EventCast (an
// intervening "if", CR 603.4) and copies from last-known information, so
// a spell countered in response is still copied (CR 608.2h). The copy is
// CopyLastKnownSpellForEffect's, so a copy of a permanent spell becomes a
// token as it resolves (CR 707.10f, 608.3f) and the copy may be given
// new targets (CR 707.10c).
//
// CR 702.78b (#5688): "If a spell has multiple instances of conspire, each
// is paid separately and triggers based on its own payment, not any other
// instance of conspire." A grant of conspire is EachSource
// (granted_optional_cost.go), so two permanents granting it, or a grant
// beside a printed conspire, give the spell two costs to announce; each
// one paid taps its own two creatures on the one conspire_ids list and is
// its own trigger, one copy each.

// ConspireKey is the optional-cost key conspire is announced under.
const ConspireKey = "conspire"

// KeywordConspire names the trigger.
const KeywordConspire = "conspire"

// conspireTrigger is the keyword's "when you cast this spell" half.
var conspireTrigger = TriggeredAbility{
	Keyword:   KeywordConspire,
	FromStack: true,
	Watches:   []EventKind{EventCast},
	AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
		return ev.CardID == source.InstanceID && conspirePaidLocked(g, source) > 0
	},
	Build: func(ev Event, source *Card, _ Characteristic, g *Game) *StackItem {
		// One copy: each paid instance is its own trigger (CR 702.78b).
		item := NewKeyedTriggeredItem(source, source.Name+" — conspire: copy it", conspireCopyBody,
			EffectParams{Amount: 1})
		if ev.Actor != uuid.Nil {
			item.Controller, item.Owner = ev.Actor, ev.Actor
		}
		return item
	},
}

// conspireCopyBody is registered under "conspire/copy" (ADR 0041 P9), so
// a conspire trigger waiting on the stack is a restore point. The key is
// an on-disk identity: never renamed, never reused.
var conspireCopyBody = SimpleDelayedBody("conspire/copy", resolveConspire)

// conspirePaidLocked is how many conspire costs were paid for the spell
// `spell`: its own (by index) and a granted one (by key).
func conspirePaidLocked(g *Game, spell *Card) int {
	if g == nil || spell == nil {
		return 0
	}
	paid := g.StackItemPaidForEffect(spell.InstanceID)
	return OptionalCostTimesPaid(*spell, paid.OptionalCosts, ConspireKey) + GrantedOptionalTimesPaid(paid, ConspireKey)
}

// resolveConspire copies the spell, offering new targets for the copy
// (CR 707.10c). Amount is 1 for every trigger built since #5688; a trigger
// restored from before it carries the number of costs paid.
func resolveConspire(g *Game, item *StackItem) error {
	for i := 0; i < item.Params.Amount; i++ {
		err := g.CopyLastKnownSpellForEffect(item.SourceCardID, item.Controller, true, nil)
		if err == ErrCardNotFound {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// harvestConspireLocked fires the conspire trigger for the spell an
// EventCast names, beside its catalog FromStack triggers: once per
// conspire cost paid (CR 702.78b).
//
// Caller must hold g.mu in write mode.
func (g *Game) harvestConspireLocked(pass *harvestPass, card *Card, lki Characteristic) {
	if !conspireTrigger.AppliesTo(pass.ev, card, lki, g) {
		return
	}
	for range conspirePaidLocked(g, card) {
		g.harvestMatchLocked(pass, *card, lki, conspireTrigger, triggerOfSpell)
	}
}

// ConspireCost reports whether `c` is a conspire cost, the optional cost
// whose taps ride conspire_ids rather than teamwork_ids (#5688).
func ConspireCost(c AdditionalCost) bool {
	return c.Key == ConspireKey && !c.TapOthers.Empty()
}

// planConspire is the conspire clause the announced plan pays, or nil.
// Each instance announced is its own plan entry (CR 702.78b), and the
// demand is their sum: two conspire costs tap four creatures, validated
// and tapped together, as planTapOthers sums a replicate cost.
func planConspire(plan []costPayment) *TapOthersCost {
	var found *TapOthersCost
	count := 0
	for _, pay := range plan {
		if !ConspireCost(pay.cost) {
			continue
		}
		if found == nil {
			found = pay.cost.TapOthers
		}
		count += pay.cost.TapOthers.Count
	}
	if found == nil || count == found.Count {
		return found
	}
	sum := *found
	sum.Count = count
	return &sum
}

// castTapLists settles which list names conspire's creatures. They ride
// conspire_ids; but a cast that announces conspire, names none there and
// pays nothing else on teamwork_ids (no tap cost of the spell's own, no
// teamwork, no escalate taps) is the announcement #4690 shipped, whose
// teamwork_ids are conspire's. Nothing else could have read that list, so
// the reading is never ambiguous.
func castTapLists(plan []costPayment, teamwork, conspire []uuid.UUID) (team, cons []uuid.UUID) {
	if planConspire(plan) == nil || len(conspire) > 0 || len(teamwork) == 0 {
		return teamwork, conspire
	}
	if planTapOthers(plan) != nil || planTeamwork(plan) > 0 || planTapCreatures(plan) > 0 {
		return teamwork, conspire
	}
	return nil, teamwork
}

// validateConspireLocked checks the creatures named to pay the plan's
// conspire cost, without tapping anything: the tap clause's own validator
// (two per conspire cost announced, distinct, the caster's, untapped creatures on the
// battlefield, each sharing a colour with the spell), and none of them
// also named to another tap of the same cast (`otherTaps`: the convoke /
// waterbend taps, and the spell's own tap cost). A creature is tapped
// once, so it pays one of the costs, never two. With no conspire cost
// announced any named creature is refused.
//
// Caller must hold g.mu.
func (g *Game) validateConspireLocked(playerID, castID uuid.UUID, plan []costPayment, ids []uuid.UUID, otherTaps ...[]uuid.UUID) error {
	tc := planConspire(plan)
	if tc == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	for _, other := range otherTaps {
		if sharesAnID(ids, other) {
			return ErrInvalidParam
		}
	}
	g.RecomputeLayersIfStaleLocked()
	return g.validateTapOthersCostLocked(playerID, castID, tc, ids)
}

// payConspireLocked taps the validated creatures, one EventTapCard each.
// They are not recorded on PaidCost.TappedOthers: that is what the
// spell's OWN tap cost tapped (Burn at the Stake counts it), and that the
// conspire cost was paid is already on the record by its key.
//
// Call only after validateConspireLocked has passed and with the spell on
// the stack. Caller must hold g.mu.
func (g *Game) payConspireLocked(playerID uuid.UUID, plan []costPayment, ids []uuid.UUID) {
	if planConspire(plan) == nil {
		return
	}
	g.payTapOthersCostLocked(playerID, ids)
}
