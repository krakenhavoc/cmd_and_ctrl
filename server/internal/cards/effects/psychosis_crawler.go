package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psychosis Crawler — Artifact Creature — Phyrexian Horror {5}:
//
//	"Psychosis Crawler's power and toughness are each equal to the
//	 number of cards in your hand."
//	"Whenever you draw a card, each opponent loses 1 life."
//
// The draw deck's win condition: a card that converts the thing the
// deck was already doing into damage, and gets bigger while doing it.
//
// Both halves are ordinary machinery in the right layers. The P/T is
// a CR 613 layer 7a characteristic-defining ability — Tarmogoyf's
// shape, with "cards in your hand" as the count, SET rather than
// modified so a later anthem or counter stacks on top of it in 7c and
// 7d. The drain watches EventDrawCard gated on the controller.
//
// "Each opponent loses 1 life" is loss, not damage: no prevention, no
// lifelink, no damage triggers.
//
// The trigger fires PER CARD, because EventDrawCard does — a spell
// that draws three fires it three times, which is exactly what makes
// the Crawler a win condition rather than a lightning bolt.
//
// # The P/T used to go stale between recomputes. It no longer does.
//
// The CDA was always correct; its INVALIDATION was not. The layer
// engine caches its resolution and bumped the version on battlefield
// zone moves, token creation, counters and tap state — and a hand
// change was on none of those lists, so the Crawler was the right
// size at every recompute and read one draw behind between them.
// Drawing a card is the one thing this deck does constantly, which
// made it the worst possible card to get wrong that way: its own
// drain trigger resolving did not refresh it.
//
// Widening the event list was the wrong fix and two tests in
// internal/game guard against it
// (TestLayerVersionDoesNotBumpOnZoneMoveOutsideBattlefield,
// TestLayerVersionDoesNotBumpOnIrrelevantEvent), because a hand
// change is the most frequent event in the game and bumping on it
// universally makes the recompute much hotter at every table,
// including the ones with no card that cares.
//
// So the invalidation asks a narrower question — "did a hand change
// WHILE SOMETHING WAS READING IT?" — and DependsOnHandSize below is
// how this card answers the second half. Both guard tests still pass
// unchanged: with no hand-size CDA on the battlefield the listener is
// the no-op they assert.
func init() {
	Register(Spec{
		OracleID:     "2876e74f-a242-4995-9702-0b737a1ab67a",
		Name:         "Psychosis Crawler",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			// The invalidation hint: a hand move drops the layer
			// cache while this permanent is on the battlefield, and
			// costs nothing when it is not.
			DependsOnHandSize: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
					n = p.Hand.Size()
				}
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Psychosis Crawler — each opponent loses 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, opp := range ctx.Opponents() {
							if err := g.ChangePlayerLifeForEffect(item.SourceCardID, opp, -1); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
