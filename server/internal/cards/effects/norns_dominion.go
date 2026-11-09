package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Norn's Dominion — Plane — New Phyrexia (Planechase Anthology, #4870):
//
//	"When you planeswalk away from Norn's Dominion, destroy each nonland
//	 permanent without a fate counter on it, then remove all fate counters
//	 from all permanents.
//	 Whenever chaos ensues, you may put a fate counter on target
//	 permanent."
//
// A plane (ADR 0216 §6). A fate counter is a plain named counter (CR 122.1),
// so the ones other cards put on permanents spare them too (the ruling). The
// "may" is asked as the chaos ability resolves, after its target is chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d58f647-3f83-4f65-a944-c6236db7c422",
		Name:         "Norn's Dominion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkAwayFromThis("Norn's Dominion — destroy each nonland permanent without a fate counter, then remove all fate counters",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := b9DestroyEachNonlandPermanent(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.Counters[nornsDominionFate] > 0
					}); err != nil {
						return err
					}
					for _, c := range g.BattlefieldCardsForEffect() {
						if n := c.Counters[nornsDominionFate]; n > 0 {
							if err := g.AddCounterForEffect(c.InstanceID, nornsDominionFate, -n); err != nil {
								return err
							}
						}
					}
					return nil
				}),
			Targeting(WheneverChaosEnsues("Norn's Dominion — you may put a fate counter on target permanent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return MayChoice{
						Question: "Norn's Dominion — put a fate counter on the target permanent?",
						Then:     nornsDominionCounter,
						Carry:    []uuid.UUID{t},
					}.Apply(ctx)
				}),
				TargetPermanent("target permanent")),
		},
	})
}

const nornsDominionFate = "fate"

// nornsDominionCounter is the keyed answer. An on-disk key.
var nornsDominionCounter = MayYes("confirm/norns-dominion-fate-counter", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return AddCounter{Target: r.Carry[0], Kind: nornsDominionFate, N: 1}.Apply(ctx)
})
