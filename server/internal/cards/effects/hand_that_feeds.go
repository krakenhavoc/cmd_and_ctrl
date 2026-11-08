package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hand That Feeds — Creature — Mutant {1}{R}, 2/2 (Duskmourn: House of
// Horror, #4466):
//
//	"Delirium — Whenever this creature attacks while there are four or
//	 more card types among cards in your graveyard, it gets +2/+0 and
//	 gains menace until end of turn."
//
// Delirium is an ability word (CR 207.2c). "While" is part of the
// trigger condition, not an intervening "if": the graveyard is read as
// the creature is declared as an attacker (b16CardTypesInGraveyard,
// Demonic Counsel's count), and a graveyard emptied in response does not
// stop the boost. The boost and the menace go on the creature only if it
// is still the permanent that attacked (socB27ThisGetsAndGains).
//
// No simplification.
func init() {
	const label = "Hand That Feeds — +2/+0 and menace until end of turn"
	Register(Spec{
		OracleID:     "6b348051-3ba8-4aa0-b337-b42d8ef8cbba",
		Name:         "Hand That Feeds",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, AllOf(ThisAttacked, youHaveDeliriumWhen), label,
				func(g *game.Game, item *game.StackItem) error {
					return socB27ThisGetsAndGains(NewContext(g, item), 2, "menace", "Hand That Feeds")
				}),
		},
	})
}

// youHaveDeliriumWhen is a trigger condition's "while there are four or
// more card types among cards in your graveyard" (delirium, CR 207.2c),
// read for the source's controller as the event happens.
func youHaveDeliriumWhen(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return b16CardTypesInGraveyard(g, source.Controller) >= 4
}
