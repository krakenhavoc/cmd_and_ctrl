package game

import (
	"slices"

	"github.com/google/uuid"
)

// mode_memory.go — ADR 0097: "choose one that hasn't been chosen"
// (#1749). Thirty-odd cards restrict a modal triggered or activated
// ability to the modes that ability has not chosen before, either
// "this turn" (Gala Greeters, Monument to Endurance, Kargan
// Intimidator) or with no duration at all (Silent Hallcreeper, Demonic
// Pact). There is no Comprehensive Rules entry for it; the card text
// and the rulings govern, and the rulings decide every shape here:
//
//   - PER OBJECT (CR 400.7; the Silent Hallcreeper and Breeches
//     rulings). The memory belongs to one ability of one object. A
//     second copy of the card remembers nothing of the first, and a
//     permanent that leaves and returns is a new object that remembers
//     nothing at all.
//   - NOT PER CONTROLLER (the Demonic Pact ruling: "it doesn't matter
//     who has chosen any particular mode"). A change of control keeps
//     the memory, because nothing here is keyed by a player.
//   - RECORDED WHEN CHOSEN (the Demonic Pact ruling: a mode chosen for
//     an instance that never resolves "still counts as being chosen").
//     A trigger records as its mode_pick is answered; an activated
//     ability records once the activation has put it on the stack.
//   - EXHAUSTION REMOVES THE TRIGGER (CR 700.2b, the Breeches ruling).
//     An ability left with too few modes to choose is not put on the
//     stack, through the one filter every mode path reads.
//
// The two scopes are two places, and each is flushed by an event the
// engine already has:
//
//   - "this turn" is TurnTally.ModesChosen, keyed by ObjectTallyKey —
//     the #936 key the "once each turn" gates use — and emptied with
//     the rest of the tally when the turn begins;
//   - "ever" is Card.ModesChosen, keyed by the ability's label and
//     cleared by MoveCard with the rest of CR 400.7's forgetting. It is
//     not a copiable value (CR 707.2), so a copy starts empty.

// ModeMemory is the "that hasn't been chosen" restriction on a
// ModeSpec: whether, and for how long, the options an ability has
// chosen are withheld from its later choices. The zero value is no
// restriction, which is every modal card but the few that print the
// words.
type ModeMemory int

const (
	// ModeMemoryNone — the ordinary modal ability: every option is
	// offered every time.
	ModeMemoryNone ModeMemory = iota
	// ModeMemoryThisTurn — "choose one that hasn't been chosen this
	// turn": each option once per turn for this object's ability.
	ModeMemoryThisTurn
	// ModeMemoryEver — "choose one that hasn't been chosen": each
	// option once, ever, for this object's ability.
	ModeMemoryEver
)

// ModeAbility names the ability whose mode memory narrows a choice:
// the object it belongs to (its instance ID and CR 400.7 epoch, read
// when the ability triggered or was activated) and the ability's
// label — a triggered row's Key, an activated row's Label.
//
// The zero value names nothing, and a choice made with it excludes
// nothing. That is a spell's choice: a spell has no permanent object
// to remember with and no printed spell says the words, which
// effects.Register enforces.
type ModeAbility struct {
	Source uuid.UUID
	Epoch  int
	Label  string
}

// ModeAbilityOf is the identity of the ability labelled `label` of
// the object `source` is right now.
func ModeAbilityOf(source Card, label string) ModeAbility {
	return ModeAbility{Source: source.InstanceID, Epoch: source.ObjectEpoch, Label: label}
}

// IsZero reports whether the identity names no ability.
func (a ModeAbility) IsZero() bool { return a.Source == uuid.Nil || a.Label == "" }

// modesChosenLocked lists the option indexes `ab` has already chosen
// under `ms`'s restriction, ascending. Nil when the spec carries no
// restriction or the identity names nothing, so every caller that is
// not an ability with the words gets the empty answer and changes
// nothing.
//
// Caller must hold g.mu.
func (g *Game) modesChosenLocked(ms *ModeSpec, ab ModeAbility) []int {
	if ms == nil || ab.IsZero() {
		return nil
	}
	switch ms.NotChosen {
	case ModeMemoryThisTurn:
		return g.TurnTally.ModesChosen[ObjectTallyKey(ab.Source, ab.Epoch, ab.Label)]
	case ModeMemoryEver:
		if c := g.modeMemoryCardLocked(ab); c != nil {
			return c.ModesChosen[ab.Label]
		}
	}
	return nil
}

// modeMemoryCardLocked is the live card `ab` names, when it is still
// the same object. A card that has changed zones since has a new
// epoch and is a new object with nothing to remember (CR 400.7), so
// it is not the answer. Caller must hold g.mu.
func (g *Game) modeMemoryCardLocked(ab ModeAbility) *Card {
	c := g.findCardByIDLocked(ab.Source)
	if c == nil {
		// Phasing is not a zone change (CR 702.26d): a phased-out
		// permanent is the same object, and its Card value — this
		// field with it — rides into g.PhasedOut and back.
		if g.PhasedOut != nil {
			for i := range g.PhasedOut.Cards {
				if g.PhasedOut.Cards[i].InstanceID == ab.Source {
					c = &g.PhasedOut.Cards[i]
					break
				}
			}
		}
	}
	if c == nil || c.ObjectEpoch != ab.Epoch {
		return nil
	}
	return c
}

// ModesChosenForEffect is modesChosenLocked on the *ForEffect
// surface, for the callers already under g.mu that must answer the
// question the gate answers: the protocol projection, which greys a
// used option, and the tests.
func (g *Game) ModesChosenForEffect(ms *ModeSpec, ab ModeAbility) []int {
	return g.modesChosenLocked(ms, ab)
}

// recordModesChosenLocked writes the options `ab` has just chosen into
// its memory (ADR 0097 Decision 4: at the choice, never at
// resolution), then re-narrows every still-open mode_pick of the same
// ability of the same object — the Gala Greeters ruling's "you must
// still choose different modes for each instance", when several
// instances queued their prompts before the first was answered.
//
// A copy of the ability (CR 700.2g) copies the modes and never comes
// here; each additional instance a trigger doubler adds does, since
// it chose for itself.
//
// Caller must hold g.mu in write mode.
func (g *Game) recordModesChosenLocked(ms *ModeSpec, ab ModeAbility, modes []int) {
	if ms == nil || ms.NotChosen == ModeMemoryNone || ab.IsZero() || len(modes) == 0 {
		return
	}
	switch ms.NotChosen {
	case ModeMemoryThisTurn:
		if g.TurnTally.ModesChosen == nil {
			g.TurnTally.ModesChosen = map[string][]int{}
		}
		key := ObjectTallyKey(ab.Source, ab.Epoch, ab.Label)
		g.TurnTally.ModesChosen[key] = mergeChosenModes(g.TurnTally.ModesChosen[key], modes, len(ms.Options))
	case ModeMemoryEver:
		c := g.modeMemoryCardLocked(ab)
		if c == nil {
			// The object is gone. What it chose can never be read by
			// anything — the card it left behind is a new object — so
			// there is nothing to write.
			break
		}
		if c.ModesChosen == nil {
			c.ModesChosen = map[string][]int{}
		}
		c.ModesChosen[ab.Label] = mergeChosenModes(c.ModesChosen[ab.Label], modes, len(ms.Options))
	}
	g.renarrowModePicksLocked(ab)
}

// mergeChosenModes is `have` plus every in-range index of `add`,
// deduplicated and ascending, in a NEW backing array — the old one
// may be shared with an undo snapshot.
func mergeChosenModes(have, add []int, n int) []int {
	out := append([]int(nil), have...)
	for _, m := range add {
		if m < 0 || m >= n || slices.Contains(out, m) {
			continue
		}
		out = append(out, m)
	}
	slices.Sort(out)
	return out
}

// renarrowModePicksLocked takes the options `ab` has now used off
// every open mode_pick that belongs to it, and withdraws a prompt left
// with too few to fill its minimum: that instance of the ability is
// removed, with no effect (CR 700.2b, the Gala Greeters and Breeches
// rulings). The withdrawal is dropChoiceLocked, the door every prune
// uses, so nothing is counted as a player decision and nothing wedges.
//
// Caller must hold g.mu in write mode.
func (g *Game) renarrowModePicksLocked(ab ModeAbility) {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || c.Kind != PendingChoiceModePick || c.modePickResume == nil {
			continue
		}
		frame := c.modePickResume
		ms := frame.ability.Modes
		if ms == nil || ModeAbilityOf(frame.source, frame.ability.Key) != ab {
			continue
		}
		used := g.modesChosenLocked(ms, ab)
		if len(used) == 0 {
			continue
		}
		var keepIdx []int
		var keepLabel []string
		usedIdx := append([]int(nil), c.ModeUsedIndex...)
		usedLabel := append([]string(nil), c.ModeUsedLabel...)
		for j, idx := range c.ModeOptionIndex {
			label := ""
			if j < len(c.ModeOptionLabel) {
				label = c.ModeOptionLabel[j]
			}
			if slices.Contains(used, idx) {
				usedIdx = append(usedIdx, idx)
				usedLabel = append(usedLabel, label)
				continue
			}
			keepIdx = append(keepIdx, idx)
			keepLabel = append(keepLabel, label)
		}
		if len(keepIdx) == len(c.ModeOptionIndex) {
			continue
		}
		if !EnoughChoosableModes(len(keepIdx), ms) {
			g.dropChoiceLocked(i)
			continue
		}
		c.ModeOptionIndex, c.ModeOptionLabel = keepIdx, keepLabel
		c.ModeUsedIndex, c.ModeUsedLabel = sortUsedModes(usedIdx, usedLabel)
		if !ms.Repeatable && c.ModeMin > len(keepIdx) {
			c.ModeMin = max(ms.Min, len(keepIdx))
		}
	}
}

// usedModeOptions is the used half of a mode_pick's offer: the
// options `used` names, with their labels, in printed order.
func usedModeOptions(ms *ModeSpec, used []int) ([]int, []string) {
	if ms == nil || len(used) == 0 {
		return nil, nil
	}
	idx := make([]int, 0, len(used))
	labels := make([]string, 0, len(used))
	for _, i := range used {
		if i < 0 || i >= len(ms.Options) {
			continue
		}
		idx = append(idx, i)
		labels = append(labels, ms.Options[i].Label)
	}
	return idx, labels
}

// sortUsedModes orders a used-option list by option index, keeping
// each label beside its index.
func sortUsedModes(idx []int, labels []string) ([]int, []string) {
	order := make([]int, len(idx))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int { return idx[a] - idx[b] })
	outIdx := make([]int, 0, len(idx))
	outLabel := make([]string, 0, len(idx))
	for _, o := range order {
		outIdx = append(outIdx, idx[o])
		if o < len(labels) {
			outLabel = append(outLabel, labels[o])
		} else {
			outLabel = append(outLabel, "")
		}
	}
	return outIdx, outLabel
}

// copyModesChosen deep-copies a mode memory map: its own map and its
// own backing array per entry, so an undo snapshot shares nothing
// with the live game. Nil for an empty map.
func copyModesChosen(m map[string][]int) map[string][]int {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string][]int, len(m))
	for k, v := range m {
		out[k] = append([]int(nil), v...)
	}
	return out
}

// anyModeChosenLocked reports whether `modes` names an option `ab`
// has already used — the answer gate's re-check against the LIVE
// memory, so an answer can never take a mode another instance took
// first. Caller must hold g.mu.
func (g *Game) anyModeChosenLocked(ms *ModeSpec, ab ModeAbility, modes []int) bool {
	used := g.modesChosenLocked(ms, ab)
	for _, m := range modes {
		if slices.Contains(used, m) {
			return true
		}
	}
	return false
}
