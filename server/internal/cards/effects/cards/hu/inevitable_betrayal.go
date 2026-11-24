package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inevitable Betrayal — Sorcery (Secret Lair Drop):
//
//	"Suspend 3—{1}{U}{U}
//	 Search target opponent's library for a creature card and put that card
//	 onto the battlefield under your control. Then that player shuffles."
//
// Bribery's search (SearchLibrary.LibraryOwner, the caster choosing and
// controlling) with the shared Suspend special action. Suspending needs no
// mana cost on the card; the card has none ({0} is unpayable here only
// because it can't be cast from hand: it has no mana cost, CR 202.1a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa5e9269-b1db-4bf0-a548-24f999f60e44",
		Name:         "Inevitable Betrayal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		SpecialActions: []game.SpecialAction{
			effects.Suspend(3, "{1}{U}{U}"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: item.Targets[0].ID,
				Predicate:    func(c game.Card) bool { return c.IsCreature() },
				Dest:         game.ZoneBattlefield,
				Limit:        1,
				Reveal:       true,
				Shuffle:      true,
				Reason:       "Inevitable Betrayal — search target opponent's library for a creature card",
			}.Apply(ctx)
		},
	})
}
