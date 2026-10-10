package game

import "github.com/google/uuid"

// suspect.go — suspect (CR 701.60, Murders at Karlov Manor), the
// keyword action whose result is a DESIGNATION with continuous effects
// of its own (ADR 0071 amendment 2026-10-08, #2698).
//
//	CR 701.60a  Some spells and abilities instruct a player to suspect a
//	            creature. That creature becomes suspected until it
//	            leaves the battlefield or a spell or ability causes it
//	            to no longer be suspected.
//	CR 701.60c  A suspected permanent has menace and "This creature
//	            can't block" for as long as it's suspected.
//	CR 701.60d  A suspected permanent can't become suspected again.
//
// (The lettering is the issue's, #2698; the sentences are the rule.)
//
// # Why this is not an ActiveWhen gate
//
// Monstrous, Solved and Harnessed switch a permanent's own PRINTED
// abilities on, and a Spec declares the gate. Suspect does the opposite:
// it GIVES any creature two things its card never prints, so there is
// no declaration to hang a gate on. The designation is a bool on the
// card (Card.Suspected), and suspectContinuousEffectsLocked turns it
// into one layer-6 effect per suspected permanent — the shape the
// keyword counters (keyword_counters.go) and the Hall of the Bandit
// Lord rider (mana_spend_rider.go) already use for "this object has a
// keyword for a reason that is not one of its abilities". The effect
// has no source permanent, so CR 613.6's silencing never reaches it.
//
// # The two halves of the grant
//
// Both are abilities in the effective ability list, appended by the one
// layer-6 effect at the moment the permanent became suspected
// (SuspectedAt, CR 613.7). A "loses all abilities" older than that
// leaves them; a newer one takes both away (CR 701.60c, #2737).
//
//   - Menace is the keyword itself. Everything downstream — combat's
//     block-count check, the enumerator, the badge — reads the list.
//   - "Can't block" is the token KeywordCantBlock. Restrictions have no
//     layer and nothing clears them, so the token is only the ABILITY;
//     foldSuspectedCantBlockLocked turns it into the CantBlock bit once
//     the pass is over (CR 613.11), the way foldUnleashLocked does for
//     unleash. The block gate, the enumerator and the view already read
//     CantBlock through Restricted, so no consumer changed.
//
// # Lifetime
//
//   - Cleared when the permanent leaves the battlefield (CR 400.7): zone
//     motion and the entry tail both reset it, beside Monstrous.
//   - KEPT through a control change. A creature stolen with Caught
//     Red-Handed and suspected by the same spell stays suspected under
//     its new controller, and so does one an opponent suspected.
//   - Not copiable (CR 707.2): CopiableValuesOf never reads it, so a
//     Clone of a suspected creature is a plain creature.
//   - Carried by clone and the snapshot (snapshot.go).
//
// # Can't become suspected
//
// Airtight Alibi's "can't become suspected" is the CantBecomeSuspected
// restriction (restrictions.go, #2733). CanBecomeSuspected is the one
// place it is read, and SuspectForEffect asks it, so no suspect
// instruction can miss it.

// SuspectForEffect is CR 701.60a: the named battlefield creature
// becomes suspected. It reports whether it did.
//
// It does nothing, and says false, for a permanent that is not on the
// battlefield (the target left in response, CR 608.2b), one that is not
// a creature (CR 701.60a — only creatures can be suspected), one that
// is already suspected (CR 701.60d: it "can't become suspected again",
// which is also what keeps its timestamp from being rewritten), and one
// an effect says can't become suspected (Airtight Alibi, #2733). None
// of those is an error: each is ordinary play.
//
// Caller must hold g.mu in write mode.
func (g *Game) SuspectForEffect(cardID uuid.UUID) bool {
	if !g.CanBecomeSuspected(cardID) {
		return false
	}
	c := findBattlefieldCard(g, cardID)
	c.Suspected = true
	c.SuspectedAt = timeNowUnixNano()
	// The menace and can't-block grants both come out of the layer pass, so the cached characteristics are stale now.
	g.layerVersion.Add(1)
	return true
}

// CanBecomeSuspected reports whether suspecting the named permanent
// would do anything: it is a creature on the battlefield, it is not
// suspected already (CR 701.60d), and no effect says it can't become
// suspected (the CantBecomeSuspected restriction, Airtight Alibi's
// "Enchanted creature … can't become suspected", #2733). The
// restriction is read from the finished layer pass, so the layers are
// brought up to date first. A card that offers a choice of what to
// suspect (Frantic Scapegoat) asks this to build the offer, so the
// offer and the action cannot disagree.
//
// Caller must hold g.mu in write mode.
func (g *Game) CanBecomeSuspected(cardID uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	c := findBattlefieldCard(g, cardID)
	return c != nil && !c.Suspected && c.IsCreature() && !Restricted(c, CantBecomeSuspected)
}

// UnsuspectForEffect is "it's no longer suspected" (CR 701.60a's
// second clause): the named battlefield permanent stops being
// suspected. It reports whether it was.
//
// Caller must hold g.mu in write mode.
func (g *Game) UnsuspectForEffect(cardID uuid.UUID) bool {
	c := findBattlefieldCard(g, cardID)
	if c == nil || !c.Suspected {
		return false
	}
	c.Suspected = false
	c.SuspectedAt = 0
	g.layerVersion.Add(1)
	return true
}

// IsSuspected reports whether the named permanent is on the battlefield
// and suspected. False for anything not on the battlefield — CR 400.7,
// a permanent that left is a new object and is not suspected.
//
// Caller must hold g.mu.
func (g *Game) IsSuspected(cardID uuid.UUID) bool {
	c := findBattlefieldCard(g, cardID)
	return c != nil && c.Suspected
}

// suspectedEffect is one suspected permanent's CR 701.60c grant, as a
// layer-6 continuous effect: the menace and can't-block abilities.
type suspectedEffect struct {
	target    uuid.UUID
	timestamp int64
}

func (e suspectedEffect) Layer() (Layer, SubLayer) { return Layer6Ability, 0 }
func (e suspectedEffect) Timestamp() int64         { return e.timestamp }
func (e suspectedEffect) AppliesTo(target *Card, _ *Game) bool {
	return target != nil && target.InstanceID == e.target
}
func (e suspectedEffect) Apply(c *Characteristic, _ *Card, _ *Game) {
	c.Abilities = AppendKeywordAbility(c.Abilities, "menace")
	c.Abilities = AppendKeywordAbility(c.Abilities, KeywordCantBlock)
}
func (e suspectedEffect) RemovesAbilities() bool      { return false }
func (e suspectedEffect) ContinuesAfterRemoval() bool { return false }

// KeywordCantBlock is the ability token for "This creature can't block"
// as CR 701.60c grants it. It is not a printed keyword: it exists so the
// grant sits in the ability list, where a layer-6 removal reaches it.
const KeywordCantBlock = "can't block"

// foldSuspectedCantBlockLocked turns a granted KeywordCantBlock ability
// into the CantBlock restriction once the layer pass is over, so a
// suspected permanent that has lost its abilities may block (CR 701.60c).
// A CantBlock from any other source is untouched.
//
// Caller must hold g.mu (write).
func (g *Game) foldSuspectedCantBlockLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective != nil && containsKeyword(c.effective.Abilities, KeywordCantBlock) {
			c.effective.Restrictions |= CantBlock
		}
	}
}

// suspectContinuousEffectsLocked is the layer pass's source list for
// the suspected designation: one effect per suspected battlefield
// permanent, ordered at the moment it became suspected. A restore from
// a point that predates SuspectedAt orders it at the permanent's own
// timestamp, the earliest moment it could have been suspected.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) suspectContinuousEffectsLocked() []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.Suspected {
			continue
		}
		ts := c.SuspectedAt
		if ts == 0 {
			ts = c.layerTimestamp()
		}
		out = append(out, suspectedEffect{target: c.InstanceID, timestamp: ts})
	}
	return out
}
