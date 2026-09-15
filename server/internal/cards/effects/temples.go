package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// temples.go — the ten-land scry cycle (Theros / Theros Beyond Death):
//
//	"This land enters tapped."
//	"When this land enters, scry 1."
//	"{T}: Add {X} or {Y}."
//
// A tapped dual that smooths the next draw. They go in almost every
// two-colour deck that isn't racing, which is why the whole cycle is
// here rather than a representative sample: the cards are data, and
// the machinery is identical across all ten.
//
// All three clauses are worth noting:
//
//   - "Enters tapped" is a real CR 614 self-replacement
//     (SelfEntersTapped), not a hook that taps after entry. The permanent is never
//     untapped on the battlefield, which is what the printed card says
//     and what a tap-watcher would otherwise misread.
//   - The scry is the land's own ETB trigger. It goes on the stack like
//     any other "When ~ enters" (#578 moved it off the direct AsEnters
//     hook), and queues its prompt when it resolves.
//   - The dual mana ability uses pipe syntax, so the controller picks
//     the colour at activation. Two separate one-colour abilities would
//     also work but would clutter the menu with a fixed pair.
//
// Grouped in one file deliberately. Ten near-identical cards in ten
// files is ten places to fix the same mistake; here the shape is stated
// once and each card is three lines of data.
func init() {
	for _, t := range []struct {
		oracleID string
		name     string
		a, b     string
	}{
		{"e6e6fce8-0f6a-4b84-865e-d4e4a4182f9f", "Temple of Silence", "W", "B"},
		{"dc55421f-dee8-4263-9df0-2365df5f14bb", "Temple of Malady", "B", "G"},
		{"79f94050-d850-41ca-b1db-5ae0cf743f0a", "Temple of Epiphany", "U", "R"},
		{"3baa8e38-ef93-435d-b63e-f781d5bfcc68", "Temple of Abandon", "R", "G"},
		{"7e26f0b7-20e6-46d5-8130-d98c14d6aa29", "Temple of Mystery", "G", "U"},
		{"89f43e27-790b-4ca1-8ba7-0882b31e0783", "Temple of Enlightenment", "W", "U"},
		{"7c439c18-31dc-41fe-b03d-3fca06e6fc0b", "Temple of Malice", "B", "R"},
		{"e521322b-0e83-458c-8936-7021a80ee279", "Temple of Plenty", "G", "W"},
		{"33b9b3bd-33ca-46f3-b8bb-a978bc3d1085", "Temple of Deceit", "U", "B"},
		{"6f0d94d9-64bb-4175-83bc-301e8f79f54f", "Temple of Triumph", "R", "W"},
	} {
		// Capture per iteration: the closures below outlive the loop
		// body, and a shared loop variable would give every Temple the
		// last entry's colours.
		produced := "{" + t.a + "|" + t.b + "}"
		label := "Add {" + t.a + "} or {" + t.b + "}"
		name := t.name
		Register(Spec{
			OracleID:     t.oracleID,
			Name:         t.name,
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			Triggered: []game.TriggeredAbility{
				WhenThisEnters(name+" — scry 1", Do(Scry{N: 1})),
			},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: produced,
				Label:    label,
			}},
		})
	}
}
