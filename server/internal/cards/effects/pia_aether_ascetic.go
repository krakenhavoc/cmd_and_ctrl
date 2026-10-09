package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pia, Aether Ascetic — Legendary Creature — Human Druid {2}{G}, 2/2:
//
//	"When Pia enters, you may discard a card. If you do, search your
//	 library for an enchantment card, reveal it, put it into your hand,
//	 then shuffle."
//
// The discard is asked as the trigger resolves and is the controller's
// choice; the search only happens when a card was actually discarded.
// The library is shuffled whether or not an enchantment was found.
//
// No simplification.
func init() {
	const label = "Pia, Aether Ascetic — you may discard a card; if you do, search for an enchantment"
	Register(Spec{
		OracleID:     "6519dc16-4960-4220-aaa8-7e4279743714",
		Name:         "Pia, Aether Ascetic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := ctx.Controller()
				return g.PlayerDiscardsThenForEffect(game.DiscardPrompt{
					Player:   player,
					Source:   ctx.Source(),
					N:        1,
					UpTo:     true,
					Question: label,
				}, func(g *game.Game, discarded game.PromptedDiscards) error {
					if discarded.Count() == 0 {
						return nil
					}
					return SearchLibrary{
						Player:    player,
						Predicate: game.Card.IsEnchantment,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Pia, Aether Ascetic — an enchantment card, to your hand",
						Source:    item.SourceCardID,
					}.Apply(NewContext(g, item))
				})
			}),
		},
	})
}
