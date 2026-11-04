package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dimensional Exile — {1}{W} Enchantment — Aura (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Enchant basic land you control
//	 When this Aura enters, exile target creature an opponent controls
//	 until this Aura leaves the battlefield."
//
// Ossification with only creatures to exile. The exile is a one-shot and the
// return a second one-shot created as the Aura leaves (CR 610.3), under the
// card's owner's control (CR 610.3c); an Aura removed before its trigger
// resolves exiles nothing (CR 610.3b). The enchant clause is re-checked as a
// state-based action (CR 303.4c): a land that stops being a basic land you
// control sends the Aura to the graveyard, which returns the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f0777bc-4fca-4601-beed-1fd060f28075",
		Name:         "Dimensional Exile",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant basic land you control", b41BasicLandYouControl()),
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetCreature("target creature an opponent controls", OpponentControls()),
			Key:       "Dimensional Exile — exile target creature an opponent controls until this Aura leaves the battlefield",
			Effect: exileChosenTargetUntilThisLeaves(
				"Dimensional Exile — the exiled card returns when Dimensional Exile leaves the battlefield"),
		}},
	})
}
