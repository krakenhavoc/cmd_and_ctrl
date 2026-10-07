package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charred Foyer // Warped Space — Enchantment — Room (ADR 0103):
//
//	Charred Foyer {3}{R}: "At the beginning of your upkeep, exile the top
//	 card of your library. You may play it this turn."
//	Warped Space {4}{R}{R}: "Once each turn, you may pay {0} rather than
//	 pay the mana cost for a spell you cast from exile."
//
// Charred Foyer is the impulse-exile primitive on an upkeep trigger.
// Warped Space is NOT implemented: a standing "pay {0} rather than the
// mana cost, once each turn, for spells cast from exile" has no
// primitive (ADR 0103's table). The door is empty, which is weaker than
// printed, so the card says so.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "9f50ec9f-78e6-403f-8f41-84d603fee0eb",
		Name:         "Charred Foyer // Warped Space",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Warped Space's free casting from exile isn't implemented, so that half does nothing."},
		Left: Door{Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Charred Foyer — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
		}},
	}))
}
