package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stitched Assistant — Creature — Zombie {2}{U}, 3/2:
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, scry 1, then draw a card."
//
// Exploit is the shared constructor (exploit.go); the payoff is an
// untargeted reflexive trigger. The scry is a question, so the draw is
// its tail (Lembas's body): the card drawn is the one the scry left on
// top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eedbc164-032c-420f-ac59-bd2675ff1b69",
		Name:         "Stitched Assistant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			Exploit("Stitched Assistant", stitchedAssistantScryBody),
		},
	})
}

// stitchedAssistantScryBody is "When this creature exploits a creature,
// scry 1, then draw a card." An on-disk key: never renamed, never reused.
var stitchedAssistantScryBody = game.ReflexiveBody("stitched-assistant/scry-draw", simpleBody(lembasScryThenDraw), nil)
