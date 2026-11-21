package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stadium Vendors — Creature — Goblin {3}{R}, 3/3 (Battlebond, #5529):
//
//	"When this creature enters, choose a player. That player adds two mana
//	 of any one color they choose."
//
// The controller names any player as the trigger resolves (ChoosePlayer,
// not a target), and the CHOSEN player adds the mana and picks its colour,
// as The Warring Triad's target does: the pick is queued for that seat and
// the two mana land in that seat's pool, emptying at the end of the step
// like any other mana (CR 500.5). Handing an opponent mana is a real
// Commander play, so the card is written the way it is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84904dd3-8987-47c2-b508-f87519157b00",
		Name:         "Stadium Vendors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stadium Vendors — choose a player; that player adds two mana of any one color",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChoosePlayer{
						Question: "Stadium Vendors — choose a player to add two mana of any one color",
						ThenKey:  stadiumVendorsChosen,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// stadiumVendorsChosen is "That player adds two mana of any one color they
// choose." An on-disk key: never renamed, never reused.
var stadiumVendorsChosen = effects.ChoosePlayerThen("option-pick/stadium-vendors-player", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	return effects.AddMana{Player: chosen, Produced: effects.OneColorOfAmount(2)}.Apply(ctx)
})
