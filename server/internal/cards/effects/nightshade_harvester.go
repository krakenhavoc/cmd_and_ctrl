package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightshade Harvester — Creature — Elf Shaman {3}{B}, 2/2 (EDHREC
// rank 4506):
//
//	"Whenever a land an opponent controls enters, that player loses 1
//	 life. Put a +1/+1 counter on this creature."
//
// A four-mana body that punishes the thing every Commander deck does
// every turn. In a four-player game it triggers three times a turn
// cycle off ordinary land drops alone, and once per fetch, once per
// Cultivate, once per Rampant Growth — the Harvester grows itself out
// of removal range while draining the table.
//
// "That player" is the land's controller, read off the entering
// permanent rather than chosen, so there is no target and nothing to
// redirect: a Simic player ramping drains themselves, not whoever the
// Harvester's controller would prefer. Land TOKENS and lands put onto
// the battlefield by a search count — the clause is "enters", not "is
// played" (CR 401.4 is about playing a land; this is a zone change).
//
// The counter is unconditional in the printed text, but AddCounter on
// a permanent that has left the battlefield does nothing, so a
// Harvester killed in response drains and gets no counter. The drain
// is life LOSS, not damage, so no prevention shield stops it.
//
// Batching gap: the engine emits one EventETB per land, so
// two lands entering simultaneously under one opponent trigger twice.
// The printed card says "a land", singular, so twice is correct here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "875c3db1-2752-48e7-9664-bbb12120c032",
		Name:         "Nightshade Harvester",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b43ALandAnOpponentControlsEntered(ev, source, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				victim, _ := b43ALandAnOpponentControlsEntered(ev, source, g)
				return game.NewTriggeredItem(source, "Nightshade Harvester — that player loses 1 life, a +1/+1 counter on this creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if g.PlayerByIDForEffect(victim) != nil {
							if err := g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1); err != nil {
								return err
							}
						}
						if !b09SourceStillOnBattlefield(g, item) {
							return nil
						}
						return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					})
			},
		}},
	})
}
