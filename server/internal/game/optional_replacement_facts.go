package game

// optional_replacement_facts.go — what an optional_replacement prompt
// is about, read off the paused event it carries (#2390).
//
// The prompt is one yes/no for every "may" replacement, so its kind
// does not say what the yes does. A policy that must weigh the answer
// rather than take it needs the one fact the effect knows: how many
// cards a dredge mills, and where a commander goes if its owner
// declines CR 903.9b. Both are computed from the frame on every view,
// never stored. A restore point carries no frame, and so no prompt to
// describe (snapshot.go's continuation census).

// DredgeOffer is N when this prompt offers a dredge — an
// optional_replacement whose effect declares ReplacementEffect.Dredge
// (CR 702.52a) — and zero on every other prompt.
func (c *PendingChoice) DredgeOffer() int {
	if c == nil || c.Kind != PendingChoiceOptionalReplacement {
		return 0
	}
	f := c.replacementResume
	if f == nil || len(f.applicable) != 1 {
		return 0
	}
	return f.applicable[0].effect.Dredge
}

// CommanderHeadedFor is where the commander goes if its owner declines
// CR 903.9b's "may": ZoneHand or ZoneLibrary, read off the paused move.
// Empty on every other prompt, and on #1397's question asked before a
// cost is paid, whose frame parks an announcement rather than a move
// (cost_commander_choice.go).
func (c *PendingChoice) CommanderHeadedFor() ZoneKind {
	if c == nil || c.Kind != PendingChoiceOptionalReplacement {
		return ""
	}
	f := c.replacementResume
	if f == nil || f.ev == nil || len(f.applicable) != 1 || !f.applicable[0].effect.commanderZone {
		return ""
	}
	return f.ev.NewZone
}
