package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sporocyst — Creature — Tyranid {X}{X}{G}, 0/0:
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Defender
//	 Spore Chimney — When this creature enters, search your library for up to
//	 X basic land cards, put them onto the battlefield tapped, then shuffle."
//
// X is the one announced for the spell (CR 107.3m), stamped on the trigger.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "ed58a617-cf86-4c2e-8044-1af37b1a1465",
		Name:                       "Sporocyst",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		PrintedKeywords:            []string{"defender"},
		EntersWithCountersFromCast: effects.RavenousEntry(),
		Triggered: []game.TriggeredAbility{
			effects.RavenousDraw("Sporocyst"),
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Key:       "Sporocyst — search for up to X basic land cards",
				XFromCast: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := ctx.X()
					if x <= 0 {
						return nil
					}
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         x,
						UpTo:          true,
						Shuffle:       true,
						TappedOnEntry: true,
						Source:        item.SourceCardID,
						Reason:        "Sporocyst — search for up to X basic land cards",
					}.Apply(ctx)
				},
			},
		},
	})
}
