package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Borderland Explorer — Creature — Elf Scout {1}{G}, 3/1 (Commander 2018, S82
// OLD batch 193, #5526):
//
//	"When this creature enters, each player may discard a card. Each player
//	 who discarded a card this way may search their library for a basic land
//	 card, reveal it, put it into their hand, then shuffle."
//
// Each player, the controller first and then the others in turn order, is
// asked "may discard a card" (an "up to one" discard prompt, so declining is
// an answer, and an empty hand asks nothing). A player's own answer carries
// their search: if they discarded, they may search for a basic land card
// (finding none is allowed, CR 701.23b), reveal it, put it into their hand
// and shuffle; a player who did not discard does not search. Every prompt is a
// restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65bc4f8a-3226-4648-862b-3f62736ff1ae",
		Name:         "Borderland Explorer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Borderland Explorer — each player may discard a card; each who did may search for a basic land card",
				func(g *game.Game, item *game.StackItem) error {
					for _, p := range effects.ApnapPlayers(g) {
						if err := g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
							Player:   p,
							Source:   item.SourceCardID,
							N:        1,
							UpTo:     true,
							Question: "Borderland Explorer — you may discard a card; if you do, you may search for a basic land card",
							Asked:    game.NewAskedItem(item),
						}, borderlandExplorerSearch, nil); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

// borderlandExplorerSearch is one player's search, once their discard is
// settled. An on-disk key.
var borderlandExplorerSearch = game.RegisterDiscardThen("choose-cards/old-b193-borderland-explorer-search", func(g *game.Game, r game.DiscardsDone) error {
	if !r.Discarded.Discarded(r.Player) || r.Item == nil {
		return nil
	}
	return effects.SearchLibrary{
		Player:    r.Player,
		Predicate: effects.IsBasicLand,
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Optional:  true,
		Shuffle:   true,
		Reason:    "Borderland Explorer — a basic land card, to your hand",
	}.Apply(effects.NewContext(g, r.Item))
})
