package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tanglespan Lookout — Creature — Satyr {2}{G}, 2/3 (Wilds of Eldraine,
// #4471):
//
//	"Whenever an Aura you control enters, draw a card."
//
// Light-Paws, Emperor's Voice's Aura entering under your control, without
// its "if you cast it": an Aura put onto the battlefield, or a Role token
// created attached to a creature (a Role is an Aura), draws too. An Aura
// an opponent controls does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6a767c2-8bc8-4686-899a-0ba3a525977a",
		Name:         "Tanglespan Lookout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, tanglespanAuraYouControlEntered, "Tanglespan Lookout — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}

// tanglespanAuraYouControlEntered is "an Aura you control enters".
func tanglespanAuraYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsAura()
}
