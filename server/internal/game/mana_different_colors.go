package game

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// mana_different_colors.go — #2558: "Add two mana of different colors"
// (Firemind Vessel, Guild Globe, Interplanar Beacon, Component Pouch).
//
// The produced-mana grammar writes it as one slot, "{W|U|B|R|G:2}"
// (ProducedManaEntry.Distinct), because the constraint runs across the
// picks: two independent "{W|U|B|R|G}" slots would add {U}{U}, which is
// stronger than printed (#259). Every place that turns a slot into mana
// handles the shape, each in the way it already handles a pick:
//
//   - a player activating the ability names the colours up front
//     (#1443): the view publishes one colour list per mana and
//     ManaAbilityView.DifferentColors, and validateUpfrontManaColors
//     refuses a repeated colour before anything is paid;
//   - otherwise the activation queues ONE PendingChoiceMana carrying
//     ManaDifferent, answered one colour at a time. Each answer is
//     struck from the options the next one offers, and nothing is added
//     until the last colour is named: CR 605.3b resolves a mana ability
//     at once, so the N mana arrive together, and CR 106.12a's "tapped
//     for mana" triggers fire once, for the one resolution;
//   - the auto-tapper offers the solver one candidate per set of N
//     colours (appendTapSource), mutually exclusive like a Gilded
//     Lotus's per-colour candidates, and the plan carries the set to the
//     executor (plannedTap.DifferentColors);
//   - an effect that adds such mana with auto-tap requirements pending
//     picks greedily, never repeating a colour (pickDifferentColors).

// pickDifferentColors is pickColorForSlot for a different-colours slot:
// n DIFFERENT colours from `options`, each paying the most restrictive
// requirement still pending that a colour not yet taken can pay, then
// the first options not yet taken. The booked requirements are ticked
// off `pending`.
func pickDifferentColors(options []string, n int, pending *[]ColorRequirement) []string {
	if n <= 0 || len(options) == 0 {
		return nil
	}
	var picked []string
	left := append([]string(nil), options...)
	for len(picked) < n && len(left) > 0 {
		var color string
		if pending != nil {
			if best, opt := mostRestrictiveRequirement(left, *pending); best >= 0 {
				*pending = append((*pending)[:best], (*pending)[best+1:]...)
				color = opt
			}
		}
		if color == "" {
			color = left[0]
		}
		picked = append(picked, color)
		left = withoutColor(left, color)
	}
	return picked
}

// withoutColor is `options` with `color` struck out, as a new slice.
func withoutColor(options []string, color string) []string {
	out := make([]string, 0, len(options))
	for _, o := range options {
		if o != color {
			out = append(out, o)
		}
	}
	return out
}

// differentColorsAllowed reports whether `colors` is a legal answer to
// a different-colours slot offering `options`: exactly n colours, every
// one offered, none repeated.
func differentColorsAllowed(options []string, n int, colors []string) bool {
	if n <= 0 || len(colors) != n {
		return false
	}
	seen := make(map[string]bool, len(colors))
	for _, c := range colors {
		if seen[c] || !containsColor(options, c) {
			return false
		}
		seen[c] = true
	}
	return true
}

// differentColorSets is every set of n different colours from
// `options`, each in the options' order (identity-first, #843), the
// sets in lexicographic order of their positions. The auto-tapper's
// candidates for a different-colours slot: five colours taken two at a
// time are ten.
func differentColorSets(options []string, n int) [][]string {
	if n <= 0 || n > len(options) {
		return nil
	}
	var out [][]string
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	for {
		set := make([]string, n)
		for i, j := range idx {
			set[i] = options[j]
		}
		out = append(out, set)
		i := n - 1
		for i >= 0 && idx[i] == len(options)-n+i {
			i--
		}
		if i < 0 {
			return out
		}
		idx[i]++
		for j := i + 1; j < n; j++ {
			idx[j] = idx[j-1] + 1
		}
	}
}

// differentColorSlot is the index of the source's one different-colours
// slot, oneColorSlotNone when it has none, or oneColorSlotUnplannable
// when it has two (no printed card does). The auto-tapper's question.
func differentColorSlot(slots []ProducedManaEntry) int {
	found := oneColorSlotNone
	for i, slot := range slots {
		if !slot.DifferentColors() {
			continue
		}
		if found != oneColorSlotNone {
			return oneColorSlotUnplannable
		}
		found = i
	}
	return found
}

// expandDifferentColors replaces the one different-colours slot of
// `slots` with one fixed slot per colour of `colors`: the shape the
// planner booked (appendTapSource), rebuilt by the executor from the
// colours the plan carried. ok is false when the plan and the source no
// longer agree — the source has such a slot and `colors` is not a legal
// answer to it, or it has none and the plan named colours — so the
// executor drops the source before paying anything for it. A source
// with no such slot and a plan with no colours comes back unchanged.
func expandDifferentColors(slots []ProducedManaEntry, colors []string) ([]ProducedManaEntry, bool) {
	idx := differentColorSlot(slots)
	switch idx {
	case oneColorSlotUnplannable:
		return nil, false
	case oneColorSlotNone:
		return slots, len(colors) == 0
	}
	slot := slots[idx]
	if !differentColorsAllowed(slot.Options, slot.DistinctCount(slot.Options), colors) {
		return nil, false
	}
	out := make([]ProducedManaEntry, 0, len(slots)-1+len(colors))
	out = append(out, slots[:idx]...)
	for _, c := range colors {
		out = append(out, ProducedManaEntry{Options: []string{c}})
	}
	return append(out, slots[idx+1:]...), true
}

// differentColorsReason is the header of a different-colours pick:
// the ability's label, then which colour of how many is being named
// and which are already taken. "Add two mana of different colors —
// color 2 of 2 (not white)".
func differentColorsReason(label string, chosen []string, n int) string {
	var b strings.Builder
	b.WriteString(label)
	b.WriteString(" — color ")
	b.WriteString(strconv.Itoa(len(chosen) + 1))
	b.WriteString(" of ")
	b.WriteString(strconv.Itoa(n))
	if len(chosen) > 0 {
		names := make([]string, len(chosen))
		for i, c := range chosen {
			names[i] = ColorName(c)
		}
		b.WriteString(" (not ")
		b.WriteString(strings.Join(names, " or "))
		b.WriteString(")")
	}
	return b.String()
}

// queueDifferentColorsPickLocked queues the first pick of a
// different-colours slot. `proto` carries everything an ordinary pick
// of the same source carries (restrictions, riders, the source
// snapshot, ManaTapped); this fills in the colours, the count and the
// header. Caller must hold g.mu.
func (g *Game) queueDifferentColorsPickLocked(proto PendingChoice, options []string, n int, label string) {
	proto.Kind = PendingChoiceMana
	proto.Count = 1
	proto.ColorOptions = append([]string(nil), options...)
	proto.ManaAmounts = nil
	proto.ManaDifferent = n
	proto.ManaChosen = nil
	proto.ManaLabel = label
	proto.Reason = differentColorsReason(label, nil, n)
	g.QueueChoiceForEffect(proto)
}

// advanceDifferentColorsPickLocked records one answer to a
// different-colours pick (#2558). While colours remain to be named it
// rewrites the open choice in place — the answer struck from the
// options, added to ManaChosen, a fresh ID so every seat sees a new
// question — and reports done=false. On the last answer it reports the
// whole set and leaves the choice for the caller to dequeue.
//
// The answer has already been checked against ColorOptions, which never
// holds a colour already chosen, so a repeat cannot reach here; the
// check below is the belt to that brace. Caller must hold g.mu.
func (g *Game) advanceDifferentColorsPickLocked(choice *PendingChoice, color string) (colors []string, done bool, err error) {
	if containsColor(choice.ManaChosen, color) {
		return nil, false, ErrInvalidParam
	}
	chosen := append(append([]string(nil), choice.ManaChosen...), color)
	left := withoutColor(choice.ColorOptions, color)
	if len(chosen) >= choice.ManaDifferent || len(left) == 0 {
		return chosen, true, nil
	}
	choice.ManaChosen = chosen
	choice.ColorOptions = left
	choice.ID = uuid.New()
	choice.Reason = differentColorsReason(choice.ManaLabel, chosen, choice.ManaDifferent)
	g.notePlayerDecisionLocked()
	return nil, false, nil
}
