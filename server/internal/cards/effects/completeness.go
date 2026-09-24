package effects

import (
	"strconv"
	"strings"
)

// completeness.go — the machine-readable half of the catalog's
// long-standing "declared simplification" convention (AGENTS.md §7).
//
// # Why this exists
//
// Since S14 a card file has carried its simplifications in prose:
// "Sandbox simplification: X is not checked", "DECLARED SIMPLIFICATION
// — NO CYCLING", "No simplification." That convention is good for the
// reader of the file and useless to any program, for three reasons
// that #338 and #350 both ran into:
//
//   - The wording is free-form. "No simplification." / "No
//     simplifications." / "No other simplification." / "No
//     simplification remains." all appear, as do half a dozen ways of
//     announcing a gap.
//   - Silence is ambiguous. A file with no note at all might be a
//     complete implementation whose author saw nothing worth saying,
//     or a card nobody ever checked. Mind Stone and Blaze are the
//     former; there is no way to tell them from the latter by reading.
//   - Notes go stale in BOTH directions. #338 was a sweep of notes
//     that still announced gaps the engine had since closed; a note
//     can equally fail to appear when a gap opens.
//
// A page that publishes "these cards work" cannot be built on a
// signal with those properties — it would promise things the engine
// does not do. So the declaration becomes a field, the field's zero
// value means "nobody has checked", and the prose stays where it is
// useful: explaining WHY, at length, to the next person editing the
// file.
//
// # The contract
//
// Completeness is about THIS CARD versus ITS PRINTED TEXT. It is not
// a quality score and not a statement about the engine at large. A
// card is Full when everything printed on it happens; it takes
// Caveats when some printed clause does not, whatever the reason.
//
// Register enforces the only two things it can check mechanically:
// a Caveats entry must say what the caveat IS, and a card that has
// nothing to declare must not smuggle caveat text in anyway. It
// deliberately does NOT reject CompletenessUnreviewed — see below.
//
// # Why Unreviewed is legal
//
// Making a missing declaration a boot panic would be the stricter
// design, and it is the wrong one here. Cards arrive in batches of
// thirty from several branches at once; a hard gate would turn every
// one of those into a merge-blocking chore and the predictable
// outcome is a reviewer stamping CompletenessFull to make the build
// go green. An unreviewed card that SAYS it is unreviewed is honest.
// A card falsely marked complete is the failure this whole file
// exists to prevent, so the design makes the lazy path (declare
// nothing) safe and visible rather than fast and wrong.
type Completeness uint8

const (
	// CompletenessUnreviewed is the zero value, and it means exactly
	// what it says: nobody has audited this spec against the printed
	// card. It is NOT a synonym for "working" — an unreviewed card
	// may be flawless or may be missing half its text. Anything that
	// publishes the catalog must show it as its own third state and
	// must never fold it in with CompletenessFull.
	//
	// The zero value carries this meaning on purpose. A spec that
	// forgets the field, a card file copy-pasted from another, and a
	// card added on a branch that predates this field all land here,
	// which is the conservative answer in every one of those cases.
	CompletenessUnreviewed Completeness = iota

	// CompletenessFull declares that every clause printed on the card
	// is implemented: the effect, its targeting restrictions, its
	// triggers, its costs and any rider. Someone compared the spec to
	// the oracle text and found nothing missing.
	//
	// Vanilla permanents qualify — a creature whose entire printed
	// text is its P/T is fully implemented by a spec that declares no
	// behaviour at all.
	CompletenessFull

	// CompletenessCaveats declares that the card works, but some
	// printed clause is not modelled. Spec.Caveats says which, and
	// Register refuses the declaration without it.
	//
	// This is the honest home for the whole S14 "sandbox
	// simplification" family: a clause deferred to a later sprint, a
	// restriction the target system cannot express yet, a trigger on
	// an event the engine does not emit. The card is still worth
	// playing; the player just deserves to know the gap before they
	// find it mid-game.
	CompletenessCaveats
)

// String renders the constant as the token used on the wire and in
// test failures. Kept stable: the catalog HTTP response and the
// client's filter both key off these exact strings.
func (c Completeness) String() string {
	switch c {
	case CompletenessFull:
		return "full"
	case CompletenessCaveats:
		return "caveats"
	default:
		return "unreviewed"
	}
}

// playerFacingJargon is the engine vocabulary a player-facing sentence
// must not contain. The list is the tone rule's whole content, so it
// lives here, beside the field it polices, rather than in a test.
var playerFacingJargon = []string{
	"OnResolve", "OnETB", "AsEnters", "TargetSpec", "InstanceID", "StackItem",
	"ctx.", "*Game", "Locked", "sub-PR",
}

// PlayerFacingProblems reports why text is not a player-facing
// sentence, or nil when it is. It is the tone rule every published
// Caveat is held to (AGENTS.md §7: "Write Caveats for a player, not
// for the next engineer"): long enough to say something, a sentence,
// and free of the engine vocabulary a card file's doc comment is for.
//
// Exported because a Caveat is not the only sentence a player reads
// about the engine: the public roadmap's summaries (internal/roadmap)
// are held to the same rule by the same function, so the two cannot
// drift apart.
func PlayerFacingProblems(text string) []string {
	var out []string
	if len(text) < 20 {
		out = append(out, "too terse to be useful")
	}
	if !strings.HasSuffix(text, ".") {
		out = append(out, "not a sentence")
	}
	for _, j := range playerFacingJargon {
		if strings.Contains(text, j) {
			out = append(out, "leaks engine jargon "+strconv.Quote(j))
		}
	}
	return out
}
