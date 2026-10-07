package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// They Came from the Pipes — Enchantment {4}{U}:
//
//	"When this enchantment enters, manifest dread twice.
//	 Whenever a face-down creature you control enters, draw a card."
//
// "Twice" is two separate manifest dreads, each with its own look at
// the top two cards and its own answer. The enchantment is already on
// the battlefield when the first one resolves, so both face-down
// creatures draw a card. The draw trigger also fires for any other way
// a face-down creature enters under your control (a morph cast, a
// cloak).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b8870fd-2d8a-42af-8173-b747d8d4c04f",
		Name:         "They Came from the Pipes",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "They Came from the Pipes — manifest dread twice", Do(ManifestDreadTimes{N: 2})),
			On(game.EventETB, faceDownCreatureEnteredUnderYourControl,
				"They Came from the Pipes — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// faceDownCreatureEnteredUnderYourControl matches "a face-down creature
// you control enters". A face-down permanent is a 2/2 creature (CR 708.2)
// whatever the card under it is.
func faceDownCreatureEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.FaceDown && c.IsCreature()
}
