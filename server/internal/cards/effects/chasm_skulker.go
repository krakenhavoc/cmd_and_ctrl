package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chasm Skulker — Creature — Squid Horror {2}{U} (1/1):
//
//	"Whenever you draw a card, put a +1/+1 counter on this creature.
//	 When this creature dies, create X 1/1 blue Squid creature tokens
//	 with islandwalk, where X is the number of +1/+1 counters on this
//	 creature. (They can't be blocked as long as defending player
//	 controls an Island.)"
//
// The draw trigger fires once per card drawn. X is read from the
// counters the Skulker had as it died (last-known information,
// CR 608.2h, via SourcePermanent), because the card is already in the
// graveyard, with no counters, by the time the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69facbc7-3859-4716-b627-5199571fb3cf",
		Name:         "Chasm Skulker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "Chasm Skulker — put a +1/+1 counter on this creature",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
			WhenThisDies("Chasm Skulker — create a 1/1 Squid with islandwalk for each +1/+1 counter it had",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok || info.Counters[game.CounterPlusOne] <= 0 {
						return nil
					}
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("1/1 blue Squid with islandwalk"),
						N:          info.Counters[game.CounterPlusOne],
					}.Apply(ctx)
				}),
		},
	})
}
