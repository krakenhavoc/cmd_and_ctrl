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
// # Declared limitation: THE P/T GOES STALE BETWEEN RECOMPUTES
//
// The CDA is correct; its INVALIDATION is not. The layer engine
// caches its resolution and bumps the version on battlefield zone
// moves, token creation, counters and tap state — see
// layerVersionBump.OnEvent in internal/game/layer_listener.go. A
// hand-size change is not on that list, and two tests in
// internal/game (TestLayerVersionDoesNotBumpOnZoneMoveOutsideBattlefield,
// TestLayerVersionDoesNotBumpOnIrrelevantEvent) deliberately guard
// against adding it, because a bump on every hand move makes the
// recompute much hotter.
//
// So the Crawler is the right size at every recompute and can read
// one draw behind between them. In practice almost anything that
// happens on a board — a token, a counter, a permanent entering or
// leaving, a tap — refreshes it, and the Crawler's OWN drain trigger
// resolving does not.
//
// Fixing it properly means making the invalidation care about WHICH
// statics are live (bump on a hand change only while a hand-size CDA
// is on the battlefield) rather than widening the event list. That is
// a layer-engine change, not a card change, and it is filed on the
// sprint issue rather than smuggled in here.
func init() {
	Register(Spec{
		OracleID: "2876e74f-a242-4995-9702-0b737a1ab67a",
		Name:     "Psychosis Crawler",
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
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
