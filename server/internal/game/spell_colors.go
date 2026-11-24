package game

import "github.com/google/uuid"

// spell_colors.go — layer 5 for SPELLS (#6088).
//
//	CR 613.1e  "Layer 5: Color-changing effects are applied."
//
// "Target spell or permanent becomes the color of your choice until end of
// turn" (Illusion) and "target spell or permanent becomes white until end
// of turn" (Eight-and-a-Half-Tails) change the colour of an object on the
// stack. A spell is an object (CR 109.1), so the change is a ScopedEffect
// record PINNED to the spell with a setColors mod, exactly as a keyword
// grant to a spell is (spell_keywords.go, ADR 0107 §3): data, ended with
// the object (CR 400.7) or with its duration, whichever is first.
//
// The stack step of the layer pass materialises it onto the stack card
// (Card.stackColors), which EffectiveColors and Effective() read in place
// of the printed colours, so "a white spell", protection from a colour (CR
// 702.16) and the stack view all see the new colour. A permanent spell
// hands the record to the permanent it becomes (CR 400.7a,
// repinSpellControlLocked), where the battlefield pass applies it in layer
// 5 for the rest of its duration.
//
// Only a setColors mod is read here. Its records apply in timestamp order
// (CR 613.7), so the latest one wins. A battlefield static that changes
// the colour of cards in other zones (zone_layers.go) is the baseline this
// step replaces, which is the timestamp order of every printed case: the
// spell's colour is set by a resolving effect after the spell was cast.

// spellSetColorsLocked is the colour the pinned setColors records give the
// spell `c`: the latest one's colours, and true; false when none does.
//
// Caller must hold g.mu (write).
func (g *Game) spellSetColorsLocked(c *Card) ([]string, bool) {
	var (
		colors []string
		ts     int64
		found  bool
	)
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if !pinsStackObject(e.Affected, c.InstanceID, c.ObjectEpoch) {
			continue
		}
		for _, m := range e.Mods {
			if m.Kind != ModSetColors || (found && e.Timestamp < ts) {
				continue
			}
			colors, ts, found = copyStrings(m.Colors), e.Timestamp, true
		}
	}
	return colors, found
}

// SetSpellColorsForEffect makes the spell `spellID` on the stack the
// colours named (none is colourless) for the duration `d`, and for no
// longer than it is that object on the stack — or, for a permanent spell,
// on the permanent it becomes (CR 400.7a) for the rest of `d`. A
// colour-changing effect (CR 611.2, 613.1e) written as a ScopedEffect
// pinned to the spell.
//
// Reports false, registering nothing, when `spellID` is not a spell on the
// stack: it resolved or was countered first (CR 608.2b).
//
// It recomputes before it returns, so the spell has the colour at once.
//
// Caller must hold g.mu (write).
func (g *Game) SetSpellColorsForEffect(sourceID, spellID uuid.UUID, colors []string, d Duration, label string) bool {
	card, _, ok := g.stolenSpellLocked(spellID)
	if !ok {
		return false
	}
	if !g.registerScopedEffectLocked(sourceID,
		[]AffectedObject{PinStackObject(spellID, card.ObjectEpoch)},
		[]Mod{SetColorsMod(colors...)},
		g.PinnedToStack(d, spellID), label, timeNowUnixNano()) {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	return true
}
