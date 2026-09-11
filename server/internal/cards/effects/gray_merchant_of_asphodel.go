package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gray Merchant of Asphodel — Creature — Zombie, {3}{B}{B}, 2/4:
//
//	"When this creature enters, each opponent loses X life, where X
//	 is your devotion to black. You gain life equal to the life lost
//	 this way. (Each {B} in the mana costs of permanents you control
//	 counts toward your devotion to black.)"
//
// "Gary". At a four-player table a mono-black board of six devotion
// drains eighteen and gains eighteen, which is a third of everyone's
// starting life in one card, and is why this is the mono-black
// finisher.
//
// The batch-02 triage (#295) filed this under "cost modification".
// Nothing about this card modifies a cost — devotion (CR 702.x, the
// Theros keyword-action-adjacent count) just READS mana costs, which
// is ordinary arithmetic over ParsedCost. It is writable today and
// always was.
//
// # Devotion, precisely
//
// CR 700.5: devotion to black is the number of {B} symbols among the
// mana costs of permanents you control. Three details the loop
// below gets right and a naive version would not:
//
//   - It counts SYMBOLS, not permanents. A card costing {B}{B}
//     contributes two. ParsedCost.Required has one ColorRequirement
//     per coloured slot, in printed order, so len-style counting over
//     Required is the right shape.
//   - HYBRID COUNTS. A {B/R} symbol is a black symbol and contributes
//     to devotion to black (and to devotion to red). ColorRequirement
//     carries Options, so testing membership rather than equality is
//     what makes hybrid work.
//   - It reads the PRINTED mana cost. Devotion is not affected by
//     the layer engine — a permanent whose colour was changed keeps
//     the symbols printed on it, and a token with no mana cost
//     contributes nothing (ParseCost of "" is the zero ParsedCost,
//     which is exactly right and not an error).
//
// Lands contribute nothing for free: they have no mana cost, so they
// parse to zero required slots.
//
// Gray Merchant itself is on the battlefield when the trigger
// resolves, so its own {B}{B} counts — which is why the floor is two
// and why it is a five-drop rather than a four.
//
// # The drain
//
// "Each opponent loses X life" is LIFE LOSS, not damage — no
// prevention, no doubler, no lifelink sees it, and there is no source
// creature dealing it. eachOpponentLosesLife is the shared body.
//
// "You gain life equal to the life lost THIS WAY" is the total
// across all opponents, not X — three opponents at six devotion gain
// eighteen, not six. It is computed as X × (live opponents) rather
// than by summing actual life deltas, which differs only for a player
// whose life total cannot change; that case has no card in the
// catalog today.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "38f3b157-0df4-409b-89cc-086e1531cd5b",
		Name:         "Gray Merchant of Asphodel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Gray Merchant of Asphodel — drain for your devotion to black",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						x := devotionTo(g, item.Controller, "B")
						if x <= 0 {
							return nil
						}
						opponents := ctx.Opponents()
						if err := eachOpponentLosesLife(g, item, x); err != nil {
							return err
						}
						return GainLife{
							Player: item.Controller,
							Amount: x * len(opponents),
						}.Apply(ctx)
					})
			},
		}},
	})
}

// devotionTo counts CR 700.5 devotion: the number of `color` mana
// symbols among the mana costs of permanents `controller` controls.
// Hybrid symbols count toward every colour they offer, which is why
// this tests Options membership rather than equality. An unparseable
// or empty mana cost contributes nothing.
func devotionTo(g *game.Game, controller uuid.UUID, color string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		cost, err := game.ParseCost(c.ManaCost)
		if err != nil {
			continue
		}
		for _, req := range cost.Required {
			for _, opt := range req.Options {
				if opt == color {
					n++
					break
				}
			}
		}
	}
	return n
}
