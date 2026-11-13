package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodland Liege — Creature — Elf Druid Noble {2}{G}, 2/2 (Foundations
// Jumpstart, #5464):
//
//	"Whenever a Beast you control enters, draw a card."
//
// Aether Charge's trigger condition with a draw for its payoff: the
// entering permanent must be under the Liege's controller's control and
// have the Beast subtype as it enters. "Beast" is a word of the card's
// text (CR 612.2), so a changeling counts. An opponent's Beast does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df63e124-1542-48d6-b255-cf45855f1e93",
		Name:         "Woodland Liege",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype(source.TextSubtype("Beast"))
			}, "Woodland Liege — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
