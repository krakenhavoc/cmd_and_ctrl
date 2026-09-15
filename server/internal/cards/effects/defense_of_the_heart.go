package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defense of the Heart — Enchantment {3}{G} (EDHREC rank 1299):
//
//	"At the beginning of your upkeep, if an opponent controls three or
//	 more creatures, sacrifice this enchantment, search your library
//	 for up to two creature cards, put those cards onto the
//	 battlefield, then shuffle."
//
// Green's Tooth and Nail for four. An upkeep trigger with an
// intervening-if, checked at trigger time and again at resolution
// (CR 603.4): an opponent who lost a creature in response leaves the
// enchantment on the battlefield and the library alone. When it does
// resolve, the sacrifice is made if the enchantment is still there
// and the search happens EITHER WAY — per the printed ruling, an
// enchantment removed in response cannot be sacrificed but the
// creatures still come. "Up to two" is the chooser's pick through the
// S22 prompt, unrevealed, and both enter untapped, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7e1b166-9267-426d-897d-24903327b48d",
		Name:         "Defense of the Heart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b11OpponentControlsCreatures(g, source.Controller, 3)
			}, "Defense of the Heart — sacrifice it, put up to two creature cards onto the battlefield", func(g *game.Game, item *game.StackItem) error {
				if !b11OpponentControlsCreatures(g, item.Controller, 3) {
					return nil
				}
				ctx := NewContext(g, item)
				if onBattlefield(g, item.SourceCardID) {
					if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
						return err
					}
				}
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() },
					Dest:      game.ZoneBattlefield,
					Limit:     2,
					Shuffle:   true,
					Reason:    "Defense of the Heart — up to two creature cards, onto the battlefield",
				}.Apply(ctx)
			}),
		},
	})
}
