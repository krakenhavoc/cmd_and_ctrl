package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hydra Omnivore — 8/8 Creature — Hydra for {4}{G}{G} (EDHREC rank
// 4005):
//
//	"Whenever this creature deals combat damage to an opponent, it
//	 deals that much damage to each other opponent."
//
// A six-mana 8/8 that hits the whole table for eight when it connects
// with one player. It is a Commander-only card in every sense — in a
// duel the second clause does nothing — and it is in the batch
// because it is the only shape in the catalog where the amount dealt
// by the trigger is read off the triggering damage event rather than
// off the source's power.
//
// # "That much" is the damage that was dealt, not the Hydra's power
//
// The two differ constantly and the distinction is the card. A
// blocker that ate four of the eight, a Fog that prevented half, a
// damage doubler, a creature pumped after damage was already dealt —
// in every one of those the trigger deals what LANDED on the first
// opponent, which is what the event carries. The amount is captured
// in Build, at trigger time, because by resolution the Hydra's power
// may be anything.
//
// # "Each OTHER opponent"
//
// The player who took the combat damage is excluded — they already
// took it. Everyone else at the table who is an opponent of the
// HYDRA's controller takes the same amount. That is not combat
// damage: it is the trigger dealing damage from the Hydra, so it does
// not trigger a second time (no loop), it is not doubled by combat-
// damage replacements, and lifelink on the Hydra would gain life from
// it as it does from any damage.
//
// Trample changes nothing here: if the Hydra is blocked and tramples
// four through, four is what "that much" means.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f504855-f3df-4284-a189-e799bcddf620",
		Name:         "Hydra Omnivore",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b38ThisDealtCombatDamageToAnOpponent(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				// Captured by value at trigger time: how much landed,
				// and on whom. Both are gone by resolution.
				amount, hit := ev.Amount, ev.Target
				return game.NewTriggeredItem(source,
					"Hydra Omnivore — that much damage to each other opponent",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, opp := range ctx.Opponents() {
							if opp == hit {
								continue
							}
							if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: amount}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
