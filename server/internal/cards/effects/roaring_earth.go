package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roaring Earth — Enchantment {1}{G} (Neon Dynasty, #5493):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 target creature or Vehicle you control.
//	 Channel — {X}{G}{G}, Discard this card: Put X +1/+1 counters on target
//	 land you control. It becomes a 0/0 green Spirit creature with haste.
//	 It's still a land."
//
// The landfall is Ride the Shoopuf's shape over the shared +1/+1 counter
// body, with a creature-or-Vehicle target. Channel is Awaken's
// counters-then-animate (CR 702.113a) with the creature type and a colour
// the printed text names: X is announced in the hand-zone cost (CR 602.2b),
// the counters go on through the counter pipeline (so Hardened Scales and
// Doubling Season apply), and the land becomes a green Spirit creature with
// haste for good (CR 611.2a), still a land (CR 305.7), only if it is still
// the object that was targeted (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43abfce5-b1dd-4784-bb91-d57b2c28f6a6",
		Name:         "Roaring Earth",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			Targeting(Landfall("Roaring Earth — put a +1/+1 counter on target creature or Vehicle you control", PlusOneCounterOnTheTarget),
				TargetPermanent("target creature or Vehicle you control", Or(Creature(), HasSubtype("Vehicle")), YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {X}{G}{G}, Discard this card: Put X +1/+1 counters on target land you control. It becomes a 0/0 green Spirit creature with haste. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    Plus(ManaCost("{X}{G}{G}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetPermanent("target land you control", Land(), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				land, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				tail := game.KeyedTail{
					Ref: neoB07AnimateLand, Source: ctx.Source(), Player: ctx.Controller(),
					Carry: []uuid.UUID{id}, N: int(land.EnteredBattlefieldAt),
				}
				if x := ctx.X(); x > 0 {
					return g.AddCounterByKeyedForEffect(ctx.Controller(), id, game.CounterPlusOne, x, tail)
				}
				return tail.Run(g, nil)
			},
		}},
	})
}
