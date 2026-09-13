package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Encroaching Dragonstorm — Enchantment {3}{G} (EDHREC rank 2463):
//
//	"When this enchantment enters, search your library for up to two
//	 basic land cards, put them onto the battlefield tapped, then
//	 shuffle.
//	 When a Dragon you control enters, return this enchantment to its
//	 owner's hand."
//
// Explosive Vegetation that comes back: the ramp is the same one
// search with Limit 2 (the S22 chooser shows the basics and takes up
// to two, tapped, then shuffles), and the bounce is a second trigger
// on any Dragon entering under the controller's control — effective
// subtypes, so a changeling Dragon counts, and a Dragon token too.
// The enchantment returns only if it is still on the battlefield
// when the trigger resolves, so it can be recast for another two
// lands, which is the whole loop.
//
// No simplification. The known engine gap on a fetched permanent
// whose own entry queues a prompt (#478) is the search path's, not
// the card's, and a basic land never asks anything on entry.
func init() {
	Register(Spec{
		OracleID:     "1e95c273-7fec-4f8d-8be9-e21ad4b93717",
		Name:         "Encroaching Dragonstorm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Encroaching Dragonstorm — up to two basic lands, tapped",
						func(g *game.Game, item *game.StackItem) error {
							return SearchLibrary{
								Player:        item.Controller,
								Predicate:     IsBasicLand,
								Dest:          game.ZoneBattlefield,
								Limit:         2,
								Reveal:        true,
								Shuffle:       true,
								TappedOnEntry: true,
								Reason:        "Encroaching Dragonstorm — up to two basic lands",
							}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b23DragonYouControlEntered(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Encroaching Dragonstorm — return it to its owner's hand",
						func(g *game.Game, item *game.StackItem) error {
							if !onBattlefield(g, item.SourceCardID) {
								return nil
							}
							return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
