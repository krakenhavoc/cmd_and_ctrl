package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slate of Ancestry — Artifact {4}:
//
//	"{4}, {T}, Discard your hand: Draw a card for each creature you
//	 control."
//
// A go-wide refill priced in cards. The hand is discarded as the
// activation is announced — the "Discard your hand" clause of #1600 —
// and the count is taken when the ability RESOLVES, over the creatures
// the activator controls then (CR 608.2h): a creature that dies in
// response is not counted, one that enters in response is. An empty
// hand pays the cost (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a07483b4-c04f-42a4-b979-8b77c11fa8f5",
		Name:         "Slate of Ancestry",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{4}, {T}, Discard your hand: Draw a card for each creature you control.",
			Cost:  Plus(ManaCost("{4}"), TapCost(), DiscardYourHand()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsCreature() {
						n++
					}
				}
				return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
			},
		}},
	})
}
