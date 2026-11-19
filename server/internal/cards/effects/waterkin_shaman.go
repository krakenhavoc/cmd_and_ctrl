package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waterkin Shaman — Creature — Elemental Shaman {1}{U}, 2/1 (Foundations
// Jumpstart, S82, #5929):
//
//	"Whenever a creature you control with flying enters, this creature
//	 gets +1/+1 until end of turn."
//
// Dragon Tempest's flier-entered trigger (judged as it entered, through
// the layered view, the source itself included) with Momo's pump on this
// creature.
//
// No simplification.
func init() {
	const label = "Waterkin Shaman — gets +1/+1 until end of turn"
	Register(Spec{
		OracleID:     "64e0c580-0e55-4f42-8594-11ccb584d633",
		Name:         "Waterkin Shaman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && game.HasKeyword(&c, "flying")
			}, label, thisGetsUntilEndOfTurn(1, 1, label)),
		},
	})
}
