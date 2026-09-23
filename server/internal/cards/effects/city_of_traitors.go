package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// City of Traitors — Land (EDHREC rank 1176):
//
//	"When you play another land, sacrifice this land.
//	 {T}: Add {C}{C}."
//
// Ancient Tomb without the pain and with a shelf life: two colourless
// until the next land drop. The mana is Sol Ring's; the drawback is a
// trigger on the controller's next land play, resolved by sacrificing
// the City (SacrificePermanent, so a "whenever you sacrifice" payoff
// sees it). A City already gone when the trigger resolves does
// nothing.
//
// Since #1326 the engine stamps CR 305.4's distinction on the settled
// entry itself (Event.Played), so the trigger reads that directly
// (b10AnotherLandPlayedByYou): a land played from hand, from an
// impulse-exile grant or from a graveyard permission counts, and a
// land an effect PUTS onto the battlefield — Cultivate, a fetchland,
// Rampant Growth, Titania's ETB, Splendid Reclamation, a flickered
// land — does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f161111d-9747-47b3-bb10-3c8bded32e21",
		Name:         "City of Traitors",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherLandPlayedByYou(ev, source, g)
			}, "City of Traitors — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
