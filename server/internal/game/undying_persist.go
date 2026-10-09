package game

// undying_persist.go — CR 702.93 and CR 702.79, the first DIES keyword
// triggers the engine derives from an object's ability list (#2075,
// ADR 0113 §4). Prowess (prowess.go) and evolve (evolve.go) are the
// model; these two differ in one way that matters: they trigger on the
// permanent LEAVING, so they are read off its last-known information
// rather than off a live permanent.
//
//	702.93a Undying is a triggered ability. "Undying" means "When this
//	        permanent is put into a graveyard from the battlefield, if
//	        it had no +1/+1 counters on it, return it to the
//	        battlefield under its owner's control with a +1/+1 counter
//	        on it."
//	702.79a Persist is the same with -1/-1 counters.
//
// WHERE THEY ARE HARVESTED. Not by TriggersForCard: a permanent on the
// battlefield has not died, and keywordTriggersFor's triggers are read
// off live permanents. harvestLTB asks ltbKeywordTriggersFor for the
// tokens on the LAST-KNOWN ability list (CR 603.10a), before it looks
// for a catalog key, so a creature with no catalog entry (Young Wolf, a
// persist token) still has them. A printed instance (stamped by the
// deck importer), a token template's and a layer-6 grant (Mikaeus, the
// Unhallowed) all reach the list; a creature that died having lost all
// abilities (CR 613.1f) has none, and a Clone that copied a persist
// creature has the copied one (layer 1).
//
// THE INTERVENING IF (CR 603.4). The counters are read twice, both
// times as the permanent last existed on the battlefield: as it dies
// (lastKnownCounters, which the harvest still holds) and as the trigger
// resolves (its PermanentInfo record). Both read the same information,
// so the second check cannot fail; it is made because the rule says so.
// A persist creature with a +1/+1 counter that got enough -1/-1
// counters to die still had -1/-1 counters "at that point" (the persist
// rulings), so it does not trigger. CR 704.5q's annihilation
// of +1/+1 against -1/-1 counters is what makes the loops work: an
// undying creature that came back with its +1/+1 counter and then got
// a -1/-1 counter has neither, and returns again.
//
// WHAT COMES BACK (CR 400.7e). The ability can find the new object the
// card became in the graveyard, and only that object. The item's
// SourceObject names it (the graveyard epoch, read when the ability
// triggers), and the body returns the card only while it is still in a
// graveyard as that same object. A card exiled in response, or exiled
// and put back, is a different object and stays where it is. A token
// has ceased to exist (CR 111.7, 704.5d): the ability triggered, and
// there is nothing to return. Neither is an error.
//
// CONTROL. The trigger is controlled by whoever controlled the
// permanent as it left (CR 603.3a — the harvest's LKI controller); the
// card returns under its OWNER's control (702.93a, 702.79a).
//
// THE COUNTER (CR 122.6). It rides the entry event, so Hardened Scales,
// Vizier of Remedies and every other CR 614 counter replacement see it
// as being put on the creature.
//
// MULTIPLE INSTANCES (CR 113.2c). Each instance triggers. Both tokens
// are cumulative (KeywordIsCumulative), so a granted instance on a
// printed one is two triggers; whichever resolves first returns the
// card, and the rest find a new object on the battlefield and do
// nothing. Two instances on one creature share a source and a label, so
// the CR 603.3b ordering prompt is skipped; the triggers are NOT marked
// Commutes, because several creatures returning in different orders get
// different timestamps and different entry triggers.

import (
	"errors"

	"github.com/google/uuid"
)

// KeywordUndying is CR 702.93's canonical token.
const KeywordUndying = "undying"

// KeywordPersist is CR 702.79's canonical token.
const KeywordPersist = "persist"

// diesReturnKeyword is what tells undying from persist: the counter the
// condition reads and the return puts on, the label's words, and the
// on-disk body key.
type diesReturnKeyword struct {
	keyword string
	counter string
	label   string
	body    BodyRef
}

// undyingReturnBody and persistReturnBody are registered under
// "undying/return" and "persist/return" (ADR 0113 §4, ADR 0041 P9): on-
// disk identities, never renamed. An older binary refuses a restore
// point naming one with ErrUnknownEffectKey, which is the designed
// rollback case.
var (
	undyingReturnBody = SimpleDelayedBody("undying/return", func(g *Game, item *StackItem) error {
		return resolveDiesReturn(g, item, CounterPlusOne)
	})
	persistReturnBody = SimpleDelayedBody("persist/return", func(g *Game, item *StackItem) error {
		return resolveDiesReturn(g, item, CounterMinusOne)
	})
)

var (
	undyingKeyword = diesReturnKeyword{keyword: KeywordUndying, counter: CounterPlusOne, label: "Undying", body: undyingReturnBody}
	persistKeyword = diesReturnKeyword{keyword: KeywordPersist, counter: CounterMinusOne, label: "Persist", body: persistReturnBody}

	// undyingTrigger and persistTrigger are the one TriggeredAbility a
	// single instance of each keyword is. Package-level values, never
	// mutated: ltbKeywordTriggersFor hands out copies.
	undyingTrigger = diesReturnTrigger(undyingKeyword)
	persistTrigger = diesReturnTrigger(persistKeyword)
)

func diesReturnTrigger(k diesReturnKeyword) TriggeredAbility {
	return TriggeredAbility{
		Keyword: k.keyword,
		Watches: []EventKind{EventLTB},
		// "When this permanent is put into a graveyard from the
		// battlefield, if it had no <kind> counters on it". A permanent
		// exiled instead (Rest in Peace) was never put into a graveyard.
		// It need not be a creature: the condition is about the
		// permanent (the Murderous Redcap ruling).
		AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
			if ev.CardID == uuid.Nil || ev.CardID != source.InstanceID || ev.NewZone != ZoneGraveyard {
				return false
			}
			return g.lastKnownCounters[ev.CardID][k.counter] <= 0
		},
		Build: func(_ Event, source *Card, lki Characteristic, g *Game) *StackItem {
			name := lki.Name
			if name == "" {
				name = source.Name
			}
			item := NewKeyedTriggeredItem(source, k.label+" — return "+name, k.body, EffectParams{})
			// CR 400.7e: the object this ability can find is the one
			// the card became in the graveyard, read now, as it
			// triggers. stampTriggerSource leaves a stamped item alone.
			if epoch := g.cardObjectEpochLocked(source.InstanceID); epoch >= 0 {
				item.SourceObject = ObjectRef{ID: source.InstanceID, Epoch: epoch}
			}
			return item
		},
	}
}

// ltbKeywordTriggersFor is the dies keyword triggers in a departed
// permanent's last-known ability list: one per undying token, then one
// per persist token (CR 113.2c). harvestLTB's counterpart of
// keywordTriggersFor.
func ltbKeywordTriggersFor(lki Characteristic) []TriggeredAbility {
	undying, persist, modular := 0, 0, 0
	for _, a := range lki.Abilities {
		switch a {
		case KeywordUndying:
			undying++
		case KeywordPersist:
			persist++
		default:
			// #2012: one modular dies trigger per instance (modular.go).
			if _, ok := ModularValue(a); ok {
				modular++
			}
		}
	}
	if undying+persist+modular == 0 {
		return nil
	}
	out := make([]TriggeredAbility, 0, undying+persist+modular)
	for i := 0; i < undying; i++ {
		out = append(out, undyingTrigger)
	}
	for i := 0; i < persist; i++ {
		out = append(out, persistTrigger)
	}
	for i := 0; i < modular; i++ {
		out = append(out, modularTrigger)
	}
	return out
}

// resolveDiesReturn re-checks the condition against the departed
// permanent's last-known counters (CR 603.4) and returns the card from
// its graveyard under its owner's control with one `counter` on the
// entry event — only while the card is still the graveyard object the
// ability triggered with (CR 400.7e). A package-level func, so it
// captures nothing and survives Clone.
func resolveDiesReturn(g *Game, item *StackItem, counter string) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	departed, ok := g.PermanentForEffect(item.Trigger.Object.Ref())
	if !ok || departed.Counters[counter] > 0 {
		return nil
	}
	ref := item.SourceObject
	if ref.ID != item.SourceCardID || ref.ID == uuid.Nil {
		return nil
	}
	z := g.findCardZoneLocked(ref.ID)
	if z == nil || z.Kind != ZoneGraveyard || g.cardObjectEpochLocked(ref.ID) != ref.Epoch {
		return nil
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(ref.ID, uuid.Nil, false, map[string]int{counter: 1})
	if errors.Is(err, ErrCardNotFound) {
		return nil
	}
	return err
}
