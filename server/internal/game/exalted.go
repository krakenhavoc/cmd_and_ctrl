package game

// exalted.go — CR 702.83, the fourth KEYWORD TRIGGER the engine derives
// from an object's ability list rather than from its catalog entry
// (#2538, ADR 0101 amendment 2026-10-08). Prowess (prowess.go) is the
// model.
//
//	702.83a Exalted is a triggered ability. "Exalted" means "Whenever a
//	        creature you control attacks alone, that creature gets
//	        +1/+1 until end of turn."
//	702.83b A creature "attacks alone" if it's the only creature
//	        declared as an attacker in a given combat phase. See rule
//	        506.5.
//
// WHY A TOKEN. Exalted used to be a catalog constructor
// (effects.Exalted) on each printed source. That reached the two
// Hierarchs and nothing else: not the cards whose only text is
// keywords (the deck importer dropped "Exalted"), not Sublime
// Archangel's "other creatures you control have exalted", and not an
// exalted counter (CR 122.1b). A token in Characteristic.Abilities
// reaches all of them, the way prowess and evolve do.
//
// MULTIPLE INSTANCES. CR 702.83 has no rule of its own on the point,
// so CR 113.2c applies: "If an object has multiple instances of the
// same ability, each instance functions independently." Exalted is
// therefore CUMULATIVE (KeywordIsCumulative), and each instance is its
// own trigger. That includes exalted counters: the Emissary of
// Soulfire ruling of 2024-06-07 says "A creature with multiple exalted
// counters will have that many instances of exalted", so the counter
// effect appends one instance per counter (keyword_counters.go).
//
// THE TRIGGER. "Whenever a creature you control attacks alone" watches
// EventAttack, which is emitted only for a declared attacker (CR
// 508.3a), so a creature put onto the battlefield attacking never
// triggers it. "Alone" is the whole declaration (CR 506.5): exactly one
// creature declared as an attacker, whoever controls it.
// DeclareAttackers stamps every attacker before it announces any
// EventAttack, so a lone declaration reads as alone (AttackedAlone).
//
// THE CONTROLLER. The trigger is controlled by its source's controller
// when it triggered (CR 603.3a); NewTriggeredItem takes it from the
// source. Exalted on a permanent you control triggers only for your
// own lone attacker.
//
// THE PUMP. The attacker is fixed as the ability triggers, by instance
// ID and object epoch (Params.Object). An attacker that leaves the
// battlefield in response, or leaves and comes back as a new object,
// gets nothing (CR 400.7). The +1/+1 is a layer-7c scoped effect until
// end of turn, a data record like prowess's.
//
// ORDERING (CR 603.3b). Exalted triggers commute (#1511's argument for
// prowess): each reads only its pinned attacker and writes only a
// +1/+1 to it, so a batch of nothing but exalted triggers needs no
// ordering prompt.

import "github.com/google/uuid"

// KeywordExalted is CR 702.83's canonical token.
const KeywordExalted = "exalted"

// exaltedLabel is the stack label every exalted trigger carries, and
// the label of its +1/+1. It is also the declared name of the retired
// catalog row (effects.Exalted's Key), which is what the restore alias
// below matches on: never change it.
const exaltedLabel = "Exalted — +1/+1 until end of turn"

// exaltedPumpKey is the body every exalted trigger names (ADR 0041 P9):
// an on-disk identity, never renamed or reused. An older binary refuses
// a restore point naming it with ErrUnknownEffectKey, which is the
// designed rollback case. Params.Object is the lone attacker.
const exaltedPumpKey = "exalted/pump"

// exaltedPumpBody is the registered body's reference, built from the
// key and registered in init, as annihilator's is.
var exaltedPumpBody = BodyRef{key: exaltedPumpKey}

func init() {
	DelayedBody(exaltedPumpKey, func(g *Game, item *StackItem, p EffectParams) error {
		return g.resolveExaltedLocked(item, p.Object)
	})
}

// exaltedTrigger is the one TriggeredAbility a single instance of
// exalted is. A package-level value, never mutated: keywordTriggersFor
// hands out copies of it.
var exaltedTrigger = TriggeredAbility{
	Keyword: KeywordExalted,
	Watches: []EventKind{EventAttack},
	AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
		return ev.Actor == source.Controller && ev.CardID != uuid.Nil && AttackedAlone(g)
	},
	Build: func(ev Event, source *Card, _ Characteristic, g *Game) *StackItem {
		attacker := ObjectRef{ID: ev.CardID}
		if c := findBattlefieldCard(g, ev.CardID); c != nil {
			attacker.Epoch = c.ObjectEpoch
		}
		item := NewKeyedTriggeredItem(source, exaltedLabel, exaltedPumpBody, EffectParams{Object: attacker})
		// Exalted instances commute with each other (#1511's argument
		// for prowess): see the file comment.
		item.Commutes = true
		return item
	},
}

// AttackedAlone reports whether exactly one creature is attacking right
// now: CR 506.5's "attacks alone", counted over the whole declaration,
// not merely one player's creatures. DeclareAttackers stamps every
// attacker's AttackingTarget before it announces any EventAttack, so a
// lone declaration reads as genuinely alone rather than "alone so far".
//
// Caller must hold g.mu (a trigger's AppliesTo, or an effect).
func AttackedAlone(g *Game) bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n == 1
}

// resolveExaltedLocked is "that creature gets +1/+1 until end of turn"
// for the attacker the trigger pinned. An attacker no longer on the
// battlefield as the same object gets nothing (CR 400.7). The source
// leaving does not matter: exalted has no "if" clause.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) resolveExaltedLocked(item *StackItem, attacker ObjectRef) error {
	if item == nil || attacker.ID == uuid.Nil {
		return nil
	}
	c := findBattlefieldCard(g, attacker.ID)
	if c == nil || c.ObjectEpoch != attacker.Epoch {
		return nil
	}
	g.RegisterScopedEffectForEffect(item.SourceCardID,
		[]AffectedObject{PinObject(c.InstanceID, c.EnteredBattlefieldAt)},
		[]Mod{ModifyPTMod(1, 1)}, g.UntilEndOfTurnDuration(), exaltedLabel)
	return nil
}

// ExaltedCount reports how many instances of exalted the card has — a
// read for tests, the bot and the view, through the same walk the
// harvester uses.
func ExaltedCount(c *Card) int {
	return countAbilityTokens(c, KeywordExalted)
}

// restoreRetiredExaltedRow is the owner's answer 3 to the ADR 0101
// amendment of 2026-10-08: a one-entry restore alias. Before #2538 the
// Hierarchs' exalted was a catalog row, so a restore point written then
// may hold its trigger as a `catalog/triggered` item whose ref names
// that row. The row is gone. Rather than restore the item as a manual
// one and flag the Hierarch for the rest of the game, the item is
// rewritten to the keyword's own body, with the lone attacker taken
// from the item's carried triggering event.
//
// It answers false, changing nothing, for every other lost ref, and
// for an exalted item that carries no triggering event to read the
// attacker from.
func restoreRetiredExaltedRow(out *StackItem) bool {
	ref, ok := retiredExaltedAttacker(out)
	if !ok {
		return false
	}
	out.Body = exaltedPumpKey
	out.Params = EffectParams{Object: ref}
	out.Effect = bodyEffect(exaltedPumpKey, out.Params)
	out.targetSpec, out.modeSpec = nil, nil
	out.Commutes = true
	return true
}

// retiredExaltedAttacker reports whether a stack item is the retired
// exalted row's trigger, and the attacker it pins.
func retiredExaltedAttacker(it *StackItem) (ObjectRef, bool) {
	if it == nil || it.Body != CatalogTriggeredBodyKey || it.Params.Ability == nil {
		return ObjectRef{}, false
	}
	return retiredExaltedRef(*it.Params.Ability, it.Trigger)
}

// retiredExaltedRef is retiredExaltedAttacker over a ref and the
// trigger context, so the boot report (LostStackAbilities) asks the
// same question restore does.
func retiredExaltedRef(ref AbilityRef, tc *TriggerContext) (ObjectRef, bool) {
	if ref.Slot != AbilitySlotTriggered || ref.Name != exaltedLabel {
		return ObjectRef{}, false
	}
	if tc == nil || tc.Event.Kind != EventAttack || tc.Event.CardID == uuid.Nil {
		return ObjectRef{}, false
	}
	attacker := ObjectRef{ID: tc.Event.CardID}
	if tc.Object != nil && tc.Object.ID == tc.Event.CardID {
		attacker.Epoch = tc.Object.Epoch
	}
	return attacker, true
}

// retiredExaltedRows is how many catalog trigger rows the entry for key
// had before #2538 turned them into its printed exalted: one per
// exalted instance it declares. The #522 parity check adds it to the
// running catalog's triggered count, so a Hierarch captured with its
// old row is not counted as having lost one (owner answer 3: the
// Hierarch is not flagged).
func retiredExaltedRows(key string) int {
	if key == "" || CatalogPrintedKeywords == nil {
		return 0
	}
	n := 0
	for _, kw := range CatalogPrintedKeywords(key) {
		if kw == KeywordExalted {
			n++
		}
	}
	return n
}
