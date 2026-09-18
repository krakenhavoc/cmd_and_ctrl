package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Etched Oracle — Artifact Creature — Wizard {4}, 0/0 (EDHREC rank
// 6423):
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for
//	 each color of mana spent to cast it.)
//	 {1}, Remove four +1/+1 counters from this creature: Target
//	 player draws three cards."
//
// Both halves of this PR on one card: sunburst reads the mana that
// paid (#761), and the draw ability spends the counters it made
// (#625's self form, now on a card that can actually make four).
//
// Sunburst (CR 702.44a) is SunburstCounters, and it carries that
// helper's declared simplification: the counters go on as the spell
// RESOLVES, a beat before the Oracle enters, because the entry
// pipeline cannot read the paid-cost record. They are on it when it
// lands, so the 0/0 body never meets the state-based check without
// them and every ETB watcher sees the finished creature — the same
// road Hangarback Walker's X counters take.
//
// Spec.WantsDistinctColors makes the cast gate spread the payment
// across colours, so an Oracle cast off four different lands really
// is a 4/4. Cast off four colourless mana it is a 0/0 and dies at
// once, which is exactly what the card does in paper.
//
// The draw ability is instant speed and has no {T}, so a freshly
// landed Oracle with four counters can cash them the turn it arrives.
func init() {
	Register(Spec{
		OracleID:            "7ecfa47e-1165-46a6-884c-290b1c14d020",
		Name:                "Etched Oracle",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"The +1/+1 counters go on as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them. With strict mana off, the game doesn't track which mana you spent, and it enters with no counters at all."},
		WantsDistinctColors: true,
		OnResolve:           SunburstCounters(game.CounterPlusOne),
		Activated: []ActivatedAbility{{
			Label: "{1}, Remove four +1/+1 counters from this creature: Target player draws three cards.",
			Cost: Plus(
				ManaCost("{1}"),
				RemoveCountersFromThis(game.CounterPlusOne, 4),
			),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, ref := range ctx.LegalTargets() {
					if ref.Kind == game.TargetPlayer {
						return DrawCards{Player: ref.ID, N: 3}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
