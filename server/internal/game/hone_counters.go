package game

import "github.com/google/uuid"

// hone_counters.go — CR 122.1j, hone counters (#4500, The Hobbit).
//
// CR 122.1j: "A hone counter on an Equipment gives +1/+0 to any
// creature that Equipment is attached to." The counter sits on the
// Equipment and changes ANOTHER permanent, so it cannot ride the
// counters-after-the-pass step that +1/+1 counters use
// (ptWithCounters reads the creature's own Counters): it is a layer
// 7c effect (CR 613.4c, "effects and counters that modify power and/or
// toughness"), gathered from the Equipment and applied to whatever the
// Equipment is attached to as the pass reaches 7c.
//
// The effect is source-less in the keyword-counter sense
// (keyword_counters.go): a counter is not an ability of the permanent it
// sits on, so a "loses all abilities" on the Equipment does not switch
// its hone counters off. It follows the Equipment because it is
// re-gathered from the battlefield on every pass: re-equipping moves
// AttachedTo, and the next pass puts the bonus on the new creature. A
// counter change (EventCounterPlaced) and an attach or unattach both
// invalidate the pass already, so the bonus is current when the board is
// read. Several Equipment on one creature each contribute their own
// effect and the +N/+0 values add.
//
// Nothing about it is stored beyond the counter itself, so the snapshot
// shape is unchanged: "hone" is one more key in the free-form
// Card.Counters map, which an older restore point simply never holds.

// CounterHone is the hone counter (CR 122.1j). Each one on an Equipment
// gives the creature it is attached to +1/+0.
const CounterHone = "hone"

// honeCounterEffect is the +N/+0 one Equipment's hone counters give the
// creature it is attached to, as a CR 613.4c layer-7c effect.
type honeCounterEffect struct {
	// target is the equipped creature's instance ID.
	target uuid.UUID
	// power is the Equipment's hone counter count.
	power int
	// timestamp is the Equipment's own CR 613.7e timestamp, which an
	// attach moves.
	timestamp int64
}

func (e honeCounterEffect) Layer() (Layer, SubLayer) { return Layer7PT, SubLayer7C_Modify }
func (e honeCounterEffect) Timestamp() int64         { return e.timestamp }
func (e honeCounterEffect) AppliesTo(target *Card, _ *Game) bool {
	return target != nil && target.InstanceID == e.target
}
func (e honeCounterEffect) Apply(c *Characteristic, _ *Card, _ *Game) {
	c.Power = SaturatingAdd(c.Power, e.power)
}
func (e honeCounterEffect) RemovesAbilities() bool      { return false }
func (e honeCounterEffect) ContinuesAfterRemoval() bool { return false }

// honeCounterEffectsLocked is the layer pass's source list for hone
// counters: one effect per attached Equipment that carries any. An
// Equipment that is not attached to a permanent gives nothing (CR
// 122.1j says "any creature that Equipment is attached to"), and a hone
// counter on a permanent that is not an Equipment is storage.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) honeCounterEffectsLocked() []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		n := c.Counters[CounterHone]
		if n <= 0 || c.AttachedTo.Kind != TargetCard || c.AttachedTo.ID == uuid.Nil {
			continue
		}
		if !c.HasSubtype("Equipment") {
			continue
		}
		out = append(out, honeCounterEffect{
			target:    c.AttachedTo.ID,
			power:     n,
			timestamp: c.layerTimestamp(),
		})
	}
	return out
}
