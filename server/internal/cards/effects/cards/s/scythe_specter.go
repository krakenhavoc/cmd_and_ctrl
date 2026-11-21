package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scythe Specter — Creature — Specter {4}{B}{B}, 4/4 (Commander Starter
// Decks):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, each opponent
//	 discards a card. Each player who discarded a card with the greatest mana
//	 value among cards discarded this way loses life equal to that mana
//	 value."
//
// Every opponent of the Specter's controller chooses a card from their own
// hand (a discard run, Osseous Sticktwister's), and once every discard has
// landed the greatest mana value among the cards really discarded is read
// off the graveyards: each player who discarded a card of that value loses
// that much life, all of it as one step. A player with an empty hand discards
// nothing and so can't lose life; if nothing was discarded, or the greatest
// mana value is 0, nobody loses any.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "529ace49-c50e-4d3e-ac29-91e08af57fdc",
		Name:            "Scythe Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Scythe Specter — each opponent discards a card; the greatest mana value among them loses life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					opponents := ctx.Opponents()
					if len(opponents) == 0 {
						return nil
					}
					return g.PlayersDiscardThenKeyedForEffect(opponents, game.DiscardPrompt{
						Source:   item.SourceCardID,
						N:        1,
						Question: "Scythe Specter — discard a card",
					}, scytheSpecterDiscarded, []uuid.UUID{ctx.Controller()})
				}),
		},
	})
}

// scytheSpecterDiscarded is the discards done: each player who discarded a
// card with the greatest mana value among the cards discarded this way loses
// that much life. An on-disk key: never renamed, never reused.
var scytheSpecterDiscarded = game.RegisterPromptRunThen("prompt-run/scythe-specter-discarded", func(g *game.Game, r game.RunDone) error {
	greatest := 0
	var losers []uuid.UUID
	var values []int
	for _, leg := range r.Discards() {
		best := 0
		for _, id := range leg.Cards {
			if c, ok := g.LookupCardForEffect(id); ok && c.ManaValue() > best {
				best = c.ManaValue()
			}
		}
		if len(leg.Cards) == 0 {
			continue
		}
		losers = append(losers, leg.Seat)
		values = append(values, best)
		if best > greatest {
			greatest = best
		}
	}
	if greatest <= 0 {
		return nil
	}
	for i, seat := range losers {
		if values[i] != greatest {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(r.Source, seat, -greatest); err != nil {
			return err
		}
	}
	return nil
})
