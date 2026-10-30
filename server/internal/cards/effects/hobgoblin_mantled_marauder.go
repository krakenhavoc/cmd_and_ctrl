package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hobgoblin, Mantled Marauder — Legendary Creature — Goblin Human Villain
// {1}{R}, 1/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying, haste
//	 Whenever you discard a card, Hobgoblin gets +2/+0 until end of turn."
//
// Glint-Horn Buccaneer's discardedByYou watch (one trigger per card
// discarded, CR 603.2c) with a +2/+0 until-end-of-turn boost on Hobgoblin
// if he is still the same permanent.
//
// No simplification.
func init() {
	const label = "Hobgoblin — +2/+0 until end of turn"
	Register(Spec{
		OracleID:        "1d453331-6b8e-4d3d-b658-683eab2760da",
		Name:            "Hobgoblin, Mantled Marauder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, label, func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: label}.Apply(NewContext(g, item))
			}),
		},
	})
}
