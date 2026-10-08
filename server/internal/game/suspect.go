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
//   - Menace is a keyword, granted in layer 6 at the moment the
//     permanent became suspected (SuspectedAt, CR 613.7). A "loses all
//     abilities" older than that leaves the menace; a newer one takes
//     it away. Everything downstream — combat's block-count check, the
//     enumerator, the badge — reads the one effective ability list.
//   - "Can't block" is a restriction bit (CantBlock, restrictions.go),
//     written by the same effect. Restrictions have no layer (CR 613)
//     and nothing clears them, so a suspected creature that loses all
//     abilities after the fact STILL cannot block. That is the one
//     place this is stricter than a literal reading of the rule, which
//     calls "can't block" an ability the creature has: it errs toward
//     the restriction, never toward a creature that blocks when the
//     table expected it not to. The block gate, the enumerator and the
//     view already read CantBlock through Restricted, so no consumer
//     changed.
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

// SuspectForEffect is CR 701.60a: the named battlefield creature
// becomes suspected. It reports whether it did.
//
// It does nothing, and says false, for a permanent that is not on the
// battlefield (the target left in response, CR 608.2b), one that is not
// a creature (CR 701.60a — only creatures can be suspected), and one
// that is already suspected (CR 701.60d: it "can't become suspected
// again", which is also what keeps its timestamp from being
// rewritten). None of those is an error: each is ordinary play.
//
// Caller must hold g.mu in write mode.
func (g *Game) SuspectForEffect(cardID uuid.UUID) bool {
	c := findBattlefieldCard(g, cardID)
	if c == nil || c.Suspected || !c.IsCreature() {
		return false
	}
	c.Suspected = true
	c.SuspectedAt = timeNowUnixNano()
	// The menace grant and the block restriction both come out of the
	// layer pass, so the cached characteristics are stale now.
	g.layerVersion.Add(1)
	return true
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
// layer-6 continuous effect: menace, and the can't-block restriction.
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
	c.Restrictions |= CantBlock
}
func (e suspectedEffect) RemovesAbilities() bool      { return false }
func (e suspectedEffect) ContinuesAfterRemoval() bool { return false }

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
