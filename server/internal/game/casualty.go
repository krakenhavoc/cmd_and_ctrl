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
		params := EffectParams{Amount: casualtyPaidLocked(g, source)}
		if ref, ok := casualtyXSacrificeLocked(g, source); ok {
			params.Object = ref
		}
		item := NewKeyedTriggeredItem(source, source.Name+" — casualty: copy it", casualtyCopyBody, params)
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

// casualtyXSacrificeLocked is the creature a paid "Casualty X" cost
// sacrificed for the spell `spell` (AdditionalCost.CasualtyX, #5668), as
// the battlefield object it was. False for a spell with no such cost,
// or one whose cost was not paid.
//
// The card's own optional costs hold at most one sacrifice when one of
// them is casualty X, and it has no mandatory sacrifice (effects.Register),
// and granted costs are paid after the card's own (castCostPayments), so
// the first sacrificed object is this cost's.
func casualtyXSacrificeLocked(g *Game, spell *Card) (ObjectRef, bool) {
	if g == nil || spell == nil {
		return ObjectRef{}, false
	}
	paid := g.StackItemPaidForEffect(spell.InstanceID)
	if len(paid.SacrificedObjects) == 0 {
		return ObjectRef{}, false
	}
	costs := OptionalCostsFor(CatalogKey(*spell))
	for _, i := range paid.OptionalCosts {
		if i >= 0 && i < len(costs) && costs[i].Key == CasualtyKey && costs[i].CasualtyX {
			return paid.SacrificedObjects[0], true
		}
	}
	return ObjectRef{}, false
}

// resolveCasualty copies the spell once per casualty cost paid, offering
// new targets for each copy.
//
// A casualty X cost (#5668) is the trigger's Params.Object, the creature
// it sacrificed: X is that creature's power as it last existed on the
// battlefield (CR 608.2h), never less than 0, and the first copy is made
// with the printed exceptions, "The copy isn't legendary and has
// starting loyalty X" (CR 707.9b). The other copies are a granted
// casualty's, which prints no exception.
func resolveCasualty(g *Game, item *StackItem) error {
	n := item.Params.Amount
	for i := 0; i < n; i++ {
		var except func(v *PrintedValues)
		if i == 0 && item.Params.Object.ID != uuid.Nil {
			x := 0
			if info, ok := g.PermanentForEffect(item.Params.Object); ok && info.Power > 0 {
				x = info.Power
			}
			except = func(v *PrintedValues) {
				v.RemoveSupertype("Legendary")
				v.SetStartingLoyalty(x)
			}
		}
		err := g.CopyLastKnownSpellForEffect(item.SourceCardID, item.Controller, true, except)
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
