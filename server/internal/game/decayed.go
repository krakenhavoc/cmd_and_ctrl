package game

// decayed.go — CR 702.147, decayed as a keyword the engine derives from
// an object's ability list rather than from its catalog entry (#2650).
//
//	702.147a Decayed represents a static ability and a triggered
//	         ability. "Decayed" means "This creature can't block" and
//	         "When this creature attacks, sacrifice it at end of
//	         combat."
//
// WHY A TOKEN. Nearly every decayed creature is a 2/2 black Zombie
// TOKEN made by some other card's text, so a catalog entry on the
// creature cannot carry it. A token in Characteristic.Abilities reaches
// the tokens, a deck-imported card, a grant and a decayed counter
// (CR 122.1b, Rot-Curse Rakshasa) through one walk.
//
// THE STATIC. "This creature can't block" is a rule over the finished
// ability list, so foldDecayedLocked adds CantBlock once the layer pass
// is over, beside unleash's fold (riot.go). The block gate, the
// enumerator and the view all read it through Restricted, and a
// creature that loses its abilities loses the restriction with them.
//
// THE TRIGGER. "When this creature attacks" watches EventAttack for the
// source itself, which is emitted only for a declared attacker (CR
// 508.3a): a decayed token put onto the battlefield attacking never
// triggers it, and is not sacrificed. Resolving, it schedules a CR
// 603.7 delayed trigger for the beginning of the end of combat step
// that sacrifices the creature. The creature is pinned by instance ID
// and object epoch, so one that left the battlefield and came back is a
// new object and is not sacrificed (CR 400.7). It is sacrificed whether
// or not it is still attacking: the trigger says "sacrifice it", not
// "sacrifice it if it's attacking".
//
// A combat ended early (CR 724.2, end_combat.go) never begins its end
// of combat step, so the delayed trigger stays queued for the next one,
// as myriad's exile does.
//
// MULTIPLE INSTANCES. CR 702.147 has no rule of its own on the point,
// so CR 113.2c applies and decayed is CUMULATIVE: each instance is its
// own trigger. The second sacrifice finds nothing to sacrifice.

import "github.com/google/uuid"

// KeywordDecayed is CR 702.147's canonical token.
const KeywordDecayed = "decayed"

// decayedLabel is the stack label of the attack trigger.
const decayedLabel = "Decayed — sacrifice it at end of combat"

// decayedSacrificeLabel is the stack label of the delayed trigger.
const decayedSacrificeLabel = "Decayed — sacrifice it"

// The bodies' keys are on-disk identities (ADR 0041 P9): never renamed
// or reused. Params.Object is the decayed creature in both.
const (
	decayedScheduleKey  = "decayed/schedule"
	decayedSacrificeKey = "decayed/sacrifice"
)

var (
	decayedScheduleBody  = BodyRef{key: decayedScheduleKey}
	decayedSacrificeBody = BodyRef{key: decayedSacrificeKey}
)

func init() {
	DelayedBody(decayedScheduleKey, func(g *Game, item *StackItem, p EffectParams) error {
		return g.scheduleDecayedSacrificeLocked(item, p.Object)
	})
	DelayedBody(decayedSacrificeKey, func(g *Game, _ *StackItem, p EffectParams) error {
		return g.resolveDecayedSacrificeLocked(p.Object)
	})
}

// decayedTrigger is the one TriggeredAbility a single instance of
// decayed is. A package-level value, never mutated: keywordTriggersFor
// hands out copies of it.
var decayedTrigger = TriggeredAbility{
	Keyword: KeywordDecayed,
	Watches: []EventKind{EventAttack},
	AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
		return ev.CardID != uuid.Nil && ev.CardID == source.InstanceID
	},
	Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
		self := ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
		return NewKeyedTriggeredItem(source, decayedLabel, decayedScheduleBody, EffectParams{Object: self})
	},
}

// scheduleDecayedSacrificeLocked is the trigger resolving: "sacrifice
// it at end of combat" becomes a delayed trigger at the beginning of
// the next end of combat step, controlled by the trigger's controller
// (CR 603.7d). It is scheduled even if the creature has already left:
// the delayed trigger checks the object when it fires.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) scheduleDecayedSacrificeLocked(item *StackItem, self ObjectRef) error {
	if item == nil || self.ID == uuid.Nil {
		return nil
	}
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   item.Controller,
		SourceCardID: item.SourceCardID,
		SourceObject: item.SourceObject,
		Label:        decayedSacrificeLabel,
		At:           StepEndCombat,
		Body:         decayedSacrificeBody,
		Params:       EffectParams{Object: self},
	})
	return nil
}

// resolveDecayedSacrificeLocked sacrifices the pinned creature if it is
// still on the battlefield as the same object (CR 400.7). Nothing to
// sacrifice is not an error (CR 608.2c).
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) resolveDecayedSacrificeLocked(self ObjectRef) error {
	c := findBattlefieldCard(g, self.ID)
	if c == nil || c.ObjectEpoch != self.Epoch {
		return nil
	}
	return g.SacrificePermanentForEffect(c.InstanceID)
}

// foldDecayedLocked adds CantBlock to every permanent whose finished
// ability list has decayed: CR 702.147a's "This creature can't block".
//
// Caller must hold g.mu (the layer pass).
func (g *Game) foldDecayedLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective == nil || !containsKeyword(c.effective.Abilities, KeywordDecayed) {
			continue
		}
		c.effective.Restrictions |= CantBlock
	}
}

// DecayedCount reports how many instances of decayed the card has — a
// read for tests, the bot and the view, through the same walk the
// harvester uses.
func DecayedCount(c *Card) int {
	return countAbilityTokens(c, KeywordDecayed)
}
