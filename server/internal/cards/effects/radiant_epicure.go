package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radiant Epicure — Creature — Vampire Wizard {4}{B}, 5/5 (Modern
// Horizons 2, #5501):
//
//	"Converge — When this creature enters, each opponent loses X life and
//	 you gain X life, where X is the number of colors of mana spent to
//	 cast this spell."
//
// Transcendent Archaic's converge enter trigger: X is read off the
// entering object's payment as the trigger is built, so an Epicure that
// was not cast has X = 0 and does nothing. You gain X, not the total the
// opponents lost.
//
// No simplification.
func init() {
	const label = "Radiant Epicure — each opponent loses X life and you gain X life"
	enters := WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
		x := item.Params.Amount
		if x <= 0 {
			return nil
		}
		if err := eachOpponentLosesLife(g, item, x); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: x}.Apply(NewContext(g, item))
	})
	enters.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = source.ManaSpentToCast().ColorCount()
		return item
	}
	Register(Spec{
		OracleID:            "cdea4344-24b6-4659-a830-c51d4b348af5",
		Name:                "Radiant Epicure",
		Completeness:        CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Triggered:           []game.TriggeredAbility{enters},
	})
}
