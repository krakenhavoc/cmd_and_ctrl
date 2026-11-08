package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valley Questcaller — Creature — Rabbit Warrior {1}{W}, 2/3
// (Bloomburrow, #4467):
//
//	"Whenever one or more other Rabbits, Bats, Birds, and/or Mice you
//	 control enter, scry 1.
//	 Other Rabbits, Bats, Birds, and Mice you control get +1/+1."
//
// The trigger is one per batch of entries (OncePerBatch): a permanent
// other than this one entering under your control with any of the four
// subtypes, read through the Questcaller's text (CR 612.2). The
// Questcaller entering beside one of them sets it off (the ruling). The
// anthem is the shared TribalAnthem over the four types, "other" and
// "you control".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81c509b2-3164-47e9-94b3-bdba0487d2f1",
		Name:         "Valley Questcaller",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Rabbit", "Bat", "Bird", "Mouse"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, valleyQuestcallerKinEntered,
				"Valley Questcaller — scry 1", Do(Scry{N: 1}))),
		},
	})
}

// valleyQuestcallerKinEntered is "another Rabbit, Bat, Bird, or Mouse you
// control entered".
func valleyQuestcallerKinEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && AnySubtype("Rabbit", "Bat", "Bird", "Mouse")(g, source.Controller, game.ThroughTextOf(source, c))
}
