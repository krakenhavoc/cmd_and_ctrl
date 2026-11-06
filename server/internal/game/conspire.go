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
// component is #3868's TapOthersCost: two creatures, tapped on the
// cast's teamwork_ids list with the spell on the stack, each sharing a
// colour with the spell (TapOthersCost.SharesColorWithSource). So it is
// offered, validated and paid by the one cast path the replicate cost
// that taps already uses (#4092), and the view and the bot read the same
// candidate walk.
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
// CR 702.78b's multiple instances are not modelled: one conspire cost per
// key (granted_optional_cost.go's rule), so a printed conspire spell cast
// under a grant offers one.

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
		item := NewKeyedTriggeredItem(source, source.Name+" — conspire: copy it", conspireCopyBody,
			EffectParams{Amount: conspirePaidLocked(g, source)})
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

// resolveConspire copies the spell once per conspire cost paid, offering
// new targets for each copy (CR 707.10c).
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
// EventCast names, beside its catalog FromStack triggers.
//
// Caller must hold g.mu in write mode.
func (g *Game) harvestConspireLocked(pass *harvestPass, card *Card, lki Characteristic) {
	if !conspireTrigger.AppliesTo(pass.ev, card, lki, g) {
		return
	}
	g.harvestMatchLocked(pass, *card, lki, conspireTrigger, triggerOfSpell)
}
