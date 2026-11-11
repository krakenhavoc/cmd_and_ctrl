package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// April O'Neil, Live on the Scene — {1}{U} Legendary Creature — Human
// Detective 2/1 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Whenever a Mutant, Ninja, or Turtle you control enters, investigate.
//	 (Create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")
//	 Partner—Character select"
//
// One trigger per permanent that enters under her controller's control
// with any of the three creature types (CR 205.3m), read as it enters,
// so a changeling counts and April herself counts only if something has
// made her one. Investigate is creating a Clue (CR 701.16a), the CR
// 111.10f token (ClueToken).
//
// Partner—Character select is a deck-building rule (CR 702.124i) that
// internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4900c157-8d9f-4f92-aaca-5246b6e2832e",
		Name:         "April O'Neil, Live on the Scene",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, aprilONeilMutantNinjaOrTurtleEntered,
				"April O'Neil, Live on the Scene — investigate",
				Do(Investigate{N: 1})),
		},
	})
}

// aprilONeilMutantNinjaOrTurtleEntered is "a Mutant, Ninja, or Turtle
// you control enters".
func aprilONeilMutantNinjaOrTurtleEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok {
		return false
	}
	for _, sub := range []string{"Mutant", "Ninja", "Turtle"} {
		if c.HasSubtype(source.TextSubtype(sub)) {
			return true
		}
	}
	return false
}
