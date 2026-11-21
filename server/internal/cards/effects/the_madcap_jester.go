package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Madcap Jester — Legendary Creature — Clown Wizard {1}{U}{B}{R}, 3/2
// (Mystery Booster Commander Edition):
//
//	"Whenever The Madcap Jester attacks, each opponent exiles the top card
//	 of their library.
//	 Whenever a spell or ability you control causes one or more cards to be
//	 exiled from an opponent's library, you may play those cards until the
//	 end of your next turn. Mana of any type can be spent to cast spells
//	 this way."
//
// The second ability reads the cause every routed zone change records
// (Ranar the Ever-Watchful's reading, game.ExiledBySpellOrAbilityOf), so an
// exile a cost made, an opponent's own effect or a sandbox move does not
// count, and neither does a card taken from a graveyard or the battlefield.
// It is "one or more": one trigger per batch of events (CR 603.2c), and the
// permission covers every card of that batch an opponent owned. The cards
// are locked as the ability resolves (CR 611.2c); the window ends at the
// cleanup step of the controller's next turn (CR 514.2), the permission
// takes lands as well as spells, and its mana may be of any type
// (CastPermission.AnyType, so {C} costs are payable with any mana).
//
// No simplification.
func init() {
	const lib = "The Madcap Jester — each opponent exiles the top card of their library"
	Register(Spec{
		OracleID:     "60f3857d-2dcd-427e-aa0f-e2028bbc5acc",
		Name:         "The Madcap Jester",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(lib, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, opp := range ctx.Opponents() {
					if p := g.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
						continue
					}
					if err := g.ExileTopKeyedForEffect(opp, ctx.Controller(), 1, game.KeyedTail{}); err != nil {
						return err
					}
				}
				return nil
			}),
			Optional(OncePerBatch(On(game.EventZoneMove, madcapJesterSawAnExile,
				"The Madcap Jester — you may play the cards exiled from an opponent's library until the end of your next turn",
				madcapJesterGrant)),
				"The Madcap Jester — play the cards exiled from an opponent's library until the end of your next turn?"),
		},
	})
}

// madcapJesterExiledFromAnOpponentsLibrary reports whether ev is a card
// leaving an opponent's library for exile because a spell or ability `you`
// control did it.
func madcapJesterExiledFromAnOpponentsLibrary(ev game.Event, you uuid.UUID, g *game.Game) bool {
	if ev.Kind != game.EventZoneMove || ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneExile ||
		!game.ExiledBySpellOrAbilityOf(ev, you) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner != uuid.Nil && c.Owner != you
}

func madcapJesterSawAnExile(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return madcapJesterExiledFromAnOpponentsLibrary(ev, source.Controller, g)
}

// madcapJesterGrant gives the controller one permission over every card
// of the triggering batch that qualifies and is still in exile.
func madcapJesterGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ev := ctx.Trigger().Event
	var cards []game.Card
	seen := map[uuid.UUID]bool{}
	for i := len(g.Events) - 1; i >= 0; i-- {
		e := g.Events[i]
		if e.Batch < ev.Batch {
			break
		}
		if e.Batch != ev.Batch || seen[e.CardID] || !madcapJesterExiledFromAnOpponentsLibrary(e, ctx.Controller(), g) {
			continue
		}
		seen[e.CardID] = true
		if z := g.FindCardZoneForEffect(e.CardID); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if c, ok := g.LookupCardForEffect(e.CardID); ok {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:   ctx.Controller(),
		Zone:     game.ZoneExile,
		Source:   ctx.Source(),
		AnyColor: true,
		AnyType:  true,
		Duration: g.UntilEndOfYourNextTurnDuration(ctx.Controller()),
	}, cards)
	return nil
}
