package game

import "github.com/google/uuid"

// casualty.go — CR 702.153, casualty (#3245):
//
//	702.153a Casualty is a keyword that represents two abilities. The
//	         first is a static ability that functions while the spell
//	         with casualty is on the stack. The second is a triggered
//	         ability that functions while the spell with casualty is on
//	         the stack. "Casualty N" means "As an additional cost to
//	         cast this spell, you may sacrifice a creature with power N
//	         or greater," and "When you cast this spell, if a casualty
//	         cost was paid for it, copy it. If the spell has any
//	         targets, you may choose new targets for the copy."
//
// The cost is an ordinary optional cost keyed CasualtyKey (the card's
// own, or one a permanent grants, granted_optional_cost.go), so it is
// offered, validated and paid by the one cast path, with the sacrifice
// made as the spell is cast. The trigger is the ENGINE's, like prowess:
// it belongs to the spell, not to whatever granted the cost, so a
// Silverquill sacrificed to pay its own grant still copies the spell.
// It is read off the paid record on the spell's EventCast (an
// intervening "if", CR 603.4), one copy per casualty cost paid, and
// copies from last-known information so a spell countered in response
// is still copied (storm's reason, CR 608.2h).

// CasualtyKey is the optional-cost key casualty is announced under.
const CasualtyKey = "casualty"

// KeywordCasualty names the trigger.
const KeywordCasualty = "casualty"

// casualtyTrigger is the keyword's "when you cast this spell" half.
var casualtyTrigger = TriggeredAbility{
	Keyword:   KeywordCasualty,
	FromStack: true,
	Watches:   []EventKind{EventCast},
	AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
		return ev.CardID == source.InstanceID && casualtyPaidLocked(g, source) > 0
	},
	Build: func(ev Event, source *Card, _ Characteristic, g *Game) *StackItem {
		item := NewKeyedTriggeredItem(source, source.Name+" — casualty: copy it", casualtyCopyBody,
			EffectParams{Amount: casualtyPaidLocked(g, source)})
		if ev.Actor != uuid.Nil {
			item.Controller, item.Owner = ev.Actor, ev.Actor
		}
		return item
	},
}

// casualtyCopyBody is registered under "casualty/copy" (ADR 0041 P9), so
// a casualty trigger waiting on the stack is a restore point.
var casualtyCopyBody = SimpleDelayedBody("casualty/copy", resolveCasualty)

// casualtyPaidLocked is how many casualty costs were paid for the spell
// `spell`: its own (by index) and any granted one (by key).
func casualtyPaidLocked(g *Game, spell *Card) int {
	if g == nil || spell == nil {
		return 0
	}
	paid := g.StackItemPaidForEffect(spell.InstanceID)
	return OptionalCostTimesPaid(*spell, paid.OptionalCosts, CasualtyKey) + GrantedOptionalTimesPaid(paid, CasualtyKey)
}

// resolveCasualty copies the spell once per casualty cost paid, offering
// new targets for each copy.
func resolveCasualty(g *Game, item *StackItem) error {
	n := item.Params.Amount
	for i := 0; i < n; i++ {
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

// harvestCasualtyLocked fires the casualty trigger for the spell an
// EventCast names, beside its catalog FromStack triggers.
//
// Caller must hold g.mu in write mode.
func (g *Game) harvestCasualtyLocked(pass *harvestPass, card *Card, lki Characteristic) {
	if !casualtyTrigger.AppliesTo(pass.ev, card, lki, g) {
		return
	}
	g.harvestMatchLocked(pass, *card, lki, casualtyTrigger, triggerOfSpell)
}
