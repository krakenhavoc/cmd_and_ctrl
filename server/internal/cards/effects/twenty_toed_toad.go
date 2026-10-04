package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twenty-Toed Toad — Creature — Frog Wizard {3}{U}, 3/3:
//
//	"Your maximum hand size is twenty.
//	 Whenever you attack with two or more creatures, put a +1/+1
//	 counter on this creature and draw a card.
//	 Whenever this creature attacks, you win the game if there are
//	 twenty or more counters on it or you have twenty or more cards in
//	 hand."
//
// The maximum is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11 timestamp order (the 2024-07-26 ruling: Toad
// then Spellbook is no maximum, the other order twenty).
//
// "Attack with two or more creatures" is Chivalric Alliance's trigger
// (b16PlayerAttackedWithAtLeast, once per declaration). The counter goes
// on the Toad only while it is the same object on the battlefield; the
// draw happens either way.
//
// The win trigger fires on every attack and checks both numbers only as
// it resolves (the 2024-07-26 ruling). "Twenty or more counters" counts
// counters of every kind on it, read from last-known information if the
// Toad has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cce643fb-cd88-457c-8f13-c203554df675",
		Name:         "Twenty-Toed Toad",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{YourMaxHandSizeIs(20)},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b16PlayerAttackedWithAtLeast(ev, source, g, 2, twentyToedToadAttackLabel)
			}, twentyToedToadAttackLabel, func(g *game.Game, item *game.StackItem) error {
				if err := putACounterOnThis(game.CounterPlusOne)(g, item); err != nil {
					return err
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			})),
			WheneverThisAttacks("Twenty-Toed Toad — you win the game with twenty counters or twenty cards in hand", func(g *game.Game, item *game.StackItem) error {
				// CR 608.2h: a Toad that has left is read as it last
				// existed on the battlefield.
				held := g.LastKnownCountersForEffect(item.SourceCardID)
				if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && onBattlefield(g, item.SourceCardID) {
					held = c.Counters
				}
				counters := 0
				for _, n := range held {
					counters += n
				}
				if counters < 20 && b14HandSize(g, item.Controller) < 20 {
					return nil
				}
				return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
			}),
		},
	})
}

const twentyToedToadAttackLabel = "Twenty-Toed Toad — you attacked with two or more creatures: a +1/+1 counter and a card"
