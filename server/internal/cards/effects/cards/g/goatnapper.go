package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goatnapper — Creature — Goblin Rogue {2}{R}, 2/2 (Lorwyn):
//
//	"When this creature enters, untap target Goat and gain control of it until
//	 end of turn. It gains haste until end of turn."
//
// Word of Seizing's body (untap, take until end of turn, haste, in printed
// order) on an enters trigger whose target is a Goat, any player's. A target
// that is gone or no longer a Goat on resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6875839a-9815-4ed2-adcf-b4d097c9bafd",
		Name:         "Goatnapper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Goatnapper — untap target Goat and gain control of it until end of turn; it gains haste until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TspB1UntapGainControlAndHaste("Goatnapper")(item, effects.NewContext(g, item))
					}),
				effects.TargetPermanent("target Goat", effects.HasSubtype("Goat"))),
		},
	})
}
